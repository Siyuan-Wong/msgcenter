package server

import "msgcenter/app"

func (s *Server) serviceLoader() {
	s.Service = app.GetService(s.LocalConfig, s.DbClient, s.Consul, s.RedisClient)
}
