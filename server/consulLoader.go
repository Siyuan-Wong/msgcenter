package server

import (
	consulapi "github.com/hashicorp/consul/api"
	"log/slog"
	"msgcenter/platform/consul"
)

func (s *Server) consulLoader() {
	client, err := consul.NewClient(s.LocalConfig.Consul.Host, s.Logger)
	if err != nil {
		slog.Error("初始化consul失败", err)
		panic(err)
	}
	s.Consul = client

	err = s.Consul.RegisterService(&consulapi.AgentServiceRegistration{
		ID:   s.LocalConfig.Consul.Service.ID,
		Name: s.LocalConfig.Consul.Service.Name,
		Port: s.LocalConfig.Consul.Service.Port,
		Tags: s.LocalConfig.Consul.Service.Tags,
		Check: &consulapi.AgentServiceCheck{
			HTTP:     "http://" + s.LocalConfig.IP + "/health", // 健康检查端点
			Method:   "GET",                                    // 请求方法
			Interval: "10s",                                    // 检查间隔
			Timeout:  "3s",                                     // 超时时间
			Header: map[string][]string{ // 自定义请求头
				"X-Consul-Check": {"true"},
			},
			TLSSkipVerify:          true,                    // 跳过TLS验证
			SuccessBeforePassing:   3,                       // 连续成功次数标记为健康
			FailuresBeforeCritical: 3,                       // 连续失败次数标记为故障
			Status:                 consulapi.HealthPassing, // 初始状态
		},
	})
	if err != nil {
		slog.Error("注册consul服务失败", err)
		panic(err)
	}

	slog.Info("注册consul服务成功", slog.Any("service", s.LocalConfig.Consul.Service))

	s.Consul.StartDynamicWatch(consul.WatchConfig{
		Services: s.LocalConfig.Consul.Services,
		Keys:     s.LocalConfig.Consul.Keys,
	})

	slog.Info("开始监听consul服务", slog.Any("services", s.LocalConfig.Consul.Services))
}

func (s *Server) DeregisterConsul() {
	if s.Consul != nil {
		err := s.Consul.DeregisterService(s.LocalConfig.Consul.Service.ID)
		if err != nil {
			return
		}
	}
}
