package server

import (
	"github.com/gofiber/fiber/v2"
	"msgcenter/api"
)

func (s *Server) fiberLoader() {
	s.App = fiber.New()
	api.InitRouter(s.App)
}
