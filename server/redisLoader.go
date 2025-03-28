package server

import (
	"github.com/redis/go-redis/v9"
	"log/slog"
	"msgcenter/platform/consul"
)

func (s *Server) redisLoader() {
	cfg := s.Consul.GetRedis()
	s.RedisClient = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		DB:           cfg.DB,
		Password:     cfg.Password,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxIdleConns: cfg.MaxIdleConns,
		PoolTimeout:  cfg.PoolTimeout,
	})
	slog.Info("redis已加载", slog.Any("cfg", cfg))
}

func (s *Server) closeRedis() {
	err := s.RedisClient.Close()
	if err != nil {
		slog.Error("redis已关闭", err.Error())
		panic(err)
	}
}

func (s *Server) updateRedisClient() {
	s.closeRedis()
	s.redisLoader()
}

func (s *Server) registerRedisClientUpdate() {
	s.Consul.RegisterCallback(consul.Redis, func(oldData, newData []byte) {
		slog.Info("redis数据库配置更新", "old", string(oldData), "new", string(newData))
		s.updateRedisClient()
		slog.Info("redis数据库配置更新完成")
	})
}
