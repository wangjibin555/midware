package Logger

import "context"

type contextKey string

const (
	loggerKey contextKey = "logger" // Logger 实例
)

// 将Logger实例添加到context中
func ToContext(ctx context.Context, logger Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, loggerKey, logger)
}

// 从context里面获取Logger实例
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return Default()
	}

	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}

	return Default()
}
