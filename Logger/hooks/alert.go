package hooks

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// AlertFunc 告警回调函数类型
type AlertFunc func(level Level, count int64, message string)

// AlertHook 告警钩子
// 当指定级别的日志在时间窗口内达到阈值时触发告警
type AlertHook struct {
	level     Level         // 监控的日志级别
	threshold int64         // 阈值
	window    time.Duration // 时间窗口
	alertFunc AlertFunc     // 告警回调
	counter   int64         // 当前计数
	lastReset time.Time     // 上次重置时间
	mu        sync.Mutex
}

// NewAlertHook 创建告警钩子
// level: 监控的日志级别（如 ErrorLevel）
// threshold: 阈值（如 10 表示 10 条日志）
// window: 时间窗口（如 1分钟）
// alertFunc: 触发告警时的回调函数
func NewAlertHook(level Level, threshold int64, window time.Duration, alertFunc AlertFunc) *AlertHook {
	hook := &AlertHook{
		level:     level,
		threshold: threshold,
		window:    window,
		alertFunc: alertFunc,
		lastReset: time.Now(),
	}

	// 启动定时重置
	go hook.resetLoop()

	return hook
}

// Levels 返回关心的级别
func (h *AlertHook) Levels() []Level {
	return []Level{h.level}
}

// Fire 增加计数并检查是否需要告警
func (h *AlertHook) Fire(entry *Entry) error {
	count := atomic.AddInt64(&h.counter, 1)

	// 检查是否达到阈值
	if count >= h.threshold {
		h.mu.Lock()
		// 双重检查，避免重复告警
		if atomic.LoadInt64(&h.counter) >= h.threshold {
			message := fmt.Sprintf("日志级别 %s 在 %v 内达到 %d 条",
				levelToString(h.level), h.window, count)

			// 触发告警
			if h.alertFunc != nil {
				go h.alertFunc(h.level, count, message)
			}

			// 重置计数器，避免频繁告警
			atomic.StoreInt64(&h.counter, 0)
			h.lastReset = time.Now()
		}
		h.mu.Unlock()
	}

	return nil
}

// resetLoop 定时重置计数器
func (h *AlertHook) resetLoop() {
	ticker := time.NewTicker(h.window)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.Lock()
		atomic.StoreInt64(&h.counter, 0)
		h.lastReset = time.Now()
		h.mu.Unlock()
	}
}

// GetCount 获取当前计数
func (h *AlertHook) GetCount() int64 {
	return atomic.LoadInt64(&h.counter)
}
