package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"msgcenter/app"
	"msgcenter/config"
	"msgcenter/ent"
	"msgcenter/platform/consul"
)

type Server struct {
	App         *fiber.App
	LocalConfig *config.Config
	Consul      *consul.Client
	DbClient    *ent.Client
	RedisClient *redis.Client
	Service     *app.ServiceApp
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) init() *Server {
	s.staticConfigLoader()
	s.consulLoader()
	s.redisLoader()
	s.registerRedisClientUpdate()

	s.dbLoader()
	s.registerDbClientUpdate()
	s.serviceLoader()
	s.fiberLoader()

	return s
}

func (s *Server) Start() {
	s.init()
	err := s.App.Listen(s.LocalConfig.Ip)
	if err != nil {
		slog.Error("启动服务失败", err)
		panic(err)
	}
	slog.Info("启动服务成功", slog.Any("ip", s.LocalConfig.Ip))
}
