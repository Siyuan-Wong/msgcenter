// config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

var (
	GlobalConfig    *Config
	configMutex     sync.RWMutex
	reloadCallbacks []func(*Config, *Config)
	watcher         *fsnotify.Watcher
)

type ConsulConfig struct {
	Host       string `yaml:"host"`
	Datacenter string `yaml:"datacenter"`
	Service    struct {
		Name string   `yaml:"name"`
		Tags []string `yaml:"tags"`
		Port int      `yaml:"port"`
		ID   string   `yaml:"id"`
	} `yaml:"service"`
	Services []string `yaml:"services"`
	Keys     []string `yaml:"keys"`
}

type LogConfig struct {
	Debug      bool   `yaml:"debug"`
	Dir        string `yaml:"log_dir"`
	MaxSize    int    `yaml:"log_max_size"`
	MaxBackups int    `yaml:"log_max_backups"`
	MaxAge     int    `yaml:"log_max_age"`
	Compress   bool   `yaml:"log_compress"`
}

type Config struct {
	Consul ConsulConfig `yaml:"consul"`
	IP     string       `yaml:"ip"`
	Env    string       `yaml:"env"`
	Log    LogConfig    `yaml:"log"`
}

func Init() error {
	cfg, configPath, err := loadConfig()
	if err != nil {
		return err
	}

	configMutex.Lock()
	GlobalConfig = cfg
	configMutex.Unlock()

	// 初始化文件监听
	if err := initWatcher(configPath); err != nil {
		return fmt.Errorf("failed to init watcher: %w", err)
	}

	return nil
}

func loadConfig() (*Config, string, error) {
	hostname, _ := os.Hostname()
	hostname = strings.ToLower(hostname)

	searchPaths := []string{
		fmt.Sprintf("app-%s.yaml", hostname),
		"app-dev.yaml",
		"app.yaml",
	}

	var configFile string
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			break
		}
	}

	if configFile == "" {
		return nil, "", fmt.Errorf("no config file found")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, "", err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, "", err
	}

	return &cfg, configFile, nil
}

func initWatcher(configPath string) error {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// 监听配置文件目录
	configDir := filepath.Dir(configPath)
	if err := watcher.Add(configDir); err != nil {
		return err
	}

	go watchLoop(configPath)
	return nil
}

func watchLoop(configPath string) {
	defer watcher.Close()

	// 防抖动定时器
	var (
		timer     *time.Timer
		lastEvent time.Time
	)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// 只处理当前配置文件的变化
			if filepath.Clean(event.Name) != configPath {
				continue
			}

			// 过滤临时文件事件
			if event.Op.Has(fsnotify.Chmod) {
				continue
			}

			// 防抖动处理（500ms内的事件合并）
			now := time.Now()
			if timer != nil {
				timer.Stop()
			}

			// 忽略1秒内的重复事件
			if now.Sub(lastEvent) < time.Second {
				continue
			}

			timer = time.AfterFunc(500*time.Millisecond, func() {
				reloadConfig(configPath)
				lastEvent = time.Now()
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Config watcher error: %v\n", err)
		}
	}
}

func reloadConfig(configPath string) {
	newCfg, _, err := loadConfig()
	if err != nil {
		fmt.Printf("Reload config failed: %v\n", err)
		return
	}

	if err := validateConfig(newCfg); err != nil {
		fmt.Printf("Invalid config: %v\n", err)
		return
	}

	configMutex.Lock()
	oldConfig := GlobalConfig
	GlobalConfig = newCfg
	configMutex.Unlock()

	// 执行回调函数
	for _, cb := range reloadCallbacks {
		go cb(oldConfig, newCfg)
	}
}

func validateConfig(cfg *Config) error {
	if cfg.Consul.Host == "" {
		return fmt.Errorf("consul.host is required")
	}
	if cfg.Consul.Service.Name == "" {
		return fmt.Errorf("consul.service.name is required")
	}
	return nil
}

// 注册热加载回调函数
func RegisterReloadCallback(cb func(old, new *Config)) {
	reloadCallbacks = append(reloadCallbacks, cb)
}

// 安全获取配置的副本
func Get() Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return *GlobalConfig
}
