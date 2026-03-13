package hooks

import (
	"sync"
	"sync/atomic"
)

// CounterHook 统计日志数量
// 可用于监控、告警等场景
type CounterHook struct {
	counters map[Level]*int64 // 每个级别的计数器
	mu       sync.RWMutex
}

// NewCounterHook 创建计数钩子
func NewCounterHook() *CounterHook {
	counters := make(map[Level]*int64)
	for level := DebugLevel; level <= PanicLevel; level++ {
		var count int64
		counters[level] = &count
	}

	return &CounterHook{
		counters: counters,
	}
}

// Levels 所有级别都统计
func (h *CounterHook) Levels() []Level {
	return []Level{
		DebugLevel,
		InfoLevel,
		WarnLevel,
		ErrorLevel,
		FatalLevel,
		PanicLevel,
	}
}

// Fire 增加计数
func (h *CounterHook) Fire(entry *Entry) error {
	h.mu.RLock()
	counter := h.counters[entry.Level]
	h.mu.RUnlock()

	if counter != nil {
		atomic.AddInt64(counter, 1)
	}
	return nil
}

// GetCount 获取指定级别的日志数量
func (h *CounterHook) GetCount(level Level) int64 {
	h.mu.RLock()
	counter := h.counters[level]
	h.mu.RUnlock()

	if counter != nil {
		return atomic.LoadInt64(counter)
	}
	return 0
}

// GetAllCounts 获取所有级别的日志数量
func (h *CounterHook) GetAllCounts() map[Level]int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[Level]int64)
	for level, counter := range h.counters {
		result[level] = atomic.LoadInt64(counter)
	}
	return result
}

// Reset 重置计数器
func (h *CounterHook) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, counter := range h.counters {
		atomic.StoreInt64(counter, 0)
	}
}
