package hooks

import (
	"fmt"
	"io"
	"sync"
)

// WriterHook 将日志写入到指定的 io.Writer
// 可用于同时输出到多个目标（文件、网络等）
type WriterHook struct {
	writer io.Writer
	levels map[Level]bool
	mu     sync.Mutex
}

// NewWriterHook 创建写入钩子
func NewWriterHook(writer io.Writer, levels ...Level) *WriterHook {
	levelMap := make(map[Level]bool)
	if len(levels) == 0 {
		// 默认所有级别
		for l := DebugLevel; l <= PanicLevel; l++ {
			levelMap[l] = true
		}
	} else {
		for _, level := range levels {
			levelMap[level] = true
		}
	}

	return &WriterHook{
		writer: writer,
		levels: levelMap,
	}
}

// Levels 返回关心的日志级别
func (h *WriterHook) Levels() []Level {
	levels := make([]Level, 0, len(h.levels))
	for level := range h.levels {
		levels = append(levels, level)
	}
	return levels
}

// Fire 执行写入
func (h *WriterHook) Fire(entry *Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 简单格式化输出
	line := fmt.Sprintf("[%s] %s %s", entry.Time.Format("2006-01-02 15:04:05"), levelToString(entry.Level), entry.Message)

	if len(entry.Fields) > 0 {
		line += " |"
		for _, field := range entry.Fields {
			line += fmt.Sprintf(" %s=%v", field.Key, field.Value)
		}
	}
	line += "\n"

	_, err := h.writer.Write([]byte(line))
	return err
}

func levelToString(level Level) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case PanicLevel:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}
