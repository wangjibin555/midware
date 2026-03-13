package hooks

import (
	"time"
)

// Level 日志级别（避免循环依赖，在这里重新定义）
type Level int8

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// Entry 日志条目，传递给 Hook
type Entry struct {
	Level   Level     // 日志级别
	Message string    // 日志消息
	Fields  []Field   // 日志字段
	Time    time.Time // 日志时间
}

// Hook 日志钩子接口
type Hook interface {
	// Levels 返回该 Hook 关心的日志级别
	Levels() []Level

	// Fire 在日志记录时触发，返回 error 不会影响日志记录
	Fire(entry *Entry) error
}
