package main

import (
	"Logger"
	"Logger/adapter"
	"errors"
	"time"
)

func main() {
	// ============================================
	// 重要：必须先初始化全局 Logger（如果要使用全局方法）
	// ============================================
	initGlobalLogger()

	// ============================================
	// 现在可以使用全局方法了
	// ============================================
	Logger.Info("全局 Logger 已初始化", Logger.String("status", "ready"))

	// ============================================
	// 或者直接使用实例（推荐）
	// ============================================
	logger := Logger.New(adapter.NewConsoleAdapter(&adapter.ConsoleOptions{
		EnableColor: true,
	}))
	logger.Info("使用 Logger 实例", Logger.String("method", "recommended"))

	// 下面是各种适配器的示例
	demonstrateAdapters()
}

// initGlobalLogger 初始化全局 Logger（必须显式调用）
func initGlobalLogger() {
	consoleAdapter := adapter.NewConsoleAdapter(&adapter.ConsoleOptions{
		EnableColor:  true,
		EnableCaller: false,
	})
	globalLogger := Logger.New(consoleAdapter, Logger.WithLevel(Logger.InfoLevel))
	Logger.SetDefault(globalLogger)
}

func demonstrateAdapters() {
	// ============================================
	// 1. 使用 Zap 适配器（生产环境推荐）
	// ============================================
	zapAdapter := adapter.NewZapAdapter(true, true)
	zapLogger := Logger.New(zapAdapter, Logger.WithLevel(Logger.InfoLevel))

	zapLogger.Info("使用 Zap 适配器",
		Logger.String("adapter", "zap"),
		Logger.String("env", "production"),
	)

	// ============================================
	// 2. 使用控制台适配器（开发环境推荐）
	// ============================================
	consoleAdapter := adapter.NewConsoleAdapter(&adapter.ConsoleOptions{
		EnableColor:  true,
		EnableCaller: true,
	})
	consoleLogger := Logger.New(consoleAdapter, Logger.WithLevel(Logger.DebugLevel))

	consoleLogger.Debug("使用控制台适配器",
		Logger.String("adapter", "console"),
		Logger.String("env", "development"),
	)

	// ============================================
	// 3. 使用 Logrus 适配器（功能丰富）
	// ============================================
	logrusAdapter := adapter.NewLogrusAdapter(true)
	logrusLogger := Logger.New(logrusAdapter, Logger.WithLevel(Logger.InfoLevel))

	logrusLogger.Info("使用 Logrus 适配器",
		Logger.String("adapter", "logrus"),
		Logger.String("feature", "rich"),
	)

	// ============================================
	// 4. 使用工厂方法快速创建
	// ============================================
	// 生产环境
	prodLogger := Logger.New(adapter.NewProductionAdapter())
	prodLogger.Info("生产环境日志", Logger.String("type", "production"))

	// 开发环境
	devLogger := Logger.New(adapter.NewDevelopmentAdapter())
	devLogger.Debug("开发环境日志", Logger.String("type", "development"))

	// ============================================
	// 5. 链式调用示例
	// ============================================
	logger := Logger.New(consoleAdapter)

	logger.WithFields(
		Logger.String("service", "user-service"),
		Logger.String("version", "1.0.0"),
	).Info("服务启动")

	// 添加错误字段
	err := errors.New("数据库连接失败")
	logger.WithError(err).Error("操作失败")

	// ============================================
	// 6. 不同字段类型示例
	// ============================================
	logger.Info("字段类型演示",
		Logger.String("string", "值"),
		Logger.Int("int", 123),
		Logger.Int64("int64", 123456789),
		Logger.Float("float", 3.14159),
		Logger.Bool("bool", true),
		Logger.Time("time", time.Now()),
		Logger.Duration("duration", time.Second*5),
		Logger.Err("error", errors.New("示例错误")),
		Logger.Any("any", map[string]interface{}{"key": "value"}),
	)

	// ============================================
	// 7. 格式化日志示例
	// ============================================
	logger.Infof("用户 %s 登录成功，ID: %d", "张三", 1001)
	logger.Warnf("系统内存使用率: %.2f%%", 85.5)

	// ============================================
	// 8. 不同日志级别示例
	// ============================================
	logger.Debug("调试信息", Logger.String("level", "debug"))
	logger.Info("普通信息", Logger.String("level", "info"))
	logger.Warn("警告信息", Logger.String("level", "warn"))
	logger.Error("错误信息", Logger.String("level", "error"))

	// 同步日志缓冲区
	logger.Sync()
}
