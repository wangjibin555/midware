package adapter

import (
	"Logger"
	"time"

	"github.com/sirupsen/logrus"
)

// LogrusAdapter Logrus 日志库适配器
type LogrusAdapter struct {
	logger *logrus.Logger
}

// NewLogrusAdapter 创建 Logrus 适配器
func NewLogrusAdapter(enableCaller bool) *LogrusAdapter {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// 设置是否报告调用者信息
	logger.SetReportCaller(enableCaller)

	return &LogrusAdapter{
		logger: logger,
	}
}

// NewLogrusAdapterWithLogger 使用已有的 logrus.Logger 创建适配器
func NewLogrusAdapterWithLogger(logger *logrus.Logger) *LogrusAdapter {
	return &LogrusAdapter{
		logger: logger,
	}
}

// Log 实现 Logger.Adapter 接口
func (a *LogrusAdapter) Log(level Logger.Level, msg string, fields []Logger.Field) {
	logrusFields := convertFieldsToLogrus(fields)
	entry := a.logger.WithFields(logrusFields)

	// 根据级别调用对应方法
	switch level {
	case Logger.TraceLevel:
		entry.Trace(msg)
	case Logger.DebugLevel:
		entry.Debug(msg)
	case Logger.InfoLevel:
		entry.Info(msg)
	case Logger.WarnLevel:
		entry.Warn(msg)
	case Logger.ErrorLevel:
		entry.Error(msg)
	case Logger.FatalLevel:
		entry.Fatal(msg)
	case Logger.PanicLevel:
		entry.Panic(msg)
	}
}

// Sync 同步日志缓冲区（Logrus 不需要显式同步）
func (a *LogrusAdapter) Sync() error {
	return nil
}

// GetLogrusLogger 获取底层的 logrus.Logger（用于高级用法）
func (a *LogrusAdapter) GetLogrusLogger() *logrus.Logger {
	return a.logger
}

// SetFormatter 设置日志格式化器
func (a *LogrusAdapter) SetFormatter(formatter logrus.Formatter) {
	a.logger.SetFormatter(formatter)
}

// SetLevel 设置日志级别
func (a *LogrusAdapter) SetLevel(level Logger.Level) {
	a.logger.SetLevel(convertLevelToLogrus(level))
}

// AddHook 添加 Logrus Hook
func (a *LogrusAdapter) AddHook(hook logrus.Hook) {
	a.logger.AddHook(hook)
}

// convertFieldsToLogrus 转换字段列表
func convertFieldsToLogrus(fields []Logger.Field) logrus.Fields {
	if len(fields) == 0 {
		return nil
	}

	logrusFields := make(logrus.Fields, len(fields))
	for _, field := range fields {
		logrusFields[field.Key] = convertFieldValueToLogrus(field)
	}
	return logrusFields
}

// convertFieldValueToLogrus 转换字段值
func convertFieldValueToLogrus(field Logger.Field) interface{} {
	switch field.Type {
	case Logger.ErrorType:
		if err, ok := field.Value.(error); ok {
			return err.Error()
		}
		return field.Value
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
	default:
		return field.Value
	}
}

// convertLevelToLogrus 转换日志级别
func convertLevelToLogrus(level Logger.Level) logrus.Level {
	switch level {
	case Logger.TraceLevel:
		return logrus.TraceLevel
	case Logger.DebugLevel:
		return logrus.DebugLevel
	case Logger.InfoLevel:
		return logrus.InfoLevel
	case Logger.WarnLevel:
		return logrus.WarnLevel
	case Logger.ErrorLevel:
		return logrus.ErrorLevel
	case Logger.FatalLevel:
		return logrus.FatalLevel
	case Logger.PanicLevel:
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}
