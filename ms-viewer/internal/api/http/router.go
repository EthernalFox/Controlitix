package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type TrendUseCase interface {
	GetTrend(ctx context.Context, query domain.TrendQuery) (domain.TrendSeries, error)
	GetTrendsBatch(
		ctx context.Context,
		query domain.TrendBatchQuery,
	) (domain.TrendBatchResult, error)
}

type TagUseCase interface {
	SearchTags(ctx context.Context, query domain.TagSearchQuery) (domain.TagSearchResult, error)
}

type DiagramUseCase interface {
	ListObjectsWithPublishedDiagrams(
		ctx context.Context,
		query domain.ObjectListQuery,
	) (domain.ObjectListResult, error)
	ListPublishedDiagrams(
		ctx context.Context,
		query domain.DiagramListQuery,
	) (domain.DiagramListResult, error)
	GetPublishedDiagram(ctx context.Context, diagramID uuid.UUID) (domain.Diagram, error)
	GetDiagramSnapshot(
		ctx context.Context,
		diagramID uuid.UUID,
	) (domain.DiagramSnapshot, error)
}

type AlarmUseCase interface {
	ListAlarms(ctx context.Context, query domain.AlarmListQuery) (domain.AlarmListResult, error)
	GetAlarm(ctx context.Context, tagID uuid.UUID) (domain.AlarmDetail, error)
	Acknowledge(
		ctx context.Context,
		tagID uuid.UUID,
		actorID string,
		note *string,
	) (domain.AlarmAcknowledgeResult, error)
	AcknowledgeBulk(
		ctx context.Context,
		tagIDs []uuid.UUID,
		actorID string,
		note *string,
	) (domain.AlarmBulkAcknowledgeResult, error)
}

type NotifierUseCase interface {
	ListChats(
		ctx context.Context,
		filter domain.TelegramChatFilter,
	) (domain.TelegramChatListResult, error)
}

type ReadinessChecker interface {
	Ping(ctx context.Context) error
}

type RouterOptions struct {
	TrendUseCase    TrendUseCase
	TagUseCase      TagUseCase
	DiagramUseCase  DiagramUseCase
	AlarmUseCase    AlarmUseCase
	NotifierUseCase NotifierUseCase
	AuditUseCase    AuditRecorder
	AuthMiddleware  func(http.Handler) http.Handler
	WSHandler       http.Handler
	DatabaseChecker ReadinessChecker
	RedisChecker    ReadinessChecker
	Logger          *slog.Logger
}

func NewRouter(options RouterOptions) http.Handler {
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}

	trendsHandler := NewTrendsHandler(options.TrendUseCase)
	tagsHandler := NewTagsHandler(options.TagUseCase)
	diagramsHandler := NewDiagramsHandler(options.DiagramUseCase, options.AuditUseCase)
	alarmsHandler := NewAlarmsHandler(options.AlarmUseCase, options.AuditUseCase)
	notifierHandler := NewNotifierHandler(options.NotifierUseCase)

	router := chi.NewRouter()
	router.Use(RequestID)
	router.Use(RequestLog(logger))
	router.Use(Recover(logger))

	router.Get("/healthz", func(responseWriter http.ResponseWriter, _ *http.Request) {
		writeJSON(responseWriter, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Get("/readyz", func(responseWriter http.ResponseWriter, request *http.Request) {
		if options.DatabaseChecker == nil || options.RedisChecker == nil {
			writeJSON(responseWriter, http.StatusServiceUnavailable, map[string]string{"status": "degraded"})
			return
		}

		if databaseError := options.DatabaseChecker.Ping(request.Context()); databaseError != nil {
			writeJSON(
				responseWriter,
				http.StatusServiceUnavailable,
				map[string]string{
					"status": "degraded",
					"reason": "database unreachable",
				},
			)
			return
		}

		if redisError := options.RedisChecker.Ping(request.Context()); redisError != nil {
			writeJSON(
				responseWriter,
				http.StatusServiceUnavailable,
				map[string]string{
					"status": "degraded",
					"reason": "redis unreachable",
				},
			)
			return
		}

		writeJSON(responseWriter, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Route("/api", func(apiRouter chi.Router) {
		if options.WSHandler != nil {
			apiRouter.Handle("/ws", options.WSHandler)
		}

		apiRouter.Group(func(securedRouter chi.Router) {
			if options.AuthMiddleware != nil {
				securedRouter.Use(options.AuthMiddleware)
			}
			securedRouter.With(
				RequireAnyRole(options.AuditUseCase, "operator", "engineer", "admin"),
			).Get("/trends/{tagId}", trendsHandler.GetTrend)
			securedRouter.With(
				RequireAnyRole(options.AuditUseCase, "operator", "engineer", "admin"),
			).Get("/trends", trendsHandler.GetTrends)
			securedRouter.With(
				RequireAnyRole(options.AuditUseCase, "operator", "engineer", "admin"),
			).Get("/tags", tagsHandler.GetTags)
			securedRouter.With(
				RequireAnyRole(options.AuditUseCase, "operator", "engineer", "admin"),
			).Get("/objects", diagramsHandler.GetObjects)
			securedRouter.With(
				RequireAnyRole(options.AuditUseCase, "operator", "engineer", "admin"),
			).Get("/objects/{objectId}/diagrams", diagramsHandler.GetObjectDiagrams)
			securedRouter.With(
				RequireAnyRole(options.AuditUseCase, "operator", "engineer", "admin"),
			).Get("/diagrams/{diagramId}", diagramsHandler.GetDiagram)
			securedRouter.With(
				RequireAnyRole(options.AuditUseCase, "operator", "engineer", "admin"),
			).Get("/diagrams/{diagramId}/snapshot", diagramsHandler.GetDiagramSnapshot)
			securedRouter.With(
				RequireAnyRole(options.AuditUseCase, "operator", "engineer", "admin"),
			).Get("/alarms", alarmsHandler.GetAlarms)
			securedRouter.With(
				RequireAnyRole(options.AuditUseCase, "operator", "engineer", "admin"),
			).Get("/alarms/{tagId}", alarmsHandler.GetAlarm)
			securedRouter.With(
				RequireAnyRole(options.AuditUseCase, "operator", "admin"),
			).Post("/alarms/{tagId}/acknowledge", alarmsHandler.Acknowledge)
			securedRouter.With(
				RequireAnyRole(options.AuditUseCase, "operator", "admin"),
			).Post("/alarms/acknowledge", alarmsHandler.AcknowledgeBulk)
			securedRouter.With(
				RequireAnyRole(options.AuditUseCase, "engineer", "admin"),
			).Get("/notifier/chats", notifierHandler.GetChats)
		})
	})

	return router
}
