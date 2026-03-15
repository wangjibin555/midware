package Auth

import "errors"

// ========== 认证相关错误 ==========

var (
	// ErrInvalidToken Token 无效
	ErrInvalidToken = errors.New("invalid token")

	// ErrExpiredToken Token 已过期
	ErrExpiredToken = errors.New("token expired")

	// ErrTokenNotYetValid Token 还未生效
	ErrTokenNotYetValid = errors.New("token not yet valid")

	// ErrInvalidSignature 签名无效
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrMissingToken 缺少 Token
	ErrMissingToken = errors.New("missing token")

	// ErrRevokedToken Token 已被撤销（在黑名单中）
	ErrRevokedToken = errors.New("token has been revoked")

	// ErrInvalidCredentials 凭证无效（用户名或密码错误）
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidTokenFormat Token 格式无效
	ErrInvalidTokenFormat = errors.New("invalid token format")
)

// ========== 授权相关错误 ==========

var (
	// ErrPermissionDenied 权限不足
	ErrPermissionDenied = errors.New("permission denied")

	// ErrRoleNotFound 角色不存在
	ErrRoleNotFound = errors.New("role not found")

	// ErrPermissionInvalid 权限无效
	ErrPermissionInvalid = errors.New("permission invalid")

	// ErrUserNotAuthorized 用户未授权
	ErrUserNotAuthorized = errors.New("user not authorized")

	// ErrInsufficientPrivileges 权限不足（更具体）
	ErrInsufficientPrivileges = errors.New("insufficient privileges")
)

// ========== 用户相关错误 ==========

var (
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("user not found")

	// ErrUserDisabled 用户已被禁用
	ErrUserDisabled = errors.New("user disabled")

	// ErrUserAlreadyExists 用户已存在
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidUserID 用户ID无效
	ErrInvalidUserID = errors.New("invalid user id")

	// ErrUserDeleted 用户已被删除
	ErrUserDeleted = errors.New("user deleted")

	// ErrInvalidUsername 用户名无效
	ErrInvalidUsername = errors.New("invalid username")

	// ErrInvalidEmail 邮箱无效
	ErrInvalidEmail = errors.New("invalid email")

	// ErrInvalidPassword 密码无效
	ErrInvalidPassword = errors.New("invalid password")

	// ErrWeakPassword 密码强度不足
	ErrWeakPassword = errors.New("weak password")
)

// ========== Session 相关错误 ==========

var (
	// ErrSessionNotFound Session 不存在
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExpired Session 已过期
	ErrSessionExpired = errors.New("session expired")

	// ErrSessionInvalid Session 无效
	ErrSessionInvalid = errors.New("session invalid")

	// ErrSessionCreationFailed Session 创建失败
	ErrSessionCreationFailed = errors.New("session creation failed")
)

// ========== OAuth2 相关错误 ==========

var (
	// ErrOAuth2InvalidCode OAuth2 授权码无效
	ErrOAuth2InvalidCode = errors.New("invalid oauth2 code")

	// ErrOAuth2InvalidState OAuth2 State 无效（CSRF 攻击）
	ErrOAuth2InvalidState = errors.New("invalid oauth2 state")

	// ErrOAuth2TokenExchange OAuth2 Token 交换失败
	ErrOAuth2TokenExchange = errors.New("oauth2 token exchange failed")

	// ErrOAuth2UserInfoFailed 获取 OAuth2 用户信息失败
	ErrOAuth2UserInfoFailed = errors.New("oauth2 user info failed")

	// ErrOAuth2InvalidConfig OAuth2 配置无效
	ErrOAuth2InvalidConfig = errors.New("invalid oauth2 config")

	// ErrOAuth2ProviderNotSupported OAuth2 提供商不支持
	ErrOAuth2ProviderNotSupported = errors.New("oauth2 provider not supported")
)

// ========== 存储相关错误 ==========

var (
	// ErrStorageNotFound 存储中找不到数据
	ErrStorageNotFound = errors.New("storage: key not found")

	// ErrStorageOperation 存储操作失败
	ErrStorageOperation = errors.New("storage: operation failed")

	// ErrStorageConnection 存储连接失败
	ErrStorageConnection = errors.New("storage: connection failed")

	// ErrStorageTimeout 存储超时
	ErrStorageTimeout = errors.New("storage: timeout")

	// ErrStorageInvalidKey 存储键无效
	ErrStorageInvalidKey = errors.New("storage: invalid key")
)

// ========== 缓存相关错误 ==========

var (
	// ErrCacheNotFound 缓存中找不到数据
	ErrCacheNotFound = errors.New("cache: key not found")

	// ErrCacheExpired 缓存已过期
	ErrCacheExpired = errors.New("cache: entry expired")

	// ErrCacheInvalid 缓存无效
	ErrCacheInvalid = errors.New("cache: invalid entry")

	// ErrCacheFull 缓存已满
	ErrCacheFull = errors.New("cache: full")
)

// ========== 配置相关错误 ==========

var (
	// ErrInvalidConfig 配置无效
	ErrInvalidConfig = errors.New("invalid config")

	// ErrMissingConfig 缺少配置
	ErrMissingConfig = errors.New("missing config")

	// ErrInvalidJWTSecret JWT Secret 无效
	ErrInvalidJWTSecret = errors.New("invalid jwt secret")

	// ErrConfigValidationFailed 配置验证失败
	ErrConfigValidationFailed = errors.New("config validation failed")
)

// ========== 数据库相关错误 ==========

var (
	// ErrDatabaseConnection 数据库连接失败
	ErrDatabaseConnection = errors.New("database: connection failed")

	// ErrDatabaseQuery 数据库查询失败
	ErrDatabaseQuery = errors.New("database: query failed")

	// ErrDatabaseTransaction 数据库事务失败
	ErrDatabaseTransaction = errors.New("database: transaction failed")

	// ErrDuplicateEntry 数据库中存在重复条目
	ErrDuplicateEntry = errors.New("database: duplicate entry")

	// ErrForeignKeyViolation 外键约束违反
	ErrForeignKeyViolation = errors.New("database: foreign key violation")
)

// ========== 其他错误 ==========

var (
	// ErrInternalServer 内部服务器错误
	ErrInternalServer = errors.New("internal server error")

	// ErrNotImplemented 功能未实现
	ErrNotImplemented = errors.New("not implemented")

	// ErrInvalidInput 输入无效
	ErrInvalidInput = errors.New("invalid input")

	// ErrOperationFailed 操作失败
	ErrOperationFailed = errors.New("operation failed")

	// ErrTimeout 操作超时
	ErrTimeout = errors.New("timeout")

	// ErrContextCanceled 上下文被取消
	ErrContextCanceled = errors.New("context canceled")
)
