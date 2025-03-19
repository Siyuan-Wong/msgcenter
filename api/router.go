package api

import (
	"github.com/gofiber/fiber/v2"
	"msgcenter/api/handler"
)

func InitRouter(app *fiber.App) {
	handler.InitHealth(app)
}
