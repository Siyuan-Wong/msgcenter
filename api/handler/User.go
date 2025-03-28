package handler

import (
	"github.com/gofiber/fiber/v2"
	"msgcenter/app/userService"
)

func InitUser(app *fiber.App) {
	app.Get("/user", createUserDemo)
}

func createUserDemo(c *fiber.Ctx) error {
	return userService.USER().InsertDemo(c.Context())
}
