package middleware

import (
	"bytes"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

func InitSonic(app *fiber.App) {
	app.Use(SonicJSONMiddleware())
}

func SonicJSONMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 只处理 JSON 请求
		if c.Get(fiber.HeaderContentType) == fiber.MIMEApplicationJSON {
			// 读取原始请求体
			body := c.Request().Body()
			reader := bytes.NewReader(body)

			// 使用 sonic 解码
			var jsonData interface{}
			if err := sonic.ConfigDefault.NewDecoder(reader).Decode(&jsonData); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON")
			}

			// 将解码后的数据存入 Locals 供后续使用
			c.Locals("jsonBody", jsonData)
		}

		// 继续处理请求
		err := c.Next()

		// 处理响应
		if c.Get(fiber.HeaderContentType) == fiber.MIMEApplicationJSON {
			// 获取响应数据
			response := c.Response().Body()

			// 使用 sonic 编码
			var buf bytes.Buffer
			if err := sonic.ConfigDefault.NewEncoder(&buf).Encode(response); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "JSON encoding failed")
			}

			c.Response().SetBodyRaw(buf.Bytes())
		}

		return err
	}
}
