package Auth

import "time"

// ========== 配置结构 ==========

// Config Auth 系统配置
type Config struct {
	// ===== JWT 配置 =====
	JWTSecret          string        `json:"jwt_secret"`           // JWT 密钥（必须至少32字节）
	JWTIssuer          string        `json:"jwt_issuer"`           // JWT 签发者
	AccessTokenExpire  time.Duration `json:"access_token_expire"`  // Access Token 过期时间（默认15分钟）
	RefreshTokenExpire time.Duration `json:"refresh_token_expire"` // Refresh Token 过期时间（默认7天）

	// ===== Session 配置 =====
	SessionExpire    time.Duration `json:"session_expire"`     // Session 过期时间（默认24小时）
	SessionKeyPrefix string        `json:"session_key_prefix"` // Session Key 前缀（Redis）

	// ===== 缓存配置 =====
	EnableLocalCache bool          `json:"enable_local_cache"` // 是否启用本地缓存（L1）
	LocalCacheSize   int           `json:"local_cache_size"`   // 本地缓存大小（条目数）
	LocalCacheTTL    time.Duration `json:"local_cache_ttl"`    // 本地缓存过期时间
	RedisCacheTTL    time.Duration `json:"redis_cache_ttl"`    // Redis 缓存过期时间（L2）

	// ===== Token 黑名单配置 =====
	EnableBlacklist    bool   `json:"enable_blacklist"`     // 是否启用 Token 黑名单
	BlacklistKeyPrefix string `json:"blacklist_key_prefix"` // 黑名单 Key 前缀（Redis）

	// ===== 安全配置 =====
	EnableCaller     bool `json:"enable_caller"`      // 是否记录调用者信息（文件:行号）
	EnableStackTrace bool `json:"enable_stack_trace"` // 是否记录堆栈信息（Error级别以上）

	// ===== 密码策略配置 =====
	MinPasswordLength int  `json:"min_password_length"` // 最小密码长度（默认8）
	RequireUppercase  bool `json:"require_uppercase"`   // 是否要求大写字母
	RequireLowercase  bool `json:"require_lowercase"`   // 是否要求小写字母
	RequireNumber     bool `json:"require_number"`      // 是否要求数字
	RequireSpecial    bool `json:"require_special"`     // 是否要求特殊字符

	// ===== 其他配置 =====
	MaxLoginAttempts  int           `json:"max_login_attempts"`  // 最大登录尝试次数（防暴力破解）
	LoginLockDuration time.Duration `json:"login_lock_duration"` // 登录锁定时长
}

// ========== 默认配置 ==========

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		// JWT 默认配置
		JWTIssuer:          "auth-service",
		AccessTokenExpire:  15 * time.Minute,   // 15分钟
		RefreshTokenExpire: 7 * 24 * time.Hour, // 7天

		// Session 默认配置
		SessionExpire:    24 * time.Hour, // 24小时
		SessionKeyPrefix: "session:",

		// 缓存默认配置
		EnableLocalCache: true,
		LocalCacheSize:   1000,             // 最多1000条
		LocalCacheTTL:    5 * time.Minute,  // 5分钟
		RedisCacheTTL:    15 * time.Minute, // 15分钟

		// Token 黑名单默认配置
		EnableBlacklist:    true,
		BlacklistKeyPrefix: "blacklist:",

		// 安全默认配置
		EnableCaller:     false,
		EnableStackTrace: false,

		// 密码策略默认配置
		MinPasswordLength: 8,
		RequireUppercase:  false,
		RequireLowercase:  false,
		RequireNumber:     false,
		RequireSpecial:    false,

		// 其他默认配置
		MaxLoginAttempts:  5,                // 5次
		LoginLockDuration: 15 * time.Minute, // 锁定15分钟
	}
}

// ========== 函数式选项 ==========

// Option 配置选项函数类型
type Option func(*Config)

// ===== JWT 选项 =====

// WithJWTSecret 设置 JWT Secret（必须设置）
func WithJWTSecret(secret string) Option {
	return func(c *Config) {
		c.JWTSecret = secret
	}
}

// WithJWTIssuer 设置 JWT Issuer
func WithJWTIssuer(issuer string) Option {
	return func(c *Config) {
		c.JWTIssuer = issuer
	}
}

// WithAccessTokenExpire 设置 Access Token 过期时间
func WithAccessTokenExpire(expire time.Duration) Option {
	return func(c *Config) {
		c.AccessTokenExpire = expire
	}
}

// WithRefreshTokenExpire 设置 Refresh Token 过期时间
func WithRefreshTokenExpire(expire time.Duration) Option {
	return func(c *Config) {
		c.RefreshTokenExpire = expire
	}
}

// ===== Session 选项 =====

// WithSessionExpire 设置 Session 过期时间
func WithSessionExpire(expire time.Duration) Option {
	return func(c *Config) {
		c.SessionExpire = expire
	}
}

// WithSessionKeyPrefix 设置 Session Key 前缀
func WithSessionKeyPrefix(prefix string) Option {
	return func(c *Config) {
		c.SessionKeyPrefix = prefix
	}
}

// ===== 缓存选项 =====

// WithLocalCache 设置本地缓存配置
func WithLocalCache(enable bool, size int, ttl time.Duration) Option {
	return func(c *Config) {
		c.EnableLocalCache = enable
		c.LocalCacheSize = size
		c.LocalCacheTTL = ttl
	}
}

// WithRedisCacheTTL 设置 Redis 缓存过期时间
func WithRedisCacheTTL(ttl time.Duration) Option {
	return func(c *Config) {
		c.RedisCacheTTL = ttl
	}
}

