package Auth

import "net/http"

type authError struct {
	status  int
	code    string
	message string
	cause   error
}

func (e *authError) Error() string {
	if e == nil {
		return ""
	}
	if e.cause == nil || e.cause.Error() == "" || e.cause.Error() == e.message {
		return e.message
	}
	return e.message + ": " + e.cause.Error()
}

func (e *authError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *authError) StatusCode() int {
	return e.status
}

func (e *authError) ErrorCode() string {
	return e.code
}

func (e *authError) PublicMessage() string {
	return e.message
}

func (e *authError) WithCause(err error) *authError {
	if e == nil {
		return nil
	}
	clone := *e
	clone.cause = err
	return &clone
}

func newAuthError(status int, code, message string) *authError {
	return &authError{
		status:  status,
		code:    code,
		message: message,
	}
}

// ========== 认证相关错误 ==========

var (
	ErrInvalidToken       = newAuthError(http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid token")
	ErrExpiredToken       = newAuthError(http.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "token expired")
	ErrTokenNotYetValid   = newAuthError(http.StatusUnauthorized, "AUTH_TOKEN_NOT_YET_VALID", "token not yet valid")
	ErrInvalidSignature   = newAuthError(http.StatusUnauthorized, "AUTH_INVALID_SIGNATURE", "invalid signature")
	ErrMissingToken       = newAuthError(http.StatusUnauthorized, "AUTH_MISSING_TOKEN", "missing token")
	ErrRevokedToken       = newAuthError(http.StatusUnauthorized, "AUTH_REVOKED_TOKEN", "token has been revoked")
	ErrInvalidCredentials = newAuthError(http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "invalid credentials")
	ErrInvalidTokenFormat = newAuthError(http.StatusUnauthorized, "AUTH_INVALID_TOKEN_FORMAT", "invalid token format")
)

// ========== 授权相关错误 ==========

var (
	ErrPermissionDenied       = newAuthError(http.StatusForbidden, "AUTH_PERMISSION_DENIED", "permission denied")
	ErrRoleNotFound           = newAuthError(http.StatusNotFound, "AUTH_ROLE_NOT_FOUND", "role not found")
	ErrPermissionInvalid      = newAuthError(http.StatusBadRequest, "AUTH_PERMISSION_INVALID", "permission invalid")
	ErrUserNotAuthorized      = newAuthError(http.StatusForbidden, "AUTH_USER_NOT_AUTHORIZED", "user not authorized")
	ErrInsufficientPrivileges = newAuthError(http.StatusForbidden, "AUTH_INSUFFICIENT_PRIVILEGES", "insufficient privileges")
)

// ========== 用户相关错误 ==========

var (
	ErrUserNotFound      = newAuthError(http.StatusNotFound, "AUTH_USER_NOT_FOUND", "user not found")
	ErrUserDisabled      = newAuthError(http.StatusForbidden, "AUTH_USER_DISABLED", "user disabled")
	ErrUserAlreadyExists = newAuthError(http.StatusConflict, "AUTH_USER_ALREADY_EXISTS", "user already exists")
	ErrInvalidUserID     = newAuthError(http.StatusBadRequest, "AUTH_INVALID_USER_ID", "invalid user id")
	ErrUserDeleted       = newAuthError(http.StatusForbidden, "AUTH_USER_DELETED", "user deleted")
	ErrInvalidUsername   = newAuthError(http.StatusBadRequest, "AUTH_INVALID_USERNAME", "invalid username")
	ErrInvalidEmail      = newAuthError(http.StatusBadRequest, "AUTH_INVALID_EMAIL", "invalid email")
	ErrInvalidPassword   = newAuthError(http.StatusBadRequest, "AUTH_INVALID_PASSWORD", "invalid password")
	ErrWeakPassword      = newAuthError(http.StatusBadRequest, "AUTH_WEAK_PASSWORD", "weak password")
)

// ========== Session 相关错误 ==========

var (
	ErrSessionNotFound       = newAuthError(http.StatusNotFound, "AUTH_SESSION_NOT_FOUND", "session not found")
	ErrSessionExpired        = newAuthError(http.StatusUnauthorized, "AUTH_SESSION_EXPIRED", "session expired")
	ErrSessionInvalid        = newAuthError(http.StatusUnauthorized, "AUTH_SESSION_INVALID", "session invalid")
	ErrSessionCreationFailed = newAuthError(http.StatusInternalServerError, "AUTH_SESSION_CREATION_FAILED", "session creation failed")
)

// ========== OAuth2 相关错误 ==========

var (
	ErrOAuth2InvalidCode          = newAuthError(http.StatusBadRequest, "AUTH_OAUTH2_INVALID_CODE", "invalid oauth2 code")
	ErrOAuth2InvalidState         = newAuthError(http.StatusForbidden, "AUTH_OAUTH2_INVALID_STATE", "invalid oauth2 state")
	ErrOAuth2TokenExchange        = newAuthError(http.StatusBadGateway, "AUTH_OAUTH2_TOKEN_EXCHANGE_FAILED", "oauth2 token exchange failed")
	ErrOAuth2UserInfoFailed       = newAuthError(http.StatusBadGateway, "AUTH_OAUTH2_USER_INFO_FAILED", "oauth2 user info failed")
	ErrOAuth2InvalidConfig        = newAuthError(http.StatusInternalServerError, "AUTH_OAUTH2_INVALID_CONFIG", "invalid oauth2 config")
	ErrOAuth2ProviderNotSupported = newAuthError(http.StatusBadRequest, "AUTH_OAUTH2_PROVIDER_NOT_SUPPORTED", "oauth2 provider not supported")
)

// ========== 存储相关错误 ==========

var (
	ErrStorageNotFound   = newAuthError(http.StatusNotFound, "AUTH_STORAGE_NOT_FOUND", "storage: key not found")
	ErrStorageOperation  = newAuthError(http.StatusInternalServerError, "AUTH_STORAGE_OPERATION_FAILED", "storage: operation failed")
	ErrStorageConnection = newAuthError(http.StatusServiceUnavailable, "AUTH_STORAGE_CONNECTION_FAILED", "storage: connection failed")
	ErrStorageTimeout    = newAuthError(http.StatusGatewayTimeout, "AUTH_STORAGE_TIMEOUT", "storage: timeout")
	ErrStorageInvalidKey = newAuthError(http.StatusBadRequest, "AUTH_STORAGE_INVALID_KEY", "storage: invalid key")
)

// ========== 缓存相关错误 ==========

var (
	ErrCacheNotFound = newAuthError(http.StatusNotFound, "AUTH_CACHE_NOT_FOUND", "cache: key not found")
	ErrCacheExpired  = newAuthError(http.StatusUnauthorized, "AUTH_CACHE_EXPIRED", "cache: entry expired")
	ErrCacheInvalid  = newAuthError(http.StatusInternalServerError, "AUTH_CACHE_INVALID", "cache: invalid entry")
	ErrCacheFull     = newAuthError(http.StatusServiceUnavailable, "AUTH_CACHE_FULL", "cache: full")
)

// ========== 配置相关错误 ==========

var (
	ErrInvalidConfig          = newAuthError(http.StatusBadRequest, "AUTH_INVALID_CONFIG", "invalid config")
	ErrMissingConfig          = newAuthError(http.StatusBadRequest, "AUTH_MISSING_CONFIG", "missing config")
	ErrInvalidJWTSecret       = newAuthError(http.StatusBadRequest, "AUTH_INVALID_JWT_SECRET", "invalid jwt secret")
	ErrConfigValidationFailed = newAuthError(http.StatusBadRequest, "AUTH_CONFIG_VALIDATION_FAILED", "config validation failed")
)

// ========== 数据库相关错误 ==========

var (
	ErrDatabaseConnection  = newAuthError(http.StatusServiceUnavailable, "AUTH_DATABASE_CONNECTION_FAILED", "database: connection failed")
	ErrDatabaseQuery       = newAuthError(http.StatusInternalServerError, "AUTH_DATABASE_QUERY_FAILED", "database: query failed")
	ErrDatabaseTransaction = newAuthError(http.StatusInternalServerError, "AUTH_DATABASE_TRANSACTION_FAILED", "database: transaction failed")
	ErrDuplicateEntry      = newAuthError(http.StatusConflict, "AUTH_DUPLICATE_ENTRY", "database: duplicate entry")
	ErrForeignKeyViolation = newAuthError(http.StatusConflict, "AUTH_FOREIGN_KEY_VIOLATION", "database: foreign key violation")
)

// ========== 其他错误 ==========

var (
	ErrInternalServer  = newAuthError(http.StatusInternalServerError, "AUTH_INTERNAL_SERVER_ERROR", "internal server error")
	ErrNotImplemented  = newAuthError(http.StatusNotImplemented, "AUTH_NOT_IMPLEMENTED", "not implemented")
	ErrInvalidInput    = newAuthError(http.StatusBadRequest, "AUTH_INVALID_INPUT", "invalid input")
	ErrOperationFailed = newAuthError(http.StatusInternalServerError, "AUTH_OPERATION_FAILED", "operation failed")
	ErrTimeout         = newAuthError(http.StatusGatewayTimeout, "AUTH_TIMEOUT", "timeout")
	ErrContextCanceled = newAuthError(http.StatusRequestTimeout, "AUTH_CONTEXT_CANCELED", "context canceled")
)
