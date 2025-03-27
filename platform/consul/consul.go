// consul/client.go
package consul

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"log"
	"log/slog"
	"msgcenter/platform/consul/config"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

type WatchConfig struct {
	Services []string `yaml:"services"`
	Keys     []string `yaml:"keys"`
}

type Client struct {
	mu               sync.RWMutex
	client           *api.Client
	config           *api.Config
	services         map[string]*api.AgentServiceRegistration
	serviceEntries   map[string][]*api.ServiceEntry
	kvCache          map[string][]byte
	serviceWatchers  map[string]context.CancelFunc
	keyWatchers      map[string]context.CancelFunc
	watchCtx         context.Context
	watchCancel      context.CancelFunc
	activeWatchers   sync.WaitGroup
	callbacks        map[string][]func([]byte, []byte)
	keyVersions      map[string]uint64
	serviceLastIndex map[string]uint64
}

func NewClient(address string) (*Client, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("consul连接失败: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		client:           client,
		config:           config,
		services:         make(map[string]*api.AgentServiceRegistration),
		serviceEntries:   make(map[string][]*api.ServiceEntry),
		kvCache:          make(map[string][]byte),
		serviceWatchers:  make(map[string]context.CancelFunc),
		keyWatchers:      make(map[string]context.CancelFunc),
		watchCtx:         ctx,
		watchCancel:      cancel,
		callbacks:        make(map[string][]func([]byte, []byte)),
		keyVersions:      make(map[string]uint64),
		serviceLastIndex: make(map[string]uint64),
	}, nil
}

// 服务注册管理
// -------------------------------------------------------------------

func (c *Client) RegisterService(service *api.AgentServiceRegistration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.client.Agent().ServiceRegister(service); err != nil {
		return fmt.Errorf("服务注册失败: %w", err)
	}

	c.services[service.ID] = service

	if service.Check != nil && service.Check.TTL != "" {
		go c.maintainTTL(service.ID, service.Check.TTL)
	}

	return nil
}

func (c *Client) DeregisterService(serviceID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.client.Agent().ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("服务注销失败: %w", err)
	}

	delete(c.services, serviceID)
	return nil
}

func (c *Client) maintainTTL(serviceID, ttl string) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.RLock()
			_, exists := c.services[serviceID]
			c.mu.RUnlock()

			if !exists {
				return
			}

			if err := c.client.Agent().UpdateTTL(serviceID, "", api.HealthPassing); err != nil {
				log.Printf("TTL更新失败 %s: %v", serviceID, err)
			}
		case <-c.watchCtx.Done():
			return
		}
	}
}

// 动态监控管理
// -------------------------------------------------------------------

func (c *Client) StartDynamicWatch(cfg WatchConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 服务监控管理
	for _, service := range cfg.Services {
		if _, exists := c.serviceWatchers[service]; !exists {
			ctx, cancel := context.WithCancel(c.watchCtx)
			c.serviceWatchers[service] = cancel
			go c.watchService(ctx, service)
		}
	}

	// 移除旧监控
	for service, cancel := range c.serviceWatchers {
		if !contains(cfg.Services, service) {
			cancel()
			delete(c.serviceWatchers, service)
		}
	}

	// 配置监控管理
	for _, key := range cfg.Keys {
		if _, exists := c.keyWatchers[key]; !exists {
			ctx, cancel := context.WithCancel(c.watchCtx)
			c.keyWatchers[key] = cancel
			go c.watchKey(ctx, key)
		}
	}

	// 移除旧监控
	for key, cancel := range c.keyWatchers {
		if !contains(cfg.Keys, key) {
			cancel()
			delete(c.keyWatchers, key)
		}
	}
}

// 服务发现监控
func (c *Client) watchService(ctx context.Context, serviceName string) {
	c.activeWatchers.Add(1)
	defer c.activeWatchers.Done()

	var lastIndex uint64
	retryCount := 0

	for {
		select {
		case <-ctx.Done():
			return
		default:
			entries, meta, err := c.client.Health().Service(
				serviceName,
				"",
				true,
				&api.QueryOptions{
					WaitIndex: lastIndex,
					WaitTime:  10 * time.Second,
				},
			)

			if err != nil {
				log.Printf("服务监控错误[%s]: %v", serviceName, err)
				retryCount++
				time.Sleep(time.Duration(retryCount) * time.Second)
				continue
			}
			retryCount = 0

			if meta.LastIndex != lastIndex {
				c.updateServiceCache(serviceName, entries)
				lastIndex = meta.LastIndex
				c.serviceLastIndex[serviceName] = lastIndex
			}
		}
	}
}

