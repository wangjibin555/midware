package adapter

import (
	"github.com/wangjibin555/midware/Logger"
	"os"
)

// AdapterType 适配器类型
type AdapterType string

const (
	AdapterTypeZap     AdapterType = "zap"
	AdapterTypeLogrus  AdapterType = "logrus"
	AdapterTypeConsole AdapterType = "console"
	AdapterTypeStdout  AdapterType = "stdout"
	AdapterTypeNoop    AdapterType = "noop"
)

// Config 适配器配置
type Config struct {
	Type         AdapterType // 适配器类型
	EnableCaller bool        // 是否启用调用者信息
	EnableStack  bool        // 是否启用堆栈跟踪
	EnableColor  bool        // 是否启用彩色输出（仅 Console）
}

// NewAdapter 工厂函数，根据配置创建适配器
func NewAdapter(config *Config) Logger.Adapter {
	if config == nil {
		// 默认配置：Console 适配器用于开发
		config = &Config{
			Type:         AdapterTypeConsole,
			EnableCaller: true,
			EnableColor:  true,
		}
	}

	switch config.Type {
	case AdapterTypeZap:
		return NewZapAdapter(config.EnableCaller, config.EnableStack)

	case AdapterTypeLogrus:
		return NewLogrusAdapter(config.EnableCaller)

	case AdapterTypeConsole:
		return NewConsoleAdapter(&ConsoleOptions{
			Writer:       os.Stdout,
			EnableColor:  config.EnableColor,
			EnableCaller: config.EnableCaller,
		})

	case AdapterTypeStdout:
		return NewStdoutAdapter()

	case AdapterTypeNoop:
		return NewNoopAdapter()

	default:
		// 默认使用 Console 适配器
		return NewConsoleAdapter(&ConsoleOptions{
			Writer:      os.Stdout,
			EnableColor: config.EnableColor,
		})
	}
}

// NewProductionAdapter 创建生产环境适配器（Zap）
func NewProductionAdapter() Logger.Adapter {
	return NewZapAdapter(true, true)
}

// NewDevelopmentAdapter 创建开发环境适配器（Console）
func NewDevelopmentAdapter() Logger.Adapter {
	return NewConsoleAdapter(&ConsoleOptions{
		Writer:       os.Stdout,
		EnableColor:  true,
		EnableCaller: true,
	})
}

// NewTestAdapter 创建测试环境适配器（Noop）
func NewTestAdapter() Logger.Adapter {
	return NewNoopAdapter()
}
