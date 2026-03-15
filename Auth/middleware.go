package Auth

import (
	"context"
	"net/http"
)

// ========== Context Key ==========

type contextKey string

const (
	// ClaimsContextKey 用于在 Context 中存储 Claims
	ClaimsContextKey contextKey = "auth_claims"
	// UserContextKey 用于在 Context 中存储 User
	UserContextKey contextKey = "auth_user"
)

// ========== 中间件 ==========

// Middleware HTTP 认证中间件
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 从 Header 提取 Token
		authHeader := r.Header.Get("Authorization")
		token, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 2. 验证 Token
		claims, err := a.Verify(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 3. 将 Claims 存入 Context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth 要求认证（快捷方法）
func RequireAuth(a *Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return a.Middleware(next)
	}
}

// RequirePermission 要求特定权限
func RequirePermission(a *Auth, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil || !claims.HasPermission(permission) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole 要求特定角色
func RequireRole(a *Auth, role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil || !claims.HasRole(role) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ========== Context 辅助方法 ==========

// GetClaims 从 Context 获取 Claims
func GetClaims(ctx context.Context) *Claims {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// GetUserID 从 Context 获取用户ID
func GetUserID(ctx context.Context) string {
	claims := GetClaims(ctx)
	if claims == nil {
		return ""
	}
	return claims.UserID
}

// GetUsername 从 Context 获取用户名
func GetUsername(ctx context.Context) string {
	claims := GetClaims(ctx)
	if claims == nil {
		return ""
	}
	return claims.Username
}

// GetUser 从 Context 获取完整用户信息（如果存储了）
func GetUser(ctx context.Context) *User {
	user, ok := ctx.Value(UserContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

// WithClaims 将 Claims 存入 Context
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, ClaimsContextKey, claims)
}

// WithUser 将 User 存入 Context
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}
