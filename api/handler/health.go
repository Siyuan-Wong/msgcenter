package handler

import "github.com/gofiber/fiber/v2"

func InitHealth(app *fiber.App) {
	app.Get("/health", health).Name("consul健康检测")
}

func health(c *fiber.Ctx) error {
	return c.SendStatus(200)
}
