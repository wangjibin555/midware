package ErrorHandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

var errRateLimitExceeded = errors.New("rate limit exceeded")

func TestResolverResolveMappedError(t *testing.T) {
	resolver := NewResolver()
	resolver.Register(errRateLimitExceeded, TooManyRequests("rate limit exceeded"))

	appErr := resolver.Resolve(errRateLimitExceeded)
	if appErr.Status != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", appErr.Status)
	}
	if appErr.Code != CodeTooManyRequests {
		t.Fatalf("expected code %s, got %s", CodeTooManyRequests, appErr.Code)
	}
}

func TestResolverResolveContextError(t *testing.T) {
	resolver := NewResolver()
	appErr := resolver.Resolve(context.DeadlineExceeded)

	if appErr.Status != http.StatusGatewayTimeout {
		t.Fatalf("expected status 504, got %d", appErr.Status)
	}
	if appErr.Code != CodeGatewayTimeout {
		t.Fatalf("expected code %s, got %s", CodeGatewayTimeout, appErr.Code)
	}
}

func TestResolverUsesErrorInterfaces(t *testing.T) {
	resolver := NewResolver()
	err := customError{}

	appErr := resolver.Resolve(err)
	if appErr.Status != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", appErr.Status)
	}
	if appErr.Code != "USER_CONFLICT" {
		t.Fatalf("expected code USER_CONFLICT, got %s", appErr.Code)
	}
	if appErr.Message != "user already exists" {
		t.Fatalf("unexpected message: %s", appErr.Message)
	}
	if appErr.Details["field"] != "username" {
		t.Fatalf("unexpected details: %+v", appErr.Details)
	}
}

func TestHTTPWrapWritesMappedErrorResponse(t *testing.T) {
	handler := NewHTTPHandler()
	handler.Register(errRateLimitExceeded, TooManyRequests("rate limit exceeded"))

	req := httptest.NewRequest(http.MethodGet, "/demo", nil)
	req.Header.Set("X-Request-ID", "req-123")
	rec := httptest.NewRecorder()

	httpHandler := handler.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		return errRateLimitExceeded
	})
	httpHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rec.Code)
	}

	var resp HTTPResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != CodeTooManyRequests {
		t.Fatalf("expected code %s, got %s", CodeTooManyRequests, resp.Code)
	}
	if resp.RequestID != "req-123" {
		t.Fatalf("expected request id req-123, got %s", resp.RequestID)
	}
}

func TestHTTPMiddlewareRecoversPanic(t *testing.T) {
	handler := NewHTTPHandler(WithTimestamp(false))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	httpHandler := handler.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	httpHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	var resp HTTPResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != CodeInternalServerError {
		t.Fatalf("expected code %s, got %s", CodeInternalServerError, resp.Code)
	}
	if resp.Message != "internal server error" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

type customError struct{}

func (customError) Error() string {
	return "duplicate record"
}

func (customError) StatusCode() int {
	return http.StatusConflict
}

func (customError) ErrorCode() string {
	return "USER_CONFLICT"
}

func (customError) PublicMessage() string {
	return "user already exists"
}

func (customError) ErrorDetails() map[string]any {
	return map[string]any{"field": "username"}
}
