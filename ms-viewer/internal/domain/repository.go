package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TrendRepository interface {
	GetTagMeta(ctx context.Context, tagID uuid.UUID) (TagMeta, error)
	GetRawPoints(ctx context.Context, tagID uuid.UUID, from time.Time, to time.Time) ([]TrendPoint, error)
	GetRawPointsByStep(
		ctx context.Context,
		tagID uuid.UUID,
		from time.Time,
		to time.Time,
		step time.Duration,
		aggregator Aggregator,
	) ([]TrendPoint, error)
	GetAggregatedPoints(
		ctx context.Context,
		tagID uuid.UUID,
		from time.Time,
		to time.Time,
		step time.Duration,
		aggregator Aggregator,
	) ([]TrendPoint, error)
}

type TagRepository interface {
	SearchTags(ctx context.Context, query TagSearchQuery) (TagSearchResult, error)
}

type ValueIngestRepository interface {
	UpsertRawValues(ctx context.Context, records []IngestRecord) error
}

type DiagramRepository interface {
	ListObjectsWithPublishedDiagrams(
		ctx context.Context,
		query ObjectListQuery,
	) (ObjectListResult, error)
	ListPublishedDiagrams(
		ctx context.Context,
		query DiagramListQuery,
	) (DiagramListResult, error)
	GetPublishedDiagram(ctx context.Context, diagramID uuid.UUID) (Diagram, error)
	ListDiagramBoundTagIDs(ctx context.Context, diagramID uuid.UUID) ([]uuid.UUID, error)
}

type AlarmRepository interface {
	LoadSetpoints(ctx context.Context) (map[uuid.UUID]Setpoints, error)
	LoadSetpoint(ctx context.Context, tagID uuid.UUID) (Setpoints, bool, error)
	GetState(ctx context.Context, tagID uuid.UUID) (AlarmStateRecord, bool, error)
	UpsertState(ctx context.Context, record AlarmStateRecord) error
	ApplyTransition(ctx context.Context, input AlarmTransitionInput) (AlarmEvent, error)
	Acknowledge(
		ctx context.Context,
		tagID uuid.UUID,
		actorID string,
		note *string,
		ts time.Time,
	) (AlarmAcknowledgeResult, error)
	AcknowledgeBulk(
		ctx context.Context,
		tagIDs []uuid.UUID,
		actorID string,
		note *string,
		ts time.Time,
	) (AlarmBulkAcknowledgeResult, error)
	ListAlarms(ctx context.Context, query AlarmListQuery) (AlarmListResult, error)
	GetAlarm(ctx context.Context, tagID uuid.UUID) (AlarmDetail, error)
	ListActiveUnacked(ctx context.Context, limit int) ([]AlarmStateRecord, error)
	ListCommLossCandidates(ctx context.Context, threshold time.Time, limit int) ([]AlarmStateRecord, error)
	DeleteByTagID(ctx context.Context, tagID uuid.UUID) error
}
