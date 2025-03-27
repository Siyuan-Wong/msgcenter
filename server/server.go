package server

import (
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"msgcenter/config"
	"msgcenter/platform/consul"
)

type Server struct {
	App         *fiber.App
	LocalConfig *config.Config
	Consul      *consul.Client
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) init() *Server {
	s.fiberLoader()
	s.staticConfigLoader()
	s.consulLoader()
	s.dbLoader()
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