// ===== Token 黑名单选项 =====

// WithBlacklist 设置是否启用 Token 黑名单
func WithBlacklist(enable bool) Option {
	return func(c *Config) {
		c.EnableBlacklist = enable
	}
}

// WithBlacklistKeyPrefix 设置黑名单 Key 前缀
func WithBlacklistKeyPrefix(prefix string) Option {
	return func(c *Config) {
		c.BlacklistKeyPrefix = prefix
	}
}

// ===== 安全选项 =====

// WithCaller 设置是否记录调用者信息
func WithCaller(enable bool) Option {
	return func(c *Config) {
		c.EnableCaller = enable
	}
}

// WithStackTrace 设置是否记录堆栈信息
func WithStackTrace(enable bool) Option {
	return func(c *Config) {
		c.EnableStackTrace = enable
	}
}

// ===== 密码策略选项 =====

// WithPasswordPolicy 设置密码策略
func WithPasswordPolicy(minLength int, requireUpper, requireLower, requireNumber, requireSpecial bool) Option {
	return func(c *Config) {
		c.MinPasswordLength = minLength
		c.RequireUppercase = requireUpper
		c.RequireLowercase = requireLower
		c.RequireNumber = requireNumber
		c.RequireSpecial = requireSpecial
	}
}

// WithMinPasswordLength 设置最小密码长度
func WithMinPasswordLength(length int) Option {
	return func(c *Config) {
		c.MinPasswordLength = length
	}
}

// ===== 防暴力破解选项 =====

// WithLoginAttempts 设置最大登录尝试次数和锁定时长
func WithLoginAttempts(maxAttempts int, lockDuration time.Duration) Option {
	return func(c *Config) {
		c.MaxLoginAttempts = maxAttempts
		c.LoginLockDuration = lockDuration
	}
}

// ========== 配置验证 ==========

// Validate 验证配置是否合法
func (c *Config) Validate() error {
	// 验证 JWT Secret
	if c.JWTSecret == "" {
		return ErrInvalidJWTSecret
	}
	if len(c.JWTSecret) < 32 {
		return ErrInvalidJWTSecret
	}

	// 验证过期时间
	if c.AccessTokenExpire <= 0 {
		return ErrInvalidConfig
	}
	if c.RefreshTokenExpire <= 0 {
		return ErrInvalidConfig
	}
	if c.SessionExpire <= 0 {
		return ErrInvalidConfig
	}

	// 验证缓存配置
	if c.EnableLocalCache {
		if c.LocalCacheSize <= 0 {
			return ErrInvalidConfig
		}
		if c.LocalCacheTTL <= 0 {
			return ErrInvalidConfig
		}
	}

	// 验证密码策略
	if c.MinPasswordLength < 6 {
		return ErrInvalidConfig
	}

	return nil
}

// ========== 预设配置 ==========

// ProductionConfig 生产环境配置（安全性优先）
func ProductionConfig(jwtSecret string) *Config {
	config := DefaultConfig()
	config.JWTSecret = jwtSecret
	config.AccessTokenExpire = 15 * time.Minute // 短期
	config.RefreshTokenExpire = 7 * 24 * time.Hour
	config.EnableBlacklist = true  // 启用黑名单
	config.EnableLocalCache = true // 启用缓存
	config.MinPasswordLength = 12  // 强密码
	config.RequireUppercase = true
	config.RequireLowercase = true
	config.RequireNumber = true
	config.RequireSpecial = true
	config.MaxLoginAttempts = 5
	config.LoginLockDuration = 30 * time.Minute // 锁定30分钟
	return config
}

// DevelopmentConfig 开发环境配置（便利性优先）
func DevelopmentConfig(jwtSecret string) *Config {
	config := DefaultConfig()
	config.JWTSecret = jwtSecret
	config.AccessTokenExpire = 24 * time.Hour // 长期，方便开发
	config.RefreshTokenExpire = 30 * 24 * time.Hour
	config.EnableBlacklist = false // 不启用黑名单
	config.EnableLocalCache = true
	config.MinPasswordLength = 6 // 弱密码策略
	config.RequireUppercase = false
	config.RequireLowercase = false
	config.RequireNumber = false
	config.RequireSpecial = false
	config.MaxLoginAttempts = 100 // 宽松限制
	config.LoginLockDuration = 1 * time.Minute
	return config
}

// TestConfig 测试环境配置（速度优先）
func TestConfig() *Config {
	config := DefaultConfig()
	config.JWTSecret = "test-secret-32-bytes-long-key!!" // 测试密钥
	config.AccessTokenExpire = 1 * time.Hour
	config.RefreshTokenExpire = 24 * time.Hour
	config.EnableBlacklist = false  // 关闭黑名单
	config.EnableLocalCache = false // 关闭缓存（简化测试）
	config.MinPasswordLength = 4    // 最简单密码
	config.MaxLoginAttempts = 1000
	return config
}

// ========== 辅助方法 ==========

// Clone 克隆配置（深拷贝）
func (c *Config) Clone() *Config {
	newConfig := *c
	return &newConfig
}

// String 返回配置摘要（隐藏敏感信息）
func (c *Config) String() string {
	return "Auth Config: " +
		"AccessTokenExpire=" + c.AccessTokenExpire.String() + ", " +
		"RefreshTokenExpire=" + c.RefreshTokenExpire.String() + ", " +
		"SessionExpire=" + c.SessionExpire.String() + ", " +
		"EnableBlacklist=" + boolToString(c.EnableBlacklist) + ", " +
		"EnableLocalCache=" + boolToString(c.EnableLocalCache)
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
