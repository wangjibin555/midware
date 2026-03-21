package RateLimit

import "net/http"

type rateLimitError struct {
	status  int
	code    string
	message string
	cause   error
	details map[string]any
}

func (e *rateLimitError) Error() string {
	if e == nil {
		return ""
	}
	if e.cause == nil || e.cause.Error() == "" || e.cause.Error() == e.message {
		return e.message
	}
	return e.message + ": " + e.cause.Error()
}

func (e *rateLimitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *rateLimitError) StatusCode() int {
	return e.status
}

func (e *rateLimitError) ErrorCode() string {
	return e.code
}

func (e *rateLimitError) PublicMessage() string {
	return e.message
}

func (e *rateLimitError) ErrorDetails() map[string]any {
	if len(e.details) == 0 {
		return nil
	}
	dst := make(map[string]any, len(e.details))
	for k, v := range e.details {
		dst[k] = v
	}
	return dst
}

func (e *rateLimitError) WithCause(err error) *rateLimitError {
	if e == nil {
		return nil
	}
	clone := *e
	clone.cause = err
	return &clone
}

func (e *rateLimitError) WithDetails(details map[string]any) *rateLimitError {
	if e == nil {
		return nil
	}
	clone := *e
	if len(details) == 0 {
		return &clone
	}
	clone.details = make(map[string]any, len(details))
	for k, v := range details {
		clone.details[k] = v
	}
	return &clone
}

func newRateLimitError(status int, code, message string) *rateLimitError {
	return &rateLimitError{
		status:  status,
		code:    code,
		message: message,
	}
}

// ========== 限流相关错误 ==========

var (
	ErrRateLimitExceeded  = newRateLimitError(http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "rate limit exceeded")
	ErrInvalidKey         = newRateLimitError(http.StatusBadRequest, "RATE_LIMIT_INVALID_KEY", "invalid rate limit key")
	ErrInvalidConfig      = newRateLimitError(http.StatusBadRequest, "RATE_LIMIT_INVALID_CONFIG", "invalid rate limit config")
	ErrStorageUnavailable = newRateLimitError(http.StatusServiceUnavailable, "RATE_LIMIT_STORAGE_UNAVAILABLE", "storage unavailable")
	ErrNotImplemented     = newRateLimitError(http.StatusNotImplemented, "RATE_LIMIT_NOT_IMPLEMENTED", "not implemented")
)

// ========== 存储相关错误 ==========

var (
	ErrRedisConnectionFailed = newRateLimitError(http.StatusServiceUnavailable, "RATE_LIMIT_REDIS_CONNECTION_FAILED", "redis connection failed")
	ErrRedisOperationFailed  = newRateLimitError(http.StatusInternalServerError, "RATE_LIMIT_REDIS_OPERATION_FAILED", "redis operation failed")
	ErrCacheMiss             = newRateLimitError(http.StatusNotFound, "RATE_LIMIT_CACHE_MISS", "cache miss")
)

// ========== 算法相关错误 ==========

var (
	ErrUnsupportedAlgorithm = newRateLimitError(http.StatusBadRequest, "RATE_LIMIT_UNSUPPORTED_ALGORITHM", "unsupported algorithm")
	ErrInvalidWindow        = newRateLimitError(http.StatusBadRequest, "RATE_LIMIT_INVALID_WINDOW", "invalid time window")
	ErrInvalidLimit         = newRateLimitError(http.StatusBadRequest, "RATE_LIMIT_INVALID_LIMIT", "invalid limit")
)

func NewRateLimitExceededError(result *Result) error {
	if result == nil {
		return ErrRateLimitExceeded
	}
	return ErrRateLimitExceeded.WithDetails(map[string]any{
		"limit":       result.Limit,
		"current":     result.Current,
		"remaining":   result.Remaining,
		"retry_after": int(result.RetryAfter.Seconds()),
		"reset_at":    result.ResetAt.Format(http.TimeFormat),
	})
}
