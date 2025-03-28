package handler

import "github.com/gofiber/fiber/v2"

func InitHealth(app *fiber.App) {
	app.Get("/health", health)
}

func health(c *fiber.Ctx) error {
	return c.SendStatus(200)
}

func InstanceInfo(c *fiber.Ctx) error {

}
