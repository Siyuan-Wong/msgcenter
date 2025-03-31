package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"log/slog"
	"msgcenter/app"
	"msgcenter/config"
	"msgcenter/platform/consul"
	"msgcenter/platform/ent/gen"
	"sync"
)

var (
	Instance *Server
	Once     sync.Once
)

type Server struct {
	App         *fiber.App
	LocalConfig *config.Config
	Consul      *consul.Client
	DbClient    *gen.Client
	RedisClient *redis.Client
	Service     *app.ServiceApp
	Logger      *zap.Logger
}

func GetServer() *Server {
	Once.Do(func() {
		Instance = &Server{}
	})
	return Instance
}

func (s *Server) init() *Server {
	s.staticConfigLoader()
	s.loadLogger()
	s.consulLoader()
	s.redisLoader()
	s.registerRedisClientUpdate()

	s.dbLoader()
	s.registerDbClientUpdate()
	s.serviceLoader()
	s.fiberLoader()
	s.loadBanner()
	return s
}

func (s *Server) Start() {
	ExitLoader()
	s.init()
	err := s.App.Listen(s.LocalConfig.IP)
	if err != nil {
		slog.Error("启动服务失败", err)
		panic(err)
	}
	slog.Info("启动服务成功", slog.Any("ip", s.LocalConfig.IP))
}
