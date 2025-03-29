package api

import (
	"github.com/gofiber/fiber/v2"
	"msgcenter/api/handler"
	"msgcenter/api/handler/middleware"
)

func InitRouter(app *fiber.App) {
	middleware.InitZapLogger(app)
	middleware.InitSonic(app)
	handler.InitHealth(app)
}
