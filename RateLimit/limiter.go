package RateLimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ========== 正确的两层限流架构 ==========

// RateLimiter 组合限流器
type RateLimiter struct {
	config       *Config
	localLimiter *LocalLimiter
	redisLimiter *RedisLimiter
}

// New 创建限流器
func New(config *Config, redisClient ...*redis.Client) *RateLimiter {
	limiter := &RateLimiter{
		config: config,
	}

	// 自动计算本地限流阈值
	if config.AutoCalculateLocal && config.ServerCount > 0 {
		// 本地阈值 = 全局阈值 / 服务器数量 * 安全系数
		safetyFactor := 0.8 // 80% 安全系数
		config.LocalLimit = int64(float64(config.GlobalLimit) / float64(config.ServerCount) * safetyFactor)
	}

	// 创建本地限流器
	if config.Strategy == StrategyLocalOnly || config.FallbackToLocal || config.LocalPreCheck {
		limiter.localLimiter = NewLocalLimiter(config.LocalLimit, config.LocalWindow)
		limiter.localLimiter.StartCleanup(config.LocalWindow)
	}

	// 创建 Redis 限流器
	if config.Strategy != StrategyLocalOnly && len(redisClient) > 0 {
		limiter.redisLimiter = NewRedisLimiter(
			config.GlobalLimit,
			config.GlobalWindow,
			config.KeyPrefix,
			redisClient[0],
		)
	}

	return limiter
}

// Allow 检查是否允许（修正版）
func (r *RateLimiter) Allow(ctx context.Context, key string) (*Result, error) {
	switch r.config.Strategy {
	case StrategyRedisFirst:
		return r.allowRedisFirst(ctx, key)
	case StrategyLocalPreCheck:
		return r.allowLocalPreCheck(ctx, key)
	case StrategyLocalOnly:
		return r.allowLocalOnly(ctx, key)
	default:
		return r.allowRedisFirst(ctx, key)
	}
}

// allowRedisFirst Redis 优先策略
func (r *RateLimiter) allowRedisFirst(ctx context.Context, key string) (*Result, error) {
	// 1. 优先使用 Redis（权威数据源）
	if r.redisLimiter != nil {
		result, err := r.redisLimiter.Allow(ctx, key)

		// Redis 正常工作
		if err == nil {
			return result, nil
		}

		// Redis 故障，降级到本地
		if r.config.FallbackToLocal && r.localLimiter != nil {
			localResult, localErr := r.localLimiter.Allow(ctx, key)
			if localErr == nil {
				// 降级模式：返回本地限流结果
				return localResult, nil
			}
		}

		// 无法降级，返回错误
		return nil, fmt.Errorf("redis limiter failed and no fallback: %w", err)
	}

	// 没有 Redis，使用本地
	if r.localLimiter != nil {
		return r.localLimiter.Allow(ctx, key)
	}

	return nil, fmt.Errorf("no limiter available")
}

// allowLocalPreCheck 本地预检策略
func (r *RateLimiter) allowLocalPreCheck(ctx context.Context, key string) (*Result, error) {
	var localResult *Result

	// 1. 本地预检（快速拒绝）
	if r.config.LocalPreCheck && r.localLimiter != nil {
		var err error
		localResult, err = r.localLimiter.Allow(ctx, key)
		if err != nil {
			return nil, err
		}

		// 本地已经明显超限，快速拒绝
		if !localResult.Allowed {
			return localResult, nil
		}
		// 本地通过，继续检查 Redis
	}

	// 2. Redis 权威判断
	if r.redisLimiter != nil {
		result, err := r.redisLimiter.Allow(ctx, key)

		if err == nil {
			return result, nil
		}

		// Redis 故障，降级到本地结果
		if r.config.FallbackToLocal && localResult != nil {
			return localResult, nil
		}

		return nil, err
	}

	// 没有 Redis，返回本地结果
	if localResult != nil {
		return localResult, nil
	}

	return nil, fmt.Errorf("no limiter available")
}

// allowLocalOnly 仅本地策略
func (r *RateLimiter) allowLocalOnly(ctx context.Context, key string) (*Result, error) {
	if r.localLimiter != nil {
		return r.localLimiter.Allow(ctx, key)
	}
	return nil, fmt.Errorf("no local limiter available")
}

// Take 获取令牌（阻塞版本）
func (r *RateLimiter) Take(ctx context.Context, key string) error {
	for {
		result, err := r.Allow(ctx, key)
		if err != nil {
			return err
		}
		if result.Allowed {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(result.RetryAfter):
			// 继续尝试
		}
	}
}

// Reset 重置限流器
func (r *RateLimiter) Reset(ctx context.Context, key string) error {
	if r.localLimiter != nil {
		if err := r.localLimiter.Reset(ctx, key); err != nil {
			return err
		}
	}
	if r.redisLimiter != nil {
		if err := r.redisLimiter.Reset(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// GetStats 获取统计信息
func (r *RateLimiter) GetStats(ctx context.Context, key string) (*Stats, error) {
	// 优先返回 Redis 统计（更准确）
	if r.redisLimiter != nil {
		stats, err := r.redisLimiter.GetStats(ctx, key)
		if err == nil {
			return stats, nil
		}
	}
	// 降级到本地统计
	if r.localLimiter != nil {
		return r.localLimiter.GetStats(ctx, key)
	}
	return nil, fmt.Errorf("no limiter available")
}

// ========== 预设配置已在 types.go 中定义 ==========
