package server

import (
	"log/slog"
	"msgcenter/config"
)

func (s *Server) staticConfigLoader() {
	err := config.Init()
	if err != nil {
		slog.Error("初始化配置失败", err)
		panic(err)
	}
	s.LocalConfig = config.GlobalConfig
	slog.Info("初始化配置成功", slog.Any("config", s.LocalConfig))
}
