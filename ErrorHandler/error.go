package ErrorHandler

import (
	"errors"
	"net/http"
)

const (
	CodeBadRequest          = "BAD_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeTooManyRequests     = "TOO_MANY_REQUESTS"
	CodeRequestTimeout      = "REQUEST_TIMEOUT"
	CodeGatewayTimeout      = "GATEWAY_TIMEOUT"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	CodeNotImplemented      = "NOT_IMPLEMENTED"
)

// AppError 是统一错误模型，可直接用于业务层和 HTTP 层。
type AppError struct {
	Status  int
	Code    string
	Message string
	Cause   error
	Details map[string]any
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause == nil || e.Cause.Error() == "" || e.Cause.Error() == e.Message {
		return e.Message
	}
	return e.Message + ": " + e.Cause.Error()
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *AppError) Clone() *AppError {
	if e == nil {
		return nil
	}

	clone := *e
	if len(e.Details) > 0 {
		clone.Details = make(map[string]any, len(e.Details))
		for k, v := range e.Details {
			clone.Details[k] = v
		}
	}
	return &clone
}

func (e *AppError) WithCause(err error) *AppError {
	if e == nil {
		return nil
	}
	clone := e.Clone()
	clone.Cause = err
	return clone
}

func (e *AppError) WithDetails(details map[string]any) *AppError {
	if e == nil {
		return nil
	}
	clone := e.Clone()
	if len(details) == 0 {
		return clone
	}
	if clone.Details == nil {
		clone.Details = make(map[string]any, len(details))
	}
	for k, v := range details {
		clone.Details[k] = v
	}
	return clone
}

func (e *AppError) WithDetail(key string, value any) *AppError {
	return e.WithDetails(map[string]any{key: value})
}

func (e *AppError) Is(target error) bool {
	other, ok := target.(*AppError)
	if !ok {
		return false
	}
	if other.Code != "" && e.Code != other.Code {
		return false
	}
	if other.Status != 0 && e.Status != other.Status {
		return false
	}
	return true
}

func New(status int, code, message string) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, status int, code, message string) *AppError {
	if err == nil {
		return New(status, code, message)
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		clone := appErr.Clone()
		if clone.Status == 0 {
			clone.Status = status
		}
		if clone.Code == "" {
			clone.Code = code
		}
		if clone.Message == "" {
			clone.Message = message
		}
		if clone.Cause == nil {
			clone.Cause = err
		}
		return clone
	}

	return New(status, code, message).WithCause(err)
}

func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, CodeBadRequest, message)
}

func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, CodeForbidden, message)
}

func NotFound(message string) *AppError {
	return New(http.StatusNotFound, CodeNotFound, message)
}

func Conflict(message string) *AppError {
	return New(http.StatusConflict, CodeConflict, message)
}

func TooManyRequests(message string) *AppError {
	return New(http.StatusTooManyRequests, CodeTooManyRequests, message)
}

func RequestTimeout(message string) *AppError {
	return New(http.StatusRequestTimeout, CodeRequestTimeout, message)
}

func GatewayTimeout(message string) *AppError {
	return New(http.StatusGatewayTimeout, CodeGatewayTimeout, message)
}

func Internal(message string) *AppError {
	return New(http.StatusInternalServerError, CodeInternalServerError, message)
}

func ServiceUnavailable(message string) *AppError {
	return New(http.StatusServiceUnavailable, CodeServiceUnavailable, message)
}

func NotImplemented(message string) *AppError {
	return New(http.StatusNotImplemented, CodeNotImplemented, message)
}
