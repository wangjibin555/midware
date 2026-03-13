package adapter

import (
	"Logger"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// ConsoleAdapter 简单的控制台输出适配器，适合开发环境使用
// 提供彩色、易读的日志输出
type ConsoleAdapter struct {
	writer       io.Writer
	enableColor  bool
	enableCaller bool
	mu           sync.Mutex
}

// ConsoleOptions 控制台适配器配置选项
type ConsoleOptions struct {
	Writer       io.Writer // 输出目标，默认为 os.Stdout
	EnableColor  bool      // 是否启用彩色输出
	EnableCaller bool      // 是否显示调用者信息
}

// NewConsoleAdapter 创建控制台适配器
func NewConsoleAdapter(opts *ConsoleOptions) *ConsoleAdapter {
	if opts == nil {
		opts = &ConsoleOptions{
			Writer:       os.Stdout,
			EnableColor:  true,
			EnableCaller: false,
		}
	}

	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	return &ConsoleAdapter{
		writer:       opts.Writer,
		enableColor:  opts.EnableColor,
		enableCaller: opts.EnableCaller,
	}
}

// Log 实现 Logger.Adapter 接口
func (a *ConsoleAdapter) Log(level Logger.Level, msg string, fields []Logger.Field) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 构建日志行
	var builder strings.Builder

	// 时间戳
	builder.WriteString(time.Now().Format("2006-01-02 15:04:05.000"))
	builder.WriteString(" ")

	// 日志级别（带颜色）
	levelStr := a.formatLevel(level)
	builder.WriteString(levelStr)
	builder.WriteString(" ")

	// 消息
	builder.WriteString(msg)

	// 字段
	if len(fields) > 0 {
		builder.WriteString(" ")
		builder.WriteString(a.formatFields(fields))
	}

	builder.WriteString("\n")

	// 输出
	fmt.Fprint(a.writer, builder.String())
}

// Sync 同步输出（控制台通常不需要）
func (a *ConsoleAdapter) Sync() error {
	if syncer, ok := a.writer.(interface{ Sync() error }); ok {
		return syncer.Sync()
	}
	return nil
}

// formatLevel 格式化日志级别
func (a *ConsoleAdapter) formatLevel(level Logger.Level) string {
	if !a.enableColor {
		return fmt.Sprintf("[%s]", strings.ToUpper(level.String()))
	}

	// ANSI 颜色码
	var color string
	switch level {
	case Logger.TraceLevel:
		color = "\033[37m" // 白色
	case Logger.DebugLevel:
		color = "\033[36m" // 青色
	case Logger.InfoLevel:
		color = "\033[32m" // 绿色
	case Logger.WarnLevel:
		color = "\033[33m" // 黄色
	case Logger.ErrorLevel:
		color = "\033[31m" // 红色
	case Logger.FatalLevel:
		color = "\033[35m" // 品红
	case Logger.PanicLevel:
		color = "\033[41m" // 红色背景
	default:
		color = "\033[0m" // 重置
	}

	reset := "\033[0m"
	return fmt.Sprintf("%s[%s]%s", color, strings.ToUpper(level.String()), reset)
}

// formatFields 格式化字段
func (a *ConsoleAdapter) formatFields(fields []Logger.Field) string {
	if len(fields) == 0 {
		return ""
	}

	var parts []string
	for _, field := range fields {
		value := a.formatFieldValue(field)
		parts = append(parts, fmt.Sprintf("%s=%v", field.Key, value))
	}

	return strings.Join(parts, " ")
}

// formatFieldValue 格式化字段值
func (a *ConsoleAdapter) formatFieldValue(field Logger.Field) interface{} {
	switch field.Type {
	case Logger.StringType:
		return fmt.Sprintf(`"%s"`, field.Value)
	case Logger.TimeType:
		if t, ok := field.Value.(time.Time); ok {
			return t.Format(time.RFC3339)
		}
		return field.Value
	case Logger.DurationType:
		if d, ok := field.Value.(time.Duration); ok {
			return d.String()
		}
		return field.Value
	case Logger.ErrorType:
		if err, ok := field.Value.(error); ok {
			return err.Error()
		}
		return field.Value
	default:
		return field.Value
	}
}

// SetWriter 设置输出目标
func (a *ConsoleAdapter) SetWriter(writer io.Writer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.writer = writer
}

// SetEnableColor 设置是否启用彩色输出
func (a *ConsoleAdapter) SetEnableColor(enable bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.enableColor = enable
}
