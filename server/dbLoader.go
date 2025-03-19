package server

import (
	"github.com/bytedance/sonic"
	"log/slog"
	"msgcenter/platform/consul/config"
)

func (s *Server) dbLoader() {
	// 获取配置
	var config config.SqlDb
	err := sonic.Unmarshal(s.Consul.GetConfigValue("beijing.okr.sqldb.mattermost"), &config)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

}
