package Logger

// Option 配置选项函数类型
type Option func(*logger)

// WithLevel 设置日志级别
func WithLevel(level Level) Option {
	return func(l *logger) {
		l.level = level
	}
}

// WithFields 设置初始字段
func WithInitialFields(fields ...Field) Option {
	return func(l *logger) {
		l.fields = append(l.fields, fields...)
	}
}

// WithCaller 设置是否记录调用者相关信息，包含行号、文件位置等
func WithCaller(enabled bool) Option {
	return func(l *logger) {
		l.enableCaller = enabled
	}
}

// WithStackTrace 设置是否记录堆栈信息（Error 级别以上）
func WithStackTrace(enabled bool) Option {
	return func(l *logger) {
		l.enableStack = enabled
	}
}
