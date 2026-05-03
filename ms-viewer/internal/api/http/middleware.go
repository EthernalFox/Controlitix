package http

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/auditctx"
	"github.com/google/uuid"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type wsPanicCloseHandlerKey struct{}
type requestIDContextKey struct{}

const requestIDHeaderName = "X-Request-Id"

func registerWSPanicCloseHandler(
	request *http.Request,
	closeHandler func(),
) *http.Request {
	if request == nil || closeHandler == nil {
		return request
	}

	return request.WithContext(
		context.WithValue(request.Context(), wsPanicCloseHandlerKey{}, closeHandler),
	)
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		requestID := strings.TrimSpace(request.Header.Get(requestIDHeaderName))
		if requestID == "" {
			requestID = uuid.NewString()
		}
		responseWriter.Header().Set(requestIDHeaderName, requestID)

		requestContext := context.WithValue(request.Context(), requestIDContextKey{}, requestID)
		requestContext = auditctx.WithRequestMetadata(requestContext, auditctx.RequestMetadata{
			RequestID: requestID,
			IP:        strings.TrimSpace(request.RemoteAddr),
			UserAgent: strings.TrimSpace(request.UserAgent()),
		})
		next.ServeHTTP(responseWriter, request.WithContext(requestContext))
	})
}

func RequestLog(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			startedAt := time.Now()
			wrappedWriter := chimiddleware.NewWrapResponseWriter(responseWriter, request.ProtoMajor)

			next.ServeHTTP(wrappedWriter, request)

			logger.Info(
				"http request handled",
				"method",
				request.Method,
				"path",
				request.URL.Path,
				"status",
				wrappedWriter.Status(),
				"latency_ms",
				time.Since(startedAt).Milliseconds(),
				"request_id",
				RequestIDFromContext(request.Context()),
			)
		})
	}
}

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			defer func() {
				if recoveredValue := recover(); recoveredValue != nil {
					logger.Error(
						"panic recovered",
						"method",
						"Recover",
						"error",
						recoveredValue,
						"stack",
						string(debug.Stack()),
						"request_id",
						RequestIDFromContext(request.Context()),
					)

					if closeHandler, ok := request.Context().Value(wsPanicCloseHandlerKey{}).(func()); ok {
						closeHandler()
						return
					}

					writeProblem(responseWriter, Problem{
						Type:   "/problems/internal-error",
						Title:  "Internal error",
						Status: http.StatusInternalServerError,
						Detail: "internal error",
					})
				}
			}()

			next.ServeHTTP(responseWriter, request)
		})
	}
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return strings.TrimSpace(requestID)
}
