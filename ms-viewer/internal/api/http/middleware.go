package http

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type wsPanicCloseHandlerKey struct{}

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
	return chimiddleware.RequestID(next)
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
				chimiddleware.GetReqID(request.Context()),
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
						chimiddleware.GetReqID(request.Context()),
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
