package RateLimit

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// ========== HTTP 中间件 ==========

// Middleware HTTP 限流中间件
func (r *RateLimiter) Middleware(keyFunc KeyFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// 1. 生成限流键
			key := keyFunc(req)
			if key == "" {
				// 无法生成键，直接通过
				next.ServeHTTP(w, req)
				return
			}

			// 2. 检查限流
			result, err := r.Allow(req.Context(), key)
			if err != nil {
				// 限流检查失败，记录错误但允许通过（可配置）
				http.Error(w, "Rate limit check failed", http.StatusInternalServerError)
				return
			}

			// 3. 设置响应头
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
			w.Header().Set("X-RateLimit-Reset", result.ResetAt.Format(http.TimeFormat))
			w.Header().Set("X-RateLimit-Used", fmt.Sprintf("%d", result.Current))

			// 4. 检查是否被限流
			if !result.Allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(result.RetryAfter.Seconds())))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// 5. 通过，继续处理
			next.ServeHTTP(w, req)
		})
	}
}

// ========== 键生成函数 ==========

// KeyFunc 键生成函数类型
type KeyFunc func(*http.Request) string

// KeyByIP 按 IP 限流
func KeyByIP(r *http.Request) string {
	ip := getClientIP(r)
	return "ip:" + ip
}

// KeyByUser 按用户 ID 限流（从 Context 获取）
func KeyByUser(r *http.Request) string {
	userID := getUserID(r.Context())
	if userID == "" {
		// 未登录用户使用 IP
		return KeyByIP(r)
	}
	return "user:" + userID
}

// KeyByEndpoint 按接口路径限流
func KeyByEndpoint(r *http.Request) string {
	return "endpoint:" + r.URL.Path
}

// KeyByIPAndEndpoint 按 IP + 接口限流
func KeyByIPAndEndpoint(r *http.Request) string {
	ip := getClientIP(r)
	return Combined(ip, r.URL.Path)
}

// KeyByUserAndEndpoint 按用户 + 接口限流
func KeyByUserAndEndpoint(r *http.Request) string {
	userID := getUserID(r.Context())
	if userID == "" {
		return KeyByIPAndEndpoint(r)
	}
	return Combined(userID, r.URL.Path)
}

// KeyByAPIKey 按 API Key 限流
func KeyByAPIKey(r *http.Request) string {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		apiKey = r.URL.Query().Get("api_key")
	}
	if apiKey == "" {
		// 无 API Key，使用 IP
		return KeyByIP(r)
	}
	return "apikey:" + apiKey
}

// ========== 辅助函数 ==========

// getClientIP 获取客户端真实 IP
func getClientIP(r *http.Request) string {
	// 尝试从 X-Forwarded-For 获取
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从 X-Real-IP 获取
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用 RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// getUserID 从 Context 获取用户 ID（需要配合认证中间件使用）
func getUserID(ctx context.Context) string {
	// 尝试从 Context 获取 user_id
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	// 尝试从 Context 获取 claims
	if claims, ok := ctx.Value("claims").(interface{ GetUserID() string }); ok {
		return claims.GetUserID()
	}
	return ""
}

// ========== 快捷中间件 ==========

// MiddlewareByIP 按 IP 限流的中间件
func MiddlewareByIP(limiter *RateLimiter) func(http.Handler) http.Handler {
	return limiter.Middleware(KeyByIP)
}

// MiddlewareByUser 按用户限流的中间件
func MiddlewareByUser(limiter *RateLimiter) func(http.Handler) http.Handler {
	return limiter.Middleware(KeyByUser)
}

// MiddlewareByEndpoint 按接口限流的中间件
func MiddlewareByEndpoint(limiter *RateLimiter) func(http.Handler) http.Handler {
	return limiter.Middleware(KeyByEndpoint)
}

// MiddlewareByAPIKey 按 API Key 限流的中间件
func MiddlewareByAPIKey(limiter *RateLimiter) func(http.Handler) http.Handler {
	return limiter.Middleware(KeyByAPIKey)
}

// ========== 自定义响应 ==========

// MiddlewareWithCustomResponse 自定义响应的中间件
func (r *RateLimiter) MiddlewareWithCustomResponse(
	keyFunc KeyFunc,
	onRateLimited func(w http.ResponseWriter, req *http.Request, result *Result),
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			key := keyFunc(req)
			if key == "" {
				next.ServeHTTP(w, req)
				return
			}

			result, err := r.Allow(req.Context(), key)
			if err != nil {
				http.Error(w, "Rate limit check failed", http.StatusInternalServerError)
				return
			}

			// 设置响应头
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
			w.Header().Set("X-RateLimit-Reset", result.ResetAt.Format(http.TimeFormat))

			if !result.Allowed {
				// 调用自定义处理函数
				onRateLimited(w, req, result)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}
