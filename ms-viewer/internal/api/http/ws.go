package http

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/realtime"
	"github.com/EthernalFox/Controlitix/shared/authctx"
)

type WSHandlerOptions struct {
	Hub                     *realtime.Hub
	Validator               *authctx.Validator
	AuditUseCase            AuditRecorder
	Logger                  *slog.Logger
	WSPingIntervalSec       int
	WSPongTimeoutSec        int
	WSWriteBufferSize       int
	WSMaxSubscriptions      int
	WSDebounceMS            int
}

type WSHandler struct {
	hub                *realtime.Hub
	validator          *authctx.Validator
	auditUseCase       AuditRecorder
	logger             *slog.Logger
	pingInterval       time.Duration
	pongTimeout        time.Duration
	writeBufferSize    int
	maxSubscriptions   int
	debounceMS         int
}

func NewWSHandler(options WSHandlerOptions) *WSHandler {
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}

	pingInterval := time.Duration(options.WSPingIntervalSec) * time.Second
	if pingInterval <= 0 {
		pingInterval = 20 * time.Second
	}
	pongTimeout := time.Duration(options.WSPongTimeoutSec) * time.Second
	if pongTimeout <= 0 {
		pongTimeout = 30 * time.Second
	}
	writeBufferSize := options.WSWriteBufferSize
	if writeBufferSize <= 0 {
		writeBufferSize = 256
	}
	maxSubscriptions := options.WSMaxSubscriptions
	if maxSubscriptions <= 0 {
		maxSubscriptions = 200
	}
	debounceMS := options.WSDebounceMS
	if debounceMS <= 0 {
		debounceMS = 100
	}

	return &WSHandler{
		hub:              options.Hub,
		validator:        options.Validator,
		auditUseCase:     options.AuditUseCase,
		logger:           logger,
		pingInterval:     pingInterval,
		pongTimeout:      pongTimeout,
		writeBufferSize:  writeBufferSize,
		maxSubscriptions: maxSubscriptions,
		debounceMS:       debounceMS,
	}
}

func (handler *WSHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if handler.hub == nil || handler.validator == nil {
		writeProblem(responseWriter, Problem{
			Type:   "/problems/internal-error",
			Title:  "Internal error",
			Status: http.StatusInternalServerError,
			Detail: "realtime hub is not configured",
		})
		return
	}

	token := extractWSToken(request)
	if token == "" {
		handler.closeWithPolicyViolation(responseWriter, request, "auth-required")
		return
	}

	principal, parseError := handler.validator.Parse(request.Context(), token)
	if parseError != nil {
		authctx.WriteProblem(responseWriter, authctx.Problem{
			Type:   "/errors/auth/invalid-token",
			Title:  "Invalid token",
			Status: http.StatusUnauthorized,
		})
		return
	}
	if !hasViewerWSRole(principal.Roles) {
		if handler.auditUseCase != nil {
			handler.auditUseCase.Record(request.Context(), domain.AuditEvent{
				Action: "auth.forbidden",
				Target: domain.AuditTarget{
					Type: "ws_endpoint",
					ID:   "/api/ws",
				},
				Details: map[string]any{
					"required_roles": []string{"operator", "engineer", "admin"},
				},
				Result: domain.AuditResultFailure,
			})
		}
		writeProblem(responseWriter, Problem{
			Type:   "/errors/auth/forbidden",
			Title:  "Forbidden",
			Status: http.StatusForbidden,
		})
		return
	}

	socket, acceptError := websocket.Accept(responseWriter, request, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionDisabled,
	})
	if acceptError != nil {
		handler.logger.Warn(
			"failed to accept websocket connection",
			"method",
			"WSHandler.ServeHTTP",
			"error",
			acceptError,
		)
		return
	}

	request = registerWSPanicCloseHandler(request, func() {
		_ = socket.Close(websocket.StatusInternalError, "internal_error")
	})

	connection := realtime.NewConnection(realtime.ConnectionOptions{
		Hub:             handler.hub,
		Socket:          socket,
		Subject:         principal.Subject,
		Roles:           principal.Roles,
		SessionID:       uuid.NewString(),
		Logger:          handler.logger,
		WriteBufferSize: handler.writeBufferSize,
		PingInterval:    handler.pingInterval,
		PongTimeout:     handler.pongTimeout,
	})

	if displaced := handler.hub.RegisterConnection(connection); displaced != nil {
		displaced.Close(websocket.StatusGoingAway, "connection_limit_exceeded")
	}
	if handler.auditUseCase != nil {
		handler.auditUseCase.Record(request.Context(), domain.AuditEvent{
			Action: "ws.connected",
			Target: domain.AuditTarget{
				Type: "ws_connection",
				ID:   connection.SessionID(),
			},
			Result: domain.AuditResultSuccess,
		})
	}

	connection.EnqueueWelcome(handler.maxSubscriptions, handler.debounceMS)
	connection.Run(request.Context())
	if handler.auditUseCase != nil {
		handler.auditUseCase.Record(request.Context(), domain.AuditEvent{
			Action: "ws.disconnected",
			Target: domain.AuditTarget{
				Type: "ws_connection",
				ID:   connection.SessionID(),
			},
			Details: map[string]any{
				"duration_ms": connection.DurationMS(),
				"close_code":  connection.CloseCode(),
				"close_reason": connection.CloseReason(),
			},
			Result: domain.AuditResultSuccess,
		})
	}
}

func (handler *WSHandler) closeWithPolicyViolation(
	responseWriter http.ResponseWriter,
	request *http.Request,
	reason string,
) {
	socket, acceptError := websocket.Accept(responseWriter, request, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionDisabled,
	})
	if acceptError != nil {
		writeProblem(responseWriter, Problem{
			Type:   "/errors/auth/invalid-token",
			Title:  "Invalid token",
			Status: http.StatusUnauthorized,
		})
		return
	}

	_ = socket.Close(websocket.StatusPolicyViolation, strings.TrimSpace(reason))
}

func extractWSToken(request *http.Request) string {
	if request == nil {
		return ""
	}

	if token := strings.TrimSpace(request.URL.Query().Get("access_token")); token != "" {
		return token
	}

	authorizationHeader := strings.TrimSpace(request.Header.Get("Authorization"))
	parts := strings.Fields(authorizationHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func hasViewerWSRole(roles []string) bool {
	for _, role := range roles {
		switch strings.ToLower(strings.TrimSpace(role)) {
		case "operator", "engineer", "admin":
			return true
		}
	}
	return false
}
