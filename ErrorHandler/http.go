package ErrorHandler

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"
)

type HandlerFunc func(http.ResponseWriter, *http.Request) error

type HTTPResponse struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"request_id,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
	Timestamp string         `json:"timestamp,omitempty"`
}

type HTTPLogEntry struct {
	RequestID string
	Method    string
	Path      string
	Status    int
	Code      string
	Message   string
	Error     error
	Stack     string
}

type HTTPLogFunc func(context.Context, *HTTPLogEntry)

type HTTPResponseWriterFunc func(http.ResponseWriter, *http.Request, int, HTTPResponse) error

type HTTPHandler struct {
	resolver         *Resolver
	requestIDHeader  string
	includeDetails   bool
	includeTimestamp bool
	logStack         bool
	writeResponse    HTTPResponseWriterFunc
	logFunc          HTTPLogFunc
	now              func() time.Time
}

type HTTPOption func(*HTTPHandler)

func NewHTTPHandler(opts ...HTTPOption) *HTTPHandler {
	h := &HTTPHandler{
		resolver:         NewResolver(),
		requestIDHeader:  "X-Request-ID",
		includeDetails:   true,
		includeTimestamp: true,
		logStack:         true,
		writeResponse:    writeJSONResponse,
		now:              time.Now,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func NewHandler(opts ...HTTPOption) *HTTPHandler {
	return NewHTTPHandler(opts...)
}

func WithResolver(resolver *Resolver) HTTPOption {
	return func(h *HTTPHandler) {
		if resolver != nil {
			h.resolver = resolver
		}
	}
}

func WithInternalMessage(message string) HTTPOption {
	return func(h *HTTPHandler) {
		if h.resolver == nil {
			h.resolver = NewResolver()
		}
		h.resolver.internalMessage = message
	}
}

func WithRequestIDHeader(header string) HTTPOption {
	return func(h *HTTPHandler) {
		if header != "" {
			h.requestIDHeader = header
		}
	}
}

func WithDetails(enabled bool) HTTPOption {
	return func(h *HTTPHandler) {
		h.includeDetails = enabled
	}
}

func WithTimestamp(enabled bool) HTTPOption {
	return func(h *HTTPHandler) {
		h.includeTimestamp = enabled
	}
}

func WithStackLogging(enabled bool) HTTPOption {
	return func(h *HTTPHandler) {
		h.logStack = enabled
	}
}

func WithLogger(logFunc HTTPLogFunc) HTTPOption {
	return func(h *HTTPHandler) {
		h.logFunc = logFunc
	}
}

func WithResponseWriter(writer HTTPResponseWriterFunc) HTTPOption {
	return func(h *HTTPHandler) {
		if writer != nil {
			h.writeResponse = writer
		}
	}
}

func WithMappings(mappings ...Mapping) HTTPOption {
	return func(h *HTTPHandler) {
		if h.resolver == nil {
			h.resolver = NewResolver()
		}
		h.resolver.mappings = append(h.resolver.mappings, mappings...)
	}
}

func (h *HTTPHandler) RegisterMapping(mapping Mapping) {
	h.resolver.RegisterMapping(mapping)
}

func (h *HTTPHandler) Register(target error, appErr *AppError) {
	h.resolver.Register(target, appErr)
}

func (h *HTTPHandler) Resolve(err error) *AppError {
	return h.resolver.Resolve(err)
}

func (h *HTTPHandler) Wrap(next HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tw := &trackingWriter{ResponseWriter: w}
		defer h.recoverPanic(tw, r)

		if err := next(tw, r); err != nil {
			h.handleError(tw, r, err, "")
		}
	})
}

func (h *HTTPHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tw := &trackingWriter{ResponseWriter: w}
		defer h.recoverPanic(tw, r)
		next.ServeHTTP(tw, r)
	})
}

func (h *HTTPHandler) Handle(w http.ResponseWriter, r *http.Request, err error) {
	tw := &trackingWriter{ResponseWriter: w}
	h.handleError(tw, r, err, "")
}

func (h *HTTPHandler) handleError(w *trackingWriter, r *http.Request, err error, stack string) {
	if err == nil {
		return
	}

	appErr := h.resolver.Resolve(err)
	h.logError(r.Context(), r, appErr, stack)

	if w.WroteHeader() {
		return
	}

	resp := HTTPResponse{
		Code:      appErr.Code,
		Message:   appErr.Message,
		RequestID: requestIDFromRequest(r, h.requestIDHeader),
	}
	if h.includeDetails && len(appErr.Details) > 0 {
		resp.Details = cloneMap(appErr.Details)
	}
	if h.includeTimestamp {
		resp.Timestamp = h.now().Format(time.RFC3339)
	}

	if err := h.writeResponse(w, r, appErr.Status, resp); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *HTTPHandler) recoverPanic(w *trackingWriter, r *http.Request) {
	if rec := recover(); rec != nil {
		stack := ""
		if h.logStack {
			stack = string(debug.Stack())
		}
		h.handleError(w, r, h.resolver.ResolvePanic(rec), stack)
	}
}

func (h *HTTPHandler) logError(ctx context.Context, r *http.Request, appErr *AppError, stack string) {
	if h.logFunc == nil || appErr == nil {
		return
	}
	h.logFunc(ctx, &HTTPLogEntry{
		RequestID: requestIDFromRequest(r, h.requestIDHeader),
		Method:    r.Method,
		Path:      r.URL.Path,
		Status:    appErr.Status,
		Code:      appErr.Code,
		Message:   appErr.Message,
		Error:     appErr,
		Stack:     stack,
	})
}

func writeJSONResponse(w http.ResponseWriter, _ *http.Request, status int, resp HTTPResponse) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(resp)
}

func requestIDFromRequest(r *http.Request, header string) string {
	if r == nil || header == "" {
		return ""
	}
	return r.Header.Get(header)
}
