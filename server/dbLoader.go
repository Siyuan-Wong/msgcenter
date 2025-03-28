package server

import (
	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	"log/slog"
	"msgcenter/ent"
	"msgcenter/platform/consul"
	"time"
)

func (s *Server) dbLoader() {
	cfg := s.Consul.GetSqldb()
	drv, err := sql.Open("postgres", cfg.Dsn())
	if err != nil {
		slog.Error("数据库连接失败", "err", err)
		panic(err)
	}

	// 2. 配置连接池
	db := drv.DB()

	// 设置默认值
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.MaxOpenConns <= 0 {
		cfg.MaxOpenConns = 50
	}
	if cfg.ConnMaxLifetime <= 0 {
		cfg.ConnMaxLifetime = 1800 // 默认30分钟(秒)
	}
	if cfg.ConnMaxIdleTime <= 0 {
		cfg.ConnMaxIdleTime = 900 // 默认15分钟(秒)
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)

	// 3. 创建 Ent 客户端
	s.DbClient = ent.NewClient(ent.Driver(drv))
}

func (s *Server) closeDb() {
	if s.DbClient != nil {
		err := s.DbClient.Close()
		if err != nil {
			slog.Error("关闭数据库连接失败", "err", err)
			panic(err)
		}
	}
	s.dbLoader()
}

func (s *Server) updateDbClient() {
	if s.DbClient != nil {
		err := s.DbClient.Close()
		if err != nil {
			slog.Error("关闭数据库连接失败", "err", err)
		}
	}
}

func (s *Server) registerDbClientUpdate() {
	s.Consul.RegisterCallback(consul.Sqldb, func(oldData, newData []byte) {
		slog.Info("数据库配置更新", "old", string(oldData), "new", string(newData))
		s.updateDbClient()
		slog.Info("数据库配置更新完成")
	})
}
