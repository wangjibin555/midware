package RateLimit

import "time"

// ========== 限流策略类型 ==========

// Algorithm 限流算法类型
type Algorithm string

const (
	// FixedWindow 固定窗口算法
	FixedWindow Algorithm = "fixed_window"
	// SlidingWindow 滑动窗口算法
	SlidingWindow Algorithm = "sliding_window"
	// TokenBucket 令牌桶算法
	TokenBucket Algorithm = "token_bucket"
	// LeakyBucket 漏桶算法
	LeakyBucket Algorithm = "leaky_bucket"
)

// ========== 限流结果 ==========

// Result 限流检查结果
type Result struct {
	Allowed     bool          // 是否允许通过
	Limit       int64         // 限流上限
	Current     int64         // 当前已使用数量
	Remaining   int64         // 剩余配额
	ResetAt     time.Time     // 重置时间
	RetryAfter  time.Duration // 建议重试时间
	RateLimited bool          // 是否被限流
}

// ========== 限流配置 ==========

// Config 限流配置（新版：两级阈值设计）
type Config struct {
	// ===== 全局限流（Redis）=====
	GlobalLimit  int64         // 全局限流上限（所有服务器共享）
	GlobalWindow time.Duration // 全局时间窗口

	// ===== 本地限流（预检 + 降级）=====
	LocalLimit  int64         // 本地限流上限（单服务器）
	LocalWindow time.Duration // 本地时间窗口

	// ===== 策略配置 =====
	Strategy LimitStrategy // 限流策略

	// ===== 其他配置 =====
	Algorithm          Algorithm // 限流算法
	KeyPrefix          string    // Redis 键前缀
	FallbackToLocal    bool      // Redis 故障时降级到本地
	LocalPreCheck      bool      // 是否启用本地预检（快速拒绝）
	AutoCalculateLocal bool      // 自动计算本地限流阈值
	ServerCount        int       // 服务器数量（用于自动计算）

	// ===== 令牌桶/漏桶参数 =====
	Burst    int64   // 突发流量大小（令牌桶算法使用）
	Rate     float64 // 令牌生成速率（令牌/秒）
	Capacity int64   // 桶容量（漏桶算法使用）
}

// LimitStrategy 限流策略
type LimitStrategy int

const (
	// StrategyRedisFirst Redis 优先（推荐）
	// Redis 作为权威，本地仅用于降级
	StrategyRedisFirst LimitStrategy = iota

	// StrategyLocalPreCheck 本地预检
	// 本地快速拒绝 + Redis 最终判断
	StrategyLocalPreCheck

	// StrategyLocalOnly 仅本地
	// 不使用 Redis（单机模式）
	StrategyLocalOnly
)

// DefaultConfig 默认配置（Redis 优先）
func DefaultConfig() *Config {
	return &Config{
		GlobalLimit:        100,
		GlobalWindow:       time.Minute,
		LocalLimit:         30, // 单服务器限流
		LocalWindow:        time.Minute,
		Strategy:           StrategyRedisFirst,
		Algorithm:          SlidingWindow,
		KeyPrefix:          "ratelimit:",
		FallbackToLocal:    true,  // 启用降级
		LocalPreCheck:      false, // 不启用预检
		AutoCalculateLocal: false,
		Burst:              10,
		Rate:               100.0 / 60.0,
		Capacity:           100,
	}
}

// ========== 预设配置 ==========

// StrictConfig 严格限流配置（低频率）
func StrictConfig() *Config {
	return &Config{
		GlobalLimit:     10,
		GlobalWindow:    time.Minute,
		LocalLimit:      10,
		LocalWindow:     time.Minute,
		Strategy:        StrategyLocalOnly,
		Algorithm:       SlidingWindow,
		KeyPrefix:       "ratelimit:",
		FallbackToLocal: false,
		Burst:           2,
	}
}

// RelaxedConfig 宽松限流配置（高频率）
func RelaxedConfig() *Config {
	return &Config{
		GlobalLimit:     1000,
		GlobalWindow:    time.Minute,
		LocalLimit:      1000,
		LocalWindow:     time.Minute,
		Strategy:        StrategyLocalOnly,
		Algorithm:       TokenBucket,
		KeyPrefix:       "ratelimit:",
		FallbackToLocal: false,
		Burst:           100,
		Rate:            1000.0 / 60.0,
		Capacity:        1000,
	}
}

// APIConfig API 限流配置（分布式）
func APIConfig() *Config {
	return &Config{
		GlobalLimit:        100,
		GlobalWindow:       time.Minute,
		LocalLimit:         30,
		LocalWindow:        time.Minute,
		Strategy:           StrategyRedisFirst,
		Algorithm:          SlidingWindow,
		KeyPrefix:          "api:ratelimit:",
		FallbackToLocal:    true,
		LocalPreCheck:      false,
		AutoCalculateLocal: false,
		Burst:              20,
	}
}

// LocalPreCheckConfig 本地预检配置（高性能）
func LocalPreCheckConfig(serverCount int) *Config {
	return &Config{
		GlobalLimit:        100,
		GlobalWindow:       time.Minute,
		LocalLimit:         0, // 自动计算
		LocalWindow:        time.Minute,
		Strategy:           StrategyLocalPreCheck,
		Algorithm:          SlidingWindow,
		KeyPrefix:          "ratelimit:",
		FallbackToLocal:    true,
		LocalPreCheck:      true,
		AutoCalculateLocal: true,
		ServerCount:        serverCount,
	}
}

// LocalOnlyConfig 仅本地配置（单机）
func LocalOnlyConfig() *Config {
	return &Config{
		GlobalLimit:        0,
		GlobalWindow:       0,
		LocalLimit:         100,
		LocalWindow:        time.Minute,
		Strategy:           StrategyLocalOnly,
		Algorithm:          SlidingWindow,
		FallbackToLocal:    false,
		LocalPreCheck:      false,
		AutoCalculateLocal: false,
	}
}

// ========== 限流维度 ==========

// Dimension 限流维度
type Dimension struct {
	Key   string // 限流键（如 user_id、ip、api_key）
	Value string // 限流值
}

// ByIP 按 IP 限流
func ByIP(ip string) string {
	return "ip:" + ip
}

// ByUserID 按用户 ID 限流
func ByUserID(userID string) string {
	return "user:" + userID
}

// ByAPIKey 按 API Key 限流
func ByAPIKey(apiKey string) string {
	return "apikey:" + apiKey
}

// ByEndpoint 按接口限流
func ByEndpoint(endpoint string) string {
	return "endpoint:" + endpoint
}

// Combined 组合多个维度
func Combined(parts ...string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ":"
		}
		result += part
	}
	return result
}

// ========== 统计信息 ==========

// Stats 限流统计信息
type Stats struct {
	TotalRequests   int64     // 总请求数
	AllowedRequests int64     // 允许的请求数
	BlockedRequests int64     // 被拒绝的请求数
	CurrentUsage    int64     // 当前使用量
	Limit           int64     // 限流上限
	WindowStart     time.Time // 窗口开始时间
	WindowEnd       time.Time // 窗口结束时间
}
