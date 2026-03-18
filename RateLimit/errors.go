package RateLimit

import "errors"

// ========== 限流相关错误 ==========

var (
	// ErrRateLimitExceeded 超过限流阈值
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrInvalidKey 无效的限流键
	ErrInvalidKey = errors.New("invalid rate limit key")

	// ErrInvalidConfig 无效的配置
	ErrInvalidConfig = errors.New("invalid rate limit config")

	// ErrStorageUnavailable 存储不可用
	ErrStorageUnavailable = errors.New("storage unavailable")

	// ErrNotImplemented 功能未实现
	ErrNotImplemented = errors.New("not implemented")
)

// ========== 存储相关错误 ==========

var (
	// ErrRedisConnectionFailed Redis 连接失败
	ErrRedisConnectionFailed = errors.New("redis connection failed")

	// ErrRedisOperationFailed Redis 操作失败
	ErrRedisOperationFailed = errors.New("redis operation failed")

	// ErrCacheMiss 缓存未命中
	ErrCacheMiss = errors.New("cache miss")
)

// ========== 算法相关错误 ==========

var (
	// ErrUnsupportedAlgorithm 不支持的算法
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")

	// ErrInvalidWindow 无效的时间窗口
	ErrInvalidWindow = errors.New("invalid time window")

	// ErrInvalidLimit 无效的限流上限
	ErrInvalidLimit = errors.New("invalid limit")
)
