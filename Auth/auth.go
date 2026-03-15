package Auth

import (
	"time"

	"github.com/wangjibin555/midware/Auth/crypto"
)

// ========== Auth 核心结构 ==========

// Auth 认证授权管理器
type Auth struct {
	config     *Config
	userStore  UserStore
	tokenStore TokenStore
}

// ========== 初始化 ==========

// New 创建 Auth 实例
func New(config *Config, opts ...Option) (*Auth, error) {
	// 应用选项
	for _, opt := range opts {
		opt(config)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Auth{
		config:     config,
		userStore:  &NoopUserStore{},  // 默认空实现
		tokenStore: &NoopTokenStore{}, // 默认空实现
	}, nil
}

// ========== 设置存储实现 ==========

// SetUserStore 设置用户存储实现
func (a *Auth) SetUserStore(store UserStore) {
	a.userStore = store
}

// SetTokenStore 设置 Token 存储实现（黑名单）
func (a *Auth) SetTokenStore(store TokenStore) {
	a.tokenStore = store
}

// ========== 核心功能：登录 ==========

// Login 用户登录
func (a *Auth) Login(username, password string) (*TokenPair, error) {
	// 1. 验证凭证
	user, err := a.userStore.ValidateCredentials(username, password)
	if err != nil {
		return nil, err
	}

	// 2. 检查用户状态
	if !user.IsActive() {
		if user.IsDisabled() {
			return nil, ErrUserDisabled
		}
		if user.IsDeleted() {
			return nil, ErrUserDeleted
		}
		return nil, ErrUserNotAuthorized
	}

	// 3. 生成 Token
	return a.GenerateTokenPair(user)
}

// ========== 核心功能：Token 生成 ==========

// GenerateTokenPair 生成 Token 对（Access + Refresh）
func (a *Auth) GenerateTokenPair(user *User) (*TokenPair, error) {
	// 1. 创建 Claims
	claims := user.ToClaims()
	claims.Issuer = a.config.JWTIssuer

	// 2. 生成 Access Token
	accessExpire := GetTokenExpiresAt(a.config.AccessTokenExpire)
	accessToken, err := GenerateJWT(claims, a.config.JWTSecret, accessExpire)
	if err != nil {
		return nil, err
	}

	// 3. 生成 Refresh Token
	refreshExpire := GetTokenExpiresAt(a.config.RefreshTokenExpire)
	refreshClaims := &Claims{
		UserID:    user.ID,
		Username:  user.Username,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: refreshExpire.Unix(),
		Issuer:    a.config.JWTIssuer,
	}
	refreshToken, err := GenerateJWT(refreshClaims, a.config.JWTSecret, refreshExpire)
	if err != nil {
		return nil, err
	}

	// 4. 返回 Token 对
	return NewTokenPair(
		accessToken,
		refreshToken,
		int64(a.config.AccessTokenExpire.Seconds()),
	), nil
}

// GenerateAccessToken 生成单个 Access Token
func (a *Auth) GenerateAccessToken(user *User) (string, error) {
	claims := user.ToClaims()
	claims.Issuer = a.config.JWTIssuer
	expireAt := GetTokenExpiresAt(a.config.AccessTokenExpire)
	return GenerateJWT(claims, a.config.JWTSecret, expireAt)
}

// ========== 核心功能：Token 验证 ==========

// Verify 验证 Access Token
func (a *Auth) Verify(token string) (*Claims, error) {
	// 1. 验证 JWT
	claims, err := VerifyJWT(token, a.config.JWTSecret)
	if err != nil {
		return nil, err
	}

	// 2. 检查黑名单（如果启用）
	if a.config.EnableBlacklist {
		inBlacklist, err := a.tokenStore.IsInBlacklist(token)
		if err != nil {
			// 黑名单查询失败，为了安全起见，拒绝访问
			return nil, ErrStorageOperation
		}
		if inBlacklist {
			return nil, ErrRevokedToken
		}
	}

	return claims, nil
}

// VerifyAndGetUser 验证 Token 并获取完整用户信息
func (a *Auth) VerifyAndGetUser(token string) (*User, error) {
	claims, err := a.Verify(token)
	if err != nil {
		return nil, err
	}
	return a.userStore.GetByID(claims.UserID)
}

// ========== 核心功能：刷新 Token ==========

// Refresh 使用 Refresh Token 刷新 Access Token
func (a *Auth) Refresh(refreshToken string) (*TokenPair, error) {
	// 1. 验证 Refresh Token
	claims, err := VerifyJWT(refreshToken, a.config.JWTSecret)
	if err != nil {
		return nil, err
	}

	// 2. 获取用户信息
	user, err := a.userStore.GetByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// 3. 检查用户状态
	if !user.IsActive() {
		return nil, ErrUserNotAuthorized
	}

	// 4. 生成新的 Token 对
	return a.GenerateTokenPair(user)
}

// ========== 核心功能：登出 ==========

// Logout 登出（将 Token 加入黑名单）
func (a *Auth) Logout(token string) error {
	if !a.config.EnableBlacklist {
		return nil // 未启用黑名单，无需操作
	}

	// 解析 Token 获取过期时间
	claims, err := VerifyJWT(token, a.config.JWTSecret)
	if err != nil {
		// Token 已经无效，无需加入黑名单
		return nil
	}

	// 将 Token 加入黑名单
	expireAt := time.Unix(claims.ExpiresAt, 0)
	return a.tokenStore.AddToBlacklist(token, expireAt)
}

// ========== 权限验证 ==========

// CheckPermission 检查用户是否有指定权限
func (a *Auth) CheckPermission(claims *Claims, permission string) bool {
	return claims.HasPermission(permission)
}

// CheckRole 检查用户是否有指定角色
func (a *Auth) CheckRole(claims *Claims, role string) bool {
	return claims.HasRole(role)
}

// RequirePermission 要求指定权限（不满足返回错误）
func (a *Auth) RequirePermission(claims *Claims, permission string) error {
	if !claims.HasPermission(permission) {
		return ErrPermissionDenied
	}
	return nil
}

// RequireRole 要求指定角色（不满足返回错误）
func (a *Auth) RequireRole(claims *Claims, role string) error {
	if !claims.HasRole(role) {
		return ErrPermissionDenied
	}
	return nil
}

// ========== 密码相关 ==========

// HashPassword 加密密码（供应用层使用）
func (a *Auth) HashPassword(password string) (string, error) {
	// 验证密码强度
	if err := crypto.ValidatePasswordStrength(
		password,
		a.config.MinPasswordLength,
		a.config.RequireUppercase,
		a.config.RequireLowercase,
		a.config.RequireNumber,
		a.config.RequireSpecial,
	); err != nil {
		return "", ErrWeakPassword
	}

	return crypto.HashPassword(password)
}

// VerifyPassword 验证密码（供应用层使用）
func (a *Auth) VerifyPassword(password, hash string) (bool, error) {
	return crypto.VerifyPassword(password, hash)
}

// ========== 辅助方法 ==========

// GetConfig 获取配置
func (a *Auth) GetConfig() *Config {
	return a.config
}

// ExtractTokenFromHeader 从 Authorization Header 提取 Token
// 格式：Bearer <token>
func ExtractTokenFromHeader(authHeader string) (string, error) {
	const prefix = "Bearer "
	if len(authHeader) < len(prefix) {
		return "", ErrMissingToken
	}
	if authHeader[:len(prefix)] != prefix {
		return "", ErrInvalidTokenFormat
	}
	return authHeader[len(prefix):], nil
}