func (c *Client) updateServiceCache(serviceName string, entries []*api.ServiceEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	oldEntries := c.serviceEntries[serviceName]
	c.serviceEntries[serviceName] = entries

	// 触发服务变更回调
	if len(oldEntries) != len(entries) {
		log.Printf("服务[%s]实例数变化: %d → %d",
			serviceName, len(oldEntries), len(entries))
	}
}

// 配置监控
func (c *Client) watchKey(ctx context.Context, key string) {
	c.activeWatchers.Add(1)
	defer c.activeWatchers.Done()

	var lastIndex uint64
	retryCount := 0

	for {
		select {
		case <-ctx.Done():
			return
		default:
			kv, meta, err := c.client.KV().Get(key, &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  10 * time.Second,
			})

			if err != nil {
				slog.Info("配置监控错误[%s]: %v", key, err)
				retryCount++
				time.Sleep(time.Duration(retryCount) * time.Second)
				continue
			}
			retryCount = 0

			if meta.LastIndex != lastIndex {
				c.updateKeyCache(key, kv)
				lastIndex = meta.LastIndex
			}
		}
	}
}

func (c *Client) updateKeyCache(key string, kv *api.KVPair) {
	c.mu.Lock()
	defer c.mu.Unlock()

	oldValue := c.kvCache[key]

	// 处理删除事件
	if kv == nil {
		if _, exists := c.kvCache[key]; exists {
			delete(c.kvCache, key)
			slog.Info("配置键[%s]已删除", key)
			c.triggerCallbacks(key, oldValue, nil)
		}
		return
	}

	// 处理更新事件
	if !bytes.Equal(oldValue, kv.Value) {
		c.kvCache[key] = kv.Value
		slog.Info("配置键[%s]已更新 (版本: %d)", key, kv.ModifyIndex)
		c.triggerCallbacks(key, oldValue, kv.Value)
	}
}

// 回调管理
// -------------------------------------------------------------------

func (c *Client) RegisterCallback(key string, callback func(old, new []byte)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.callbacks[key] = append(c.callbacks[key], callback)
}

func (c *Client) triggerCallbacks(key string, old, new []byte) {
	if callbacks, exists := c.callbacks[key]; exists {
		for _, cb := range callbacks {
			go cb(old, new)
		}
	}
}

// 连接管理
// -------------------------------------------------------------------

func (c *Client) RefreshConnection(newAddress string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 停止所有监控
	c.watchCancel()
	c.activeWatchers.Wait()

	// 创建新连接
	config := api.DefaultConfig()
	config.Address = newAddress
	client, err := api.NewClient(config)
	if err != nil {
		return fmt.Errorf("连接刷新失败: %w", err)
	}

	// 重新初始化
	ctx, cancel := context.WithCancel(context.Background())

	c.client = client
	c.config = config
	c.watchCtx = ctx
	c.watchCancel = cancel

	// 重新注册服务
	for _, service := range c.services {
		if err := client.Agent().ServiceRegister(service); err != nil {
			log.Printf("服务重新注册失败 %s: %v", service.ID, err)
		}
	}

	// 恢复监控
	go c.StartDynamicWatch(WatchConfig{
		Services: keys(c.serviceWatchers),
		Keys:     keys(c.keyWatchers),
	})

	return nil
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.watchCancel()
	c.activeWatchers.Wait()

	for serviceID := range c.services {
		if err := c.client.Agent().ServiceDeregister(serviceID); err != nil {
			log.Printf("服务注销失败 %s: %v", serviceID, err)
		}
	}
}

// 公共访问接口
// -------------------------------------------------------------------

func (c *Client) GetServiceInstances(serviceName string) []*api.ServiceEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serviceEntries[serviceName]
}

func (c *Client) GetConfigValue(key string) []byte {
	c.mu.RLock()
	value, exists := c.kvCache[key]
	c.mu.RUnlock()

	if !exists {
		// Attempt to sync from remote if not in cache
		kv, _, err := c.client.KV().Get(key, nil)
		if err != nil {
			slog.Error("远程获取配置失败", "key", key, "error", err)
			return nil
		}

		if kv != nil {
			c.mu.Lock()
			c.kvCache[key] = kv.Value
			c.keyVersions[key] = kv.ModifyIndex
			c.mu.Unlock()
			return kv.Value
		}
		return nil
	}
	return value
}

// 工具函数
// -------------------------------------------------------------------

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (c *Client) GetSqldb() config.SqlDb {
	var sqldb config.SqlDb
	err := sonic.Unmarshal(c.GetConfigValue("datacenter/sqldb"), &sqldb)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}
	slog.Info("获取配置成功", slog.Any("config", sqldb))
	return sqldb
}
