package server

import (
	"entgo.io/ent/dialect/sql"
	"log/slog"
	"msgcenter/ent"
)

func (s *Server) dbLoader() {
	config := s.Consul.GetSqldb()
	dsn := config.Dsn()
	drv, err := ent.Open("postgres", dsn)
	if err != nil {
		slog.Error("failed opening connection to postgres: %v", err)
		panic(err)
	}

}
