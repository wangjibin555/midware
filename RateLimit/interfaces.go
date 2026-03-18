package RateLimit

import "context"

// ========== 限流器接口 ==========

// Limiter 限流器接口
type Limiter interface {
	// Allow 检查是否允许通过（返回详细结果）
	Allow(ctx context.Context, key string) (*Result, error)

	// Take 获取一个令牌（阻塞直到有可用令牌或超时）
	Take(ctx context.Context, key string) error

	// Reset 重置限流器
	Reset(ctx context.Context, key string) error

	// GetStats 获取统计信息
	GetStats(ctx context.Context, key string) (*Stats, error)
}

// ========== 存储接口 ==========

// Store 限流存储接口
type Store interface {
	// Increment 增加计数
	// 返回：当前计数、错误
	Increment(ctx context.Context, key string, window int64) (int64, error)

	// Get 获取当前计数
	Get(ctx context.Context, key string) (int64, error)

	// Reset 重置计数
	Reset(ctx context.Context, key string) error

	// SetWithExpire 设置值并指定过期时间
	SetWithExpire(ctx context.Context, key string, value int64, expire int64) error

	// GetMulti 批量获取
	GetMulti(ctx context.Context, keys []string) (map[string]int64, error)
}

// ========== 算法接口 ==========

// Algorithm 限流算法接口
type AlgorithmImpl interface {
	// Allow 检查是否允许
	Allow(ctx context.Context, key string, limit int64, window int64) (*Result, error)

	// Reset 重置
	Reset(ctx context.Context, key string) error

	// Name 算法名称
	Name() string
}
