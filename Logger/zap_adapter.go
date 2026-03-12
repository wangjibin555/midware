package Logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapAdapter struct {
	logger *zap.Logger
}

// 创建Zap适配器
func NewZapAdapter(enableCaller, enableStactTrace bool) *ZapAdapter {
	config := zap.NewProductionConfig()
	var options []zap.Option
	//允许记录调用者信息
	if enableCaller {
		options = append(options, zap.AddCaller())
		options = append(options, zap.AddCallerSkip(3)) //跳过包装层
	}
	//允许记录堆栈信息
	if enableStactTrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	logger, _ := config.Build(options...)
	return &ZapAdapter{
		logger: logger,
	}
}

// Sync 同步日志缓冲区
func (a *ZapAdapter) Sync() error {
	return a.logger.Sync()
}

// 实现logger里面的Adapter接口
func (a *ZapAdapter) Log(level Level, msg string, fields []Field) {
	convertLevel(level)
	zapFields := convertFields(fields)

	// 根据级别调用对应方法
	switch level {
	case DebugLevel:
		a.logger.Debug(msg, zapFields...)
	case InfoLevel:
		a.logger.Info(msg, zapFields...)
	case WarnLevel:
		a.logger.Warn(msg, zapFields...)
	case ErrorLevel:
		a.logger.Error(msg, zapFields...)
	case FatalLevel:
		a.logger.Fatal(msg, zapFields...)
	case PanicLevel:
		a.logger.Panic(msg, zapFields...)
	}
}

// convertLevel 转换日志级别
func convertLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	case PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// convertFields 转换字段列表
func convertFields(fields []Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = convertField(field)
	}
	return zapFields
}

// convertField 转换单个字段（关键：类型安全的转换）
func convertField(field Field) zap.Field {
	switch field.Type {
	case StringType:
		return zap.String(field.Key, field.Value.(string))
	case IntType:
		return zap.Int(field.Key, field.Value.(int))
	case Int64Type:
		return zap.Int64(field.Key, field.Value.(int64))
	case FloatType:
		return zap.Float64(field.Key, field.Value.(float64))
	case BoolType:
		return zap.Bool(field.Key, field.Value.(bool))
	case TimeType:
		return zap.Time(field.Key, field.Value.(time.Time))
	case DurationType:
		return zap.Duration(field.Key, field.Value.(time.Duration))
	case ErrorType:
		if err, ok := field.Value.(error); ok {
			return zap.NamedError(field.Key, err) // ← 使用 Key 作为字段名
		}
		return zap.Skip()
	case AnyType:
		return zap.Any(field.Key, field.Value)
	default:
		return zap.Any(field.Key, field.Value)
	}
}
