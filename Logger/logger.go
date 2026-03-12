package Logger

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// 定义Level等级
type Level int8

const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	default:
		return "unknown"
	}
}

// Adapter 适配器接口，用于适配不同的日志库（Zap, Logrus等）
type Adapter interface {
	Log(level Level, msg string, fields []Field)
	Sync() error
}

// logger 日志记录器实现
type logger struct {
	adapter      Adapter         //底层日志Zap或者logrus
	level        Level           //当前日志级别
	fields       []Field         //链式调用（注意：改为复数形式）
	ctx          context.Context //关联的 context
	enableCaller bool            //是否记录调用者信息（文件:行号）
	enableStack  bool            //是否记录堆栈信息（Error级别以上）
	mu           sync.RWMutex    //保护并发访问
}

// New 创建新的 Logger 实例
func New(adapter Adapter, opts ...Option) Logger {
	l := &logger{
		adapter: adapter,
		level:   InfoLevel, // 默认级别
		fields:  make([]Field, 0),
	}

	// 应用选项
	for _, opt := range opts {
		opt(l)
	}

	return l
}

// clone 克隆 logger（用于链式调用，避免污染原实例）
func (l *logger) clone() *logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 复制字段切片
	newFields := make([]Field, len(l.fields))
	copy(newFields, l.fields)

	return &logger{
		adapter:      l.adapter,
		level:        l.level,
		fields:       newFields,
		ctx:          l.ctx,
		enableCaller: l.enableCaller,
		enableStack:  l.enableStack,
	}
}

// shouldLog 检查是否应该记录该级别的日志
func (l *logger) shouldLog(level Level) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return level >= l.level
}

// log 内部日志记录方法
func (l *logger) log(level Level, msg string, fields ...Field) {
	if !l.shouldLog(level) {
		return
	}

	l.mu.RLock()
	// 合并累积的字段和新字段
	allFields := make([]Field, 0, len(l.fields)+len(fields))
	allFields = append(allFields, l.fields...)
	allFields = append(allFields, fields...)
	l.mu.RUnlock()

	// 调用适配器记录日志
	l.adapter.Log(level, msg, allFields)

	// Fatal 级别退出程序
	if level == FatalLevel {
		l.adapter.Sync()
		os.Exit(1)
	}

	// Panic 级别触发 panic
	if level == PanicLevel {
		l.adapter.Sync()
		panic(msg)
	}
}

// Debug 记录 Debug 级别日志
func (l *logger) Debug(msg string, fields ...Field) {
	l.log(DebugLevel, msg, fields...)
}

// Info 记录 Info 级别日志
func (l *logger) Info(msg string, fields ...Field) {
	l.log(InfoLevel, msg, fields...)
}

// Warn 记录 Warn 级别日志
func (l *logger) Warn(msg string, fields ...Field) {
	l.log(WarnLevel, msg, fields...)
}

// Error 记录 Error 级别日志
func (l *logger) Error(msg string, fields ...Field) {
	l.log(ErrorLevel, msg, fields...)
}

// Fatal 记录 Fatal 级别日志并退出程序
func (l *logger) Fatal(msg string, fields ...Field) {
	l.log(FatalLevel, msg, fields...)
}

// Panic 记录 Panic 级别日志并触发 panic
func (l *logger) Panic(msg string, fields ...Field) {
	l.log(PanicLevel, msg, fields...)
}

// Debugf 格式化记录 Debug 级别日志
func (l *logger) Debugf(format string, args ...interface{}) {
	if !l.shouldLog(DebugLevel) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.log(DebugLevel, msg)
}

// Infof 格式化记录 Info 级别日志
func (l *logger) Infof(format string, args ...interface{}) {
	if !l.shouldLog(InfoLevel) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.log(InfoLevel, msg)
}

// Warnf 格式化记录 Warn 级别日志
func (l *logger) Warnf(format string, args ...interface{}) {
	if !l.shouldLog(WarnLevel) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.log(WarnLevel, msg)
}

// Errorf 格式化记录 Error 级别日志
func (l *logger) Errorf(format string, args ...interface{}) {
	if !l.shouldLog(ErrorLevel) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.log(ErrorLevel, msg)
}

// Fatalf 格式化记录 Fatal 级别日志并退出程序
func (l *logger) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.log(FatalLevel, msg)
}

