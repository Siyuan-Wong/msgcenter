package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"time"
)

func InitZapLogger(app *fiber.App) {
	app.Use(ZapLogger())
}

func ZapLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// 处理请求
		err := c.Next()

		// 记录日志
		zap.L().Info("🫡🫡🫡🫡HTTP Request---",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
		)

		return err
	}
}
