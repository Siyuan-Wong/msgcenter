package banner

import (
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"log/slog"
	"os"
	"time"
)

func Show(appName string, flag string, ip string, pid int, handlerList []string) {
	// 创建 slog 日志记录器
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// 使用 go-figure 创建 ASCII art 标题
	title := figure.NewFigure(appName, "", true)
	title.Print()

	// 打印基本信息
	fmt.Println()
	fmt.Printf("➜ %s\n", flag)
	fmt.Printf("➜ PID: %d\n", pid)
	fmt.Printf("➜ Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("➜ Listening on: http://%s\n", ip)
	fmt.Println()

	// 使用 slog 记录日志
	logger.Info("Application starting",
		"name", appName,
		"flag", flag,
		"pid", pid,
		"address", fmt.Sprintf("%s", ip),
		"startTime", time.Now().Format(time.RFC3339),
	)

	// 打印路由列表
	if len(handlerList) > 0 {
		fmt.Println("Registered routes:")
		for _, route := range handlerList {
			fmt.Printf("  - %s\n", route)
		}
		fmt.Println()

		logger.Info("Registered routes", "count", len(handlerList))
	}
}
