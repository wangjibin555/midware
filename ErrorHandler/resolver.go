package ErrorHandler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type StatusCoder interface {
	StatusCode() int
}

type CodeCarrier interface {
	ErrorCode() string
}

type PublicMessageCarrier interface {
	PublicMessage() string
}

type DetailsCarrier interface {
	ErrorDetails() map[string]any
}

type Mapping struct {
	Match func(error) bool
	Map   func(error) *AppError
}

type Resolver struct {
	internalMessage string
	mappings        []Mapping
}

type ResolverOption func(*Resolver)

func NewResolver(opts ...ResolverOption) *Resolver {
	r := &Resolver{
		internalMessage: "internal server error",
	}
	r.RegisterDefaultMappings()
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func WithResolverInternalMessage(message string) ResolverOption {
	return func(r *Resolver) {
		if message != "" {
			r.internalMessage = message
		}
	}
}

func WithResolverMappings(mappings ...Mapping) ResolverOption {
	return func(r *Resolver) {
		r.mappings = append(r.mappings, mappings...)
	}
}

func (r *Resolver) RegisterMapping(mapping Mapping) {
	if mapping.Match == nil || mapping.Map == nil {
		return
	}
	r.mappings = append(r.mappings, mapping)
}

func (r *Resolver) Register(target error, appErr *AppError) {
	if target == nil || appErr == nil {
		return
	}
	r.RegisterMapping(Mapping{
		Match: func(err error) bool {
			return errors.Is(err, target)
		},
		Map: func(err error) *AppError {
			return appErr.WithCause(err)
		},
	})
}

func (r *Resolver) RegisterDefaultMappings() {
	r.RegisterMapping(Mapping{
		Match: func(err error) bool { return errors.Is(err, context.Canceled) },
		Map: func(err error) *AppError {
			return RequestTimeout("request canceled").WithCause(err)
		},
	})
	r.RegisterMapping(Mapping{
		Match: func(err error) bool { return errors.Is(err, context.DeadlineExceeded) },
		Map: func(err error) *AppError {
			return GatewayTimeout("request timed out").WithCause(err)
		},
	})
}

func (r *Resolver) Resolve(err error) *AppError {
	if err == nil {
		return nil
	}

	for i := len(r.mappings) - 1; i >= 0; i-- {
		mapping := r.mappings[i]
		if mapping.Match(err) {
			mapped := mapping.Map(err)
			if mapped != nil {
				return sanitizeAppError(mapped, r.internalMessage)
			}
		}
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return sanitizeAppError(appErr, r.internalMessage)
	}

	resolved := &AppError{
		Status:  http.StatusInternalServerError,
		Code:    CodeInternalServerError,
		Message: r.internalMessage,
		Cause:   err,
	}

	var statusCoder StatusCoder
	if errors.As(err, &statusCoder) {
		resolved.Status = statusCoder.StatusCode()
	}

	var codeCarrier CodeCarrier
	if errors.As(err, &codeCarrier) {
		resolved.Code = codeCarrier.ErrorCode()
	}

	var msgCarrier PublicMessageCarrier
	if errors.As(err, &msgCarrier) {
		resolved.Message = msgCarrier.PublicMessage()
	}

	var detailsCarrier DetailsCarrier
	if errors.As(err, &detailsCarrier) {
		resolved.Details = cloneMap(detailsCarrier.ErrorDetails())
	}

	return sanitizeAppError(resolved, r.internalMessage)
}

func (r *Resolver) ResolvePanic(rec any) *AppError {
	return r.Resolve(Internal(r.internalMessage).WithCause(panicToError(rec)))
}

func sanitizeAppError(appErr *AppError, internalMessage string) *AppError {
	if appErr == nil {
		return Internal(internalMessage)
	}

	resolved := appErr.Clone()
	if resolved.Status == 0 {
		resolved.Status = http.StatusInternalServerError
	}
	if resolved.Code == "" {
		resolved.Code = CodeInternalServerError
	}
	if resolved.Message == "" {
		if resolved.Status >= http.StatusInternalServerError {
			resolved.Message = internalMessage
		} else {
			resolved.Message = http.StatusText(resolved.Status)
		}
	}
	return resolved
}

func cloneMap(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func panicToError(rec any) error {
	switch v := rec.(type) {
	case error:
		return v
	case string:
		return errors.New(v)
	default:
		return fmt.Errorf("panic: %v", v)
	}
}
