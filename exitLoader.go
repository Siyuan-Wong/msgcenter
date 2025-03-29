package main

import (
	"github.com/tebeka/atexit"
	"msgcenter/server"
	"os"
	"os/signal"
	"syscall"
)

func exitLoader() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		atexit.Exit(1) // 触发钩子
	}()
	atexit.Register(cleanUp)
}

func cleanUp() {
	s := server.GetServer()
	s.Logger.Info("服务已退出")
	s.DeregisterConsul()
	s.Logger.Info("服务已从consul注销")
	s.CloseDb()
	s.Logger.Info("数据库已关闭")
	s.CloseRedis()
	s.Logger.Info("redis已关闭")
	_ = s.Logger.Sync()
}
