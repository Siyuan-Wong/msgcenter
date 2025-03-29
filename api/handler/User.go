package handler

import (
	"github.com/gofiber/fiber/v2"
	"msgcenter/app/userService"
)

func InitUser(app *fiber.App) {
	app.Get("/user", createUserDemo).Name("demo")
}

func createUserDemo(c *fiber.Ctx) error {
	err := userService.USER().InsertDemo()
	if err != nil {
		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"code": 200,
		"msg":  "success",
	})
}
