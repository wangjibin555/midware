package RateLimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ========== Redis 分布式限流器 ==========

// RedisLimiter Redis 分布式限流器（使用滑动窗口算法 + Lua 脚本）
type RedisLimiter struct {
	limit     int64
	window    time.Duration
	keyPrefix string
	client    *redis.Client
}

// NewRedisLimiter 创建 Redis 限流器
func NewRedisLimiter(limit int64, window time.Duration, keyPrefix string, client *redis.Client) *RedisLimiter {
	return &RedisLimiter{
		limit:     limit,
		window:    window,
		keyPrefix: keyPrefix,
		client:    client,
	}
}

// Allow 检查是否允许（使用 Lua 脚本保证原子性）
func (l *RedisLimiter) Allow(ctx context.Context, key string) (*Result, error) {
	fullKey := l.keyPrefix + key

	now := time.Now().UnixMilli()
	windowStart := now - l.window.Milliseconds()

	// Lua 脚本：滑动窗口限流
	script := `
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local window_ms = tonumber(ARGV[4])

		-- 删除过期数据
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)

		-- 获取当前计数
		local current = redis.call('ZCARD', key)

		if current < limit then
			-- 允许通过，添加时间戳
			redis.call('ZADD', key, now, now)
			redis.call('PEXPIRE', key, window_ms)
			return {1, current + 1, limit - current - 1}
		else
			-- 拒绝，获取最早的时间戳
			local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
			local reset_at = tonumber(oldest[2]) + window_ms
			return {0, current, 0, reset_at}
		end
	`

	result, err := l.client.Eval(ctx, script, []string{fullKey},
		now, windowStart, l.limit, l.window.Milliseconds()).Result()

	if err != nil {
		return nil, fmt.Errorf("redis eval failed: %w", err)
	}

	// 解析结果
	values, ok := result.([]interface{})
	if !ok || len(values) < 3 {
		return nil, fmt.Errorf("invalid redis response")
	}

	allowed := values[0].(int64) == 1
	current := values[1].(int64)
	remaining := values[2].(int64)

	res := &Result{
		Allowed:     allowed,
		Limit:       l.limit,
		Current:     current,
		Remaining:   remaining,
		RateLimited: !allowed,
	}

	if !allowed && len(values) >= 4 {
		resetAtMs := values[3].(int64)
		res.ResetAt = time.UnixMilli(resetAtMs)
		res.RetryAfter = time.Until(res.ResetAt)
	} else {
		res.ResetAt = time.Now().Add(l.window)
	}

	return res, nil
}

// Take 获取令牌（阻塞版本）
func (l *RedisLimiter) Take(ctx context.Context, key string) error {
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
func (l *RedisLimiter) Reset(ctx context.Context, key string) error {
	fullKey := l.keyPrefix + key
	return l.client.Del(ctx, fullKey).Err()
}

// GetStats 获取统计信息
func (l *RedisLimiter) GetStats(ctx context.Context, key string) (*Stats, error) {
	fullKey := l.keyPrefix + key

	now := time.Now().UnixMilli()
	windowStart := now - l.window.Milliseconds()

	// 清理过期数据
	l.client.ZRemRangeByScore(ctx, fullKey, "0", fmt.Sprintf("%d", windowStart))

	// 获取当前计数
	current, err := l.client.ZCard(ctx, fullKey).Result()
	if err != nil {
		return nil, err
	}

	return &Stats{
		CurrentUsage: current,
		Limit:        l.limit,
		WindowStart:  time.UnixMilli(windowStart),
		WindowEnd:    time.Now(),
	}, nil
}

// ========== 固定窗口算法（更简单，性能更好） ==========

// AllowFixedWindow 固定窗口算法（使用 INCR + EXPIRE）
func (l *RedisLimiter) AllowFixedWindow(ctx context.Context, key string) (*Result, error) {
	fullKey := l.keyPrefix + key

	// Lua 脚本：固定窗口限流
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window_seconds = tonumber(ARGV[2])

		local current = redis.call('GET', key)
		if current == false then
			current = 0
		else
			current = tonumber(current)
		end

		if current < limit then
			local new_val = redis.call('INCR', key)
			if new_val == 1 then
				redis.call('EXPIRE', key, window_seconds)
			end
			local ttl = redis.call('TTL', key)
			return {1, new_val, limit - new_val, ttl}
		else
			local ttl = redis.call('TTL', key)
			return {0, current, 0, ttl}
		end
	`

	result, err := l.client.Eval(ctx, script, []string{fullKey},
		l.limit, int64(l.window.Seconds())).Result()

	if err != nil {
		return nil, fmt.Errorf("redis eval failed: %w", err)
	}

	// 解析结果
	values, ok := result.([]interface{})
	if !ok || len(values) < 4 {
		return nil, fmt.Errorf("invalid redis response")
	}

	allowed := values[0].(int64) == 1
	current := values[1].(int64)
	remaining := values[2].(int64)
	ttl := values[3].(int64)

	return &Result{
		Allowed:     allowed,
		Limit:       l.limit,
		Current:     current,
		Remaining:   remaining,
		ResetAt:     time.Now().Add(time.Duration(ttl) * time.Second),
		RetryAfter:  time.Duration(ttl) * time.Second,
		RateLimited: !allowed,
	}, nil
}

// ========== 令牌桶算法 ==========

// AllowTokenBucket 令牌桶算法
func (l *RedisLimiter) AllowTokenBucket(ctx context.Context, key string, capacity int64, rate float64) (*Result, error) {
	fullKey := l.keyPrefix + key

	// Lua 脚本：令牌桶算法
	script := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		local bucket = redis.call('HMGET', key, 'tokens', 'last_update')
		local tokens = tonumber(bucket[1])
		local last_update = tonumber(bucket[2])

		if tokens == nil then
			tokens = capacity
			last_update = now
		else
			-- 计算新增令牌
			local delta = now - last_update
			local new_tokens = math.min(capacity, tokens + delta * rate / 1000)
			tokens = new_tokens
			last_update = now
		end

		if tokens >= 1 then
			tokens = tokens - 1
			redis.call('HMSET', key, 'tokens', tokens, 'last_update', last_update)
			redis.call('EXPIRE', key, 3600)
			return {1, math.floor(tokens), capacity}
		else
			return {0, 0, capacity}
		end
	`

	res, err := l.client.Eval(ctx, script, []string{fullKey},
		capacity, rate, time.Now().UnixMilli()).Result()

	if err != nil {
		return nil, fmt.Errorf("redis eval failed: %w", err)
	}

	// 解析结果
	values, ok := res.([]interface{})
	if !ok || len(values) < 3 {
		return nil, fmt.Errorf("invalid redis response")
	}

	allowed := values[0].(int64) == 1
	remaining := values[1].(int64)
	bucketCapacity := values[2].(int64)

	return &Result{
		Allowed:     allowed,
		Limit:       bucketCapacity,
		Current:     bucketCapacity - remaining,
		Remaining:   remaining,
		ResetAt:     time.Now().Add(time.Second),
		RetryAfter:  time.Millisecond * 100, // 令牌桶可以快速重试
		RateLimited: !allowed,
	}, nil
}
