package banner

import (
	"github.com/common-nighthawk/go-figure"
	"go.uber.org/zap"
	"time"
)

func Show(appName string, flag string, ip string, pid int, handlerList []string, logger *zap.Logger) {

	// 使用 go-figure 创建 ASCII art 标题
	title := figure.NewFigure(appName, "", true)
	title.Print()

	now := time.Now()

	// 打印基本信息
	logger.Info("") // 空行
	logger.Info("➜ " + flag)
	logger.Info("➜ ", zap.Int("PID", pid))
	logger.Info("➜ ", zap.String("Time", now.Format("2006-01-02 15:04:05")))
	logger.Info("➜ ", zap.String("Listening on", "http://"+ip))
	logger.Info("") // 空行

	logger.Info("Application starting",
		zap.String("name", appName),
		zap.String("flag", flag),   // Assuming flag is boolean, adjust type if needed
		zap.Int("pid", pid),        // Assuming pid is int, adjust type if needed
		zap.String("address", ip),  // More efficient than fmt.Sprintf
		zap.Time("startTime", now), // zap will format time appropriately
	)

	// 打印路由列表
	if len(handlerList) > 0 {
		// 先打印美观的标题
		logger.Info("Registered routes:")

		// 为每个路由单独记录一行
		for _, route := range handlerList {
			logger.Info("  → " + route) // 使用统一前缀保持美观
		}

		// 同时保留结构化数据（可选）
		logger.Debug("Routes details",
			zap.Int("count", len(handlerList)),
			zap.Strings("routes", handlerList),
		)
	}

	banner := generateCoolBanner(appName)

	logger.Info(banner)
}

func generateCoolBanner(appName string) string {
	banner := `
╔══════════════════════════════════════════╗
║                                          ║
║    🚀 ` + appName + ` 服务已成功启动! 🚀        ║
║                                          ║
╚══════════════════════════════════════════╝
======================================================================
`
	return banner
}
