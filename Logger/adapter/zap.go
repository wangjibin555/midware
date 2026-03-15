package adapter

import (
	"github.com/wangjibin555/midware/Logger"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapAdapter Zap 日志库适配器
type ZapAdapter struct {
	logger *zap.Logger
}

// NewZapAdapter 创建 Zap 适配器
// enableCaller: 是否记录调用者信息（文件名、行号）
// enableStackTrace: 是否记录堆栈信息（Error 级别以上）
func NewZapAdapter(enableCaller, enableStackTrace bool) *ZapAdapter {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var options []zap.Option

	// 允许记录调用者信息
	if enableCaller {
		options = append(options, zap.AddCaller())
		options = append(options, zap.AddCallerSkip(3)) // 跳过包装层
	}

	// 允许记录堆栈信息
	if enableStackTrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	logger, _ := config.Build(options...)
	return &ZapAdapter{
		logger: logger,
	}
}

// NewZapAdapterWithConfig 使用自定义配置创建 Zap 适配器
func NewZapAdapterWithConfig(config zap.Config, options ...zap.Option) *ZapAdapter {
	logger, _ := config.Build(options...)
	return &ZapAdapter{
		logger: logger,
	}
}

// NewZapAdapterWithLogger 使用已有的 zap.Logger 创建适配器
func NewZapAdapterWithLogger(logger *zap.Logger) *ZapAdapter {
	return &ZapAdapter{
		logger: logger,
	}
}

// Log 实现 Logger.Adapter 接口
func (a *ZapAdapter) Log(level Logger.Level, msg string, fields []Logger.Field) {
	zapFields := convertFieldsToZap(fields)

	// 根据级别调用对应方法
	switch level {
	case Logger.TraceLevel, Logger.DebugLevel:
		a.logger.Debug(msg, zapFields...)
	case Logger.InfoLevel:
		a.logger.Info(msg, zapFields...)
	case Logger.WarnLevel:
		a.logger.Warn(msg, zapFields...)
	case Logger.ErrorLevel:
		a.logger.Error(msg, zapFields...)
	case Logger.FatalLevel:
		a.logger.Fatal(msg, zapFields...)
	case Logger.PanicLevel:
		a.logger.Panic(msg, zapFields...)
	}
}

// Sync 同步日志缓冲区
func (a *ZapAdapter) Sync() error {
	return a.logger.Sync()
}

// GetZapLogger 获取底层的 zap.Logger（用于高级用法）
func (a *ZapAdapter) GetZapLogger() *zap.Logger {
	return a.logger
}

// convertFieldsToZap 转换字段列表
func convertFieldsToZap(fields []Logger.Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = convertFieldToZap(field)
	}
	return zapFields
}

// convertFieldToZap 转换单个字段（类型安全）
func convertFieldToZap(field Logger.Field) zap.Field {
	switch field.Type {
	case Logger.StringType:
		return zap.String(field.Key, field.Value.(string))
	case Logger.IntType:
		return zap.Int(field.Key, field.Value.(int))
	case Logger.Int64Type:
		return zap.Int64(field.Key, field.Value.(int64))
	case Logger.FloatType:
		return zap.Float64(field.Key, field.Value.(float64))
	case Logger.BoolType:
		return zap.Bool(field.Key, field.Value.(bool))
	case Logger.TimeType:
		return zap.Time(field.Key, field.Value.(time.Time))
	case Logger.DurationType:
		return zap.Duration(field.Key, field.Value.(time.Duration))
	case Logger.ErrorType:
		if err, ok := field.Value.(error); ok {
			return zap.NamedError(field.Key, err)
		}
		return zap.Skip()
	case Logger.AnyType:
		return zap.Any(field.Key, field.Value)
	default:
		return zap.Any(field.Key, field.Value)
	}
}