// Panicf 格式化记录 Panic 级别日志并触发 panic
func (l *logger) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.log(PanicLevel, msg)
}

// WithField 添加单个字段（链式调用）
func (l *logger) WithField(field Field) Logger {
	newLogger := l.clone()
	newLogger.fields = append(newLogger.fields, field)
	return newLogger
}

// WithFields 添加多个字段（链式调用）
func (l *logger) WithFields(fields ...Field) Logger {
	newLogger := l.clone()
	newLogger.fields = append(newLogger.fields, fields...)
	return newLogger
}

// WithError 添加错误字段（链式调用）
func (l *logger) WithError(err error) Logger {
	if err == nil {
		return l
	}
	return l.WithField(Err("error", err))
}

// WithContext 关联 context（链式调用）
func (l *logger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}

	newLogger := l.clone()
	newLogger.ctx = ctx

	// TODO: 可以在这里自动提取 context 中的 TraceID、UserID 等字段
	// 例如：
	// if traceID := ctx.Value("trace_id"); traceID != nil {
	//     newLogger.fields = append(newLogger.fields, String("trace_id", traceID.(string)))
	// }

	return newLogger
}

// SetLevel 设置日志级别
func (l *logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel 获取当前日志级别
func (l *logger) GetLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// Sync 同步日志缓冲区
func (l *logger) Sync() error {
	return l.adapter.Sync()
}

// ============================================
// 全局单例模式
// ============================================

var (
	defaultLogger Logger
	once          sync.Once
)

// initDefault 初始化默认 logger（延迟初始化）
func initDefault() {
	// TODO: 这里需要一个默认的 Adapter 实现
	defaultLogger = New(nil, WithLevel(InfoLevel))
}

// Default 获取全局默认 logger
func Default() Logger {
	once.Do(initDefault)
	return defaultLogger
}

// SetDefault 设置全局默认 logger
func SetDefault(l Logger) {
	defaultLogger = l
}

// Debug 使用全局 logger 记录 Debug 日志
func Debug(msg string, fields ...Field) {
	Default().Debug(msg, fields...)
}

// Info 使用全局 logger 记录 Info 日志
func Info(msg string, fields ...Field) {
	Default().Info(msg, fields...)
}

// Warn 使用全局 logger 记录 Warn 日志
func Warn(msg string, fields ...Field) {
	Default().Warn(msg, fields...)
}

// Error 使用全局 logger 记录 Error 日志
func Error(msg string, fields ...Field) {
	Default().Error(msg, fields...)
}

// Fatal 使用全局 logger 记录 Fatal 日志
func Fatal(msg string, fields ...Field) {
	Default().Fatal(msg, fields...)
}

// Panic 使用全局 logger 记录 Panic 日志
func Panic(msg string, fields ...Field) {
	Default().Panic(msg, fields...)
}

// Debugf 使用全局 logger 格式化记录 Debug 日志
func Debugf(format string, args ...interface{}) {
	Default().Debugf(format, args...)
}

// Infof 使用全局 logger 格式化记录 Info 日志
func Infof(format string, args ...interface{}) {
	Default().Infof(format, args...)
}

// Warnf 使用全局 logger 格式化记录 Warn 日志
func Warnf(format string, args ...interface{}) {
	Default().Warnf(format, args...)
}

// Errorf 使用全局 logger 格式化记录 Error 日志
func Errorf(format string, args ...interface{}) {
	Default().Errorf(format, args...)
}

// Fatalf 使用全局 logger 格式化记录 Fatal 日志
func Fatalf(format string, args ...interface{}) {
	Default().Fatalf(format, args...)
}

// Panicf 使用全局 logger 格式化记录 Panic 日志
func Panicf(format string, args ...interface{}) {
	Default().Panicf(format, args...)
}

// WithField 使用全局 logger 添加字段
func WithField(field Field) Logger {
	return Default().WithField(field)
}

// WithFields 使用全局 logger 添加多个字段
func WithFields(fields ...Field) Logger {
	return Default().WithFields(fields...)
}

// WithError 使用全局 logger 添加错误字段
func WithError(err error) Logger {
	return Default().WithError(err)
}

// WithContext 使用全局 logger 关联 context
func WithContext(ctx context.Context) Logger {
	return Default().WithContext(ctx)
}
