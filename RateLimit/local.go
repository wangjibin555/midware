package RateLimit

import (
	"context"
	"sync"
	"time"
)

// ========== 本地内存限流器 ==========

// LocalLimiter 本地内存限流器（使用滑动窗口算法）
type LocalLimiter struct {
	limit  int64
	window time.Duration
	store  *localStore
	mu     sync.RWMutex
}

// NewLocalLimiter 创建本地限流器
func NewLocalLimiter(limit int64, window time.Duration) *LocalLimiter {
	return &LocalLimiter{
		limit:  limit,
		window: window,
		store:  newLocalStore(),
	}
}

// Allow 检查是否允许
func (l *LocalLimiter) Allow(ctx context.Context, key string) (*Result, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-l.window)

	// 获取窗口内的请求
	entry := l.store.getOrCreate(key)
	entry.mu.Lock()
	defer entry.mu.Unlock()

	// 清理过期请求
	l.cleanExpired(entry, windowStart)

	// 检查是否超限
	current := int64(len(entry.timestamps))
	if current >= l.limit {
		return &Result{
			Allowed:     false,
			Limit:       l.limit,
			Current:     current,
			Remaining:   0,
			ResetAt:     entry.timestamps[0].Add(l.window),
			RetryAfter:  time.Until(entry.timestamps[0].Add(l.window)),
			RateLimited: true,
		}, nil
	}

	// 允许通过，记录时间戳
	entry.timestamps = append(entry.timestamps, now)

	return &Result{
		Allowed:     true,
		Limit:       l.limit,
		Current:     current + 1,
		Remaining:   l.limit - current - 1,
		ResetAt:     now.Add(l.window),
		RetryAfter:  0,
		RateLimited: false,
	}, nil
}

// Take 获取令牌（阻塞版本）
func (l *LocalLimiter) Take(ctx context.Context, key string) error {
	for {
		result, err := l.Allow(ctx, key)
		if err != nil {
			return err
		}
		if result.Allowed {
			return nil
		}

		// 等待重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(result.RetryAfter):
			// 继续尝试
		}
	}
}

// Reset 重置限流器
func (l *LocalLimiter) Reset(ctx context.Context, key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.store.delete(key)
	return nil
}

// GetStats 获取统计信息
func (l *LocalLimiter) GetStats(ctx context.Context, key string) (*Stats, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	entry := l.store.get(key)
	if entry == nil {
		return &Stats{
			Limit: l.limit,
		}, nil
	}

	entry.mu.RLock()
	defer entry.mu.RUnlock()

	now := time.Now()
	windowStart := now.Add(-l.window)

	// 计算窗口内的请求数
	count := int64(0)
	for _, ts := range entry.timestamps {
		if ts.After(windowStart) {
			count++
		}
	}

	return &Stats{
		TotalRequests:   entry.total,
		AllowedRequests: entry.allowed,
		BlockedRequests: entry.blocked,
		CurrentUsage:    count,
		Limit:           l.limit,
		WindowStart:     windowStart,
		WindowEnd:       now,
	}, nil
}

// cleanExpired 清理过期的时间戳
func (l *LocalLimiter) cleanExpired(entry *localEntry, windowStart time.Time) {
	validIdx := 0
	for i, ts := range entry.timestamps {
		if ts.After(windowStart) {
			validIdx = i
			break
		}
	}
	entry.timestamps = entry.timestamps[validIdx:]
}

// ========== 本地存储 ==========

type localStore struct {
	entries map[string]*localEntry
	mu      sync.RWMutex
}

type localEntry struct {
	timestamps []time.Time // 请求时间戳列表
	total      int64       // 总请求数
	allowed    int64       // 允许的请求数
	blocked    int64       // 被拒绝的请求数
	mu         sync.RWMutex
}

func newLocalStore() *localStore {
	return &localStore{
		entries: make(map[string]*localEntry),
	}
}

func (s *localStore) getOrCreate(key string) *localEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, ok := s.entries[key]; ok {
		return entry
	}

	entry := &localEntry{
		timestamps: make([]time.Time, 0),
	}
	s.entries[key] = entry
	return entry
}

func (s *localStore) get(key string) *localEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.entries[key]
}

func (s *localStore) delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

// ========== 定期清理（可选） ==========

// StartCleanup 启动定期清理
func (l *LocalLimiter) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			l.cleanup()
		}
	}()
}

func (l *LocalLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-l.window * 2) // 清理 2 倍窗口之前的数据

	for key, entry := range l.store.entries {
		entry.mu.Lock()
		// 如果所有时间戳都过期，删除整个条目
		if len(entry.timestamps) == 0 || entry.timestamps[len(entry.timestamps)-1].Before(windowStart) {
			delete(l.store.entries, key)
		}
		entry.mu.Unlock()
	}
}
