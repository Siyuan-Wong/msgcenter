package server

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

func (s *Server) loadLogger() *Server {
	// 确保日志目录存在
	if s.LocalConfig.Log.Dir != "" {
		if err := os.MkdirAll(s.LocalConfig.Log.Dir, 0755); err != nil {
			slog.Error("创建日志目录失败", "error", err, "dir", s.LocalConfig.Log.Dir)
			panic(err)
		}
	}

	// 配置日志轮转和归档
	logFile := filepath.Join(s.LocalConfig.Log.Dir, "app.log")
	logWriter, err := s.createLogWriter(logFile)
	if err != nil {
		slog.Error("创建日志写入器失败", "error", err)
		panic(err)
	}

	// 日志级别配置
	logLevel := zap.InfoLevel
	if s.LocalConfig.Log.Debug {
		logLevel = zap.DebugLevel
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if s.LocalConfig.Env == "development" {
		encoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 创建核心
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(logWriter),
		),
		zap.NewAtomicLevelAt(logLevel),
	)

	// 构建Logger
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	// 设置全局Logger
	zap.ReplaceGlobals(logger)
	s.Logger = logger

	// 兼容标准库log
	zap.RedirectStdLog(logger)
	log.SetFlags(0)

	logger.Info("日志系统初始化完成",
		zap.String("env", s.LocalConfig.Env),
		zap.String("log_level", logLevel.String()),
		zap.String("log_file", logFile),
		zap.Int("max_size", s.LocalConfig.Log.MaxSize),
		zap.Int("max_backups", s.LocalConfig.Log.MaxBackups),
		zap.Int("max_age", s.LocalConfig.Log.MaxAge),
		zap.Bool("compress", s.LocalConfig.Log.Compress),
	)

	return s
}

func (s *Server) createLogWriter(logFile string) (zapcore.WriteSyncer, error) {
	// 确保日志文件所在目录存在
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 配置日志轮转
	rotateWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    s.LocalConfig.Log.MaxSize,    // 单个日志文件最大大小(MB)
		MaxBackups: s.LocalConfig.Log.MaxBackups, // 保留的旧日志文件最大数量
		MaxAge:     s.LocalConfig.Log.MaxAge,     // 保留旧日志文件的最大天数
		Compress:   s.LocalConfig.Log.Compress,   // 是否压缩归档日志
		LocalTime:  true,                         // 使用本地时间
	}

	return zapcore.AddSync(rotateWriter), nil
}
