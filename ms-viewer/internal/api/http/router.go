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
}

type ReadinessChecker interface {
	Ping(ctx context.Context) error
}

type RouterOptions struct {
	TrendUseCase    TrendUseCase
	TagUseCase      TagUseCase
	DiagramUseCase  DiagramUseCase
	AlarmUseCase    AlarmUseCase
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
	diagramsHandler := NewDiagramsHandler(options.DiagramUseCase)
	alarmsHandler := NewAlarmsHandler(options.AlarmUseCase)

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
			securedRouter.Get("/trends/{tagId}", trendsHandler.GetTrend)
			securedRouter.Get("/trends", trendsHandler.GetTrends)
			securedRouter.Get("/tags", tagsHandler.GetTags)
			securedRouter.Get("/objects", diagramsHandler.GetObjects)
			securedRouter.Get("/objects/{objectId}/diagrams", diagramsHandler.GetObjectDiagrams)
			securedRouter.Get("/diagrams/{diagramId}", diagramsHandler.GetDiagram)
			securedRouter.Get("/diagrams/{diagramId}/snapshot", diagramsHandler.GetDiagramSnapshot)
			securedRouter.Get("/alarms", alarmsHandler.GetAlarms)
			securedRouter.Get("/alarms/{tagId}", alarmsHandler.GetAlarm)
			securedRouter.Post("/alarms/{tagId}/acknowledge", alarmsHandler.Acknowledge)
		})
	})

	return router
}
