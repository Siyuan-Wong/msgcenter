package server

import (
	"github.com/gofiber/fiber/v2"
	"msgcenter/api"
)

func (s *Server) fiberLoader() {
	s.App = fiber.New(fiber.Config{
		AppName:               "msgcenter",
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// 处理错误
			return c.JSON(fiber.Map{
				"code": 500,
				"msg":  err.Error(),
			})
		},
	})
	api.InitRouter(s.App)
}
