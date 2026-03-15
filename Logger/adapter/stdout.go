package adapter

import (
	"fmt"
	"github.com/wangjibin555/midware/Logger"
	"os"
	"strings"
	"sync"
	"time"
)

// StdoutAdapter 简单的标准输出适配器
// 不依赖外部日志库，适合快速开发和测试
type StdoutAdapter struct {
	mu sync.Mutex
}

// NewStdoutAdapter 创建标准输出适配器
func NewStdoutAdapter() *StdoutAdapter {
	return &StdoutAdapter{}
}

// Log 实现 Logger.Adapter 接口
func (a *StdoutAdapter) Log(level Logger.Level, msg string, fields []Logger.Field) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 构建日志行
	var builder strings.Builder

	// 时间戳
	builder.WriteString(time.Now().Format("2006-01-02 15:04:05"))
	builder.WriteString(" ")

	// 级别
	builder.WriteString("[")
	builder.WriteString(strings.ToUpper(level.String()))
	builder.WriteString("]")
	builder.WriteString(" ")

	// 消息
	builder.WriteString(msg)

	// 字段
	if len(fields) > 0 {
		builder.WriteString(" ")
		for i, field := range fields {
			if i > 0 {
				builder.WriteString(" ")
			}
			builder.WriteString(field.Key)
			builder.WriteString("=")
			builder.WriteString(formatStdoutFieldValue(field))
		}
	}

	builder.WriteString("\n")

	// 输出
	fmt.Fprint(os.Stdout, builder.String())
}

// Sync 实现 Logger.Adapter 接口
func (a *StdoutAdapter) Sync() error {
	return nil
}

// formatStdoutFieldValue 格式化字段值
func formatStdoutFieldValue(field Logger.Field) string {
	switch field.Type {
	case Logger.StringType:
		return fmt.Sprintf(`"%v"`, field.Value)
	case Logger.ErrorType:
		if err, ok := field.Value.(error); ok {
			return fmt.Sprintf(`"%s"`, err.Error())
		}
		return fmt.Sprintf("%v", field.Value)
	case Logger.TimeType:
		if t, ok := field.Value.(time.Time); ok {
			return t.Format(time.RFC3339)
		}
		return fmt.Sprintf("%v", field.Value)
	case Logger.DurationType:
		if d, ok := field.Value.(time.Duration); ok {
			return d.String()
		}
		return fmt.Sprintf("%v", field.Value)
	default:
		return fmt.Sprintf("%v", field.Value)
	}
}
