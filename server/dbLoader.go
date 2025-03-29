package server

import (
	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"msgcenter/platform/consul"
	"msgcenter/platform/ent/gen"
	"time"
)

func (s *Server) dbLoader() {
	cfg := s.Consul.GetSqldb()
	drv, err := sql.Open("postgres", cfg.Dsn())
	if err != nil {
		s.Logger.Error("数据库连接失败",
			zap.Error(err),
		)
		panic(err)
	}

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
	s.DbClient = gen.NewClient(gen.Driver(drv))
}

func (s *Server) CloseDb() {
	if s.DbClient != nil {
		err := s.DbClient.Close()
		if err != nil {
			s.Logger.Error("关闭数据库连接失败",
				zap.Error(err),
			)
			panic(err)
		}
	}
}

func (s *Server) updateDbClient() {
	if s.DbClient != nil {
		s.CloseDb()
	}
	s.dbLoader()
}

func (s *Server) registerDbClientUpdate() {
	s.Consul.RegisterCallback(consul.Sqldb, func(oldData, newData []byte) {
		s.Logger.Info("数据库配置更新",
			zap.ByteString("old", oldData),
			zap.ByteString("new", newData),
		)
		s.updateDbClient()
		s.Logger.Info("数据库配置更新完成")
	})
}
