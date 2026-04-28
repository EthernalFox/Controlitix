package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

const (
	defaultAlarmListLimit = 50
	maxAlarmListLimit     = 500
)

type AlarmEventPublisher interface {
	Publish(ctx context.Context, event domain.AlarmEvent) error
}

type AlarmRealtimeBroadcaster interface {
	BroadcastAlarm(event domain.AlarmEvent)
}

type AlarmUseCase struct {
	repository         domain.AlarmRepository
	publisher          AlarmEventPublisher
	realtimeBroadcaster AlarmRealtimeBroadcaster
	hysteresisPercent  float64
	commLossTimeout    time.Duration
	logger             *slog.Logger

	setpointsMutex sync.RWMutex
	setpointsByTag map[uuid.UUID]domain.Setpoints
}

func NewAlarmUseCase(
	repository domain.AlarmRepository,
	publisher AlarmEventPublisher,
	realtimeBroadcaster AlarmRealtimeBroadcaster,
	hysteresisPercent float64,
	commLossTimeout time.Duration,
	logger *slog.Logger,
) *AlarmUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	if hysteresisPercent <= 0 {
		hysteresisPercent = 2.0
	}
	if commLossTimeout <= 0 {
		commLossTimeout = 60 * time.Second
	}

	return &AlarmUseCase{
		repository:          repository,
		publisher:           publisher,
		realtimeBroadcaster: realtimeBroadcaster,
		hysteresisPercent:   hysteresisPercent,
		commLossTimeout:     commLossTimeout,
		logger:              logger,
		setpointsByTag:      make(map[uuid.UUID]domain.Setpoints),
	}
}

func (useCase *AlarmUseCase) Initialize(ctx context.Context) error {
	if useCase.repository == nil {
		return errors.New("alarm repository is not configured")
	}

	setpointsByTag, loadError := useCase.repository.LoadSetpoints(ctx)
	if loadError != nil {
		return fmt.Errorf("load alarm setpoints: %w", loadError)
	}

	useCase.setpointsMutex.Lock()
	useCase.setpointsByTag = setpointsByTag
	useCase.setpointsMutex.Unlock()

	return nil
}

func (useCase *AlarmUseCase) IngestValue(ctx context.Context, record domain.IngestRecord) error {
	if useCase.repository == nil {
		return nil
	}

	currentStateRecord, exists, stateError := useCase.repository.GetState(ctx, record.TagID)
	if stateError != nil {
		return stateError
	}

	prevState := domain.AlarmStateOK
	enteredAt := record.Timestamp.UTC()
	if exists {
		prevState = currentStateRecord.State
		enteredAt = currentStateRecord.EnteredAt.UTC()
	}

	setpoints := useCase.getSetpoints(record.TagID)
	nextState, changed := domain.EvaluateAlarmState(
		prevState,
		record.Value,
		record.Quality,
		setpoints,
		useCase.hysteresisPercent,
	)

	if !exists || !changed {
		if changed {
			enteredAt = record.Timestamp.UTC()
		}

		upsertError := useCase.repository.UpsertState(ctx, domain.AlarmStateRecord{
			TagID:       record.TagID,
			State:       nextState,
			LastValue:   record.Value,
			LastQuality: record.Quality,
			EnteredAt:   enteredAt,
			LastSeenAt:  record.Timestamp.UTC(),
			Suppressed:  false,
		})
		if upsertError != nil {
			return upsertError
		}
		if !changed {
			return nil
		}
	}

	transitionEvent, transitionError := useCase.repository.ApplyTransition(ctx, domain.AlarmTransitionInput{
		TagID:      record.TagID,
		StateFrom:  prevState,
		StateTo:    nextState,
		Value:      record.Value,
		Quality:    record.Quality,
		TS:         record.Timestamp.UTC(),
		EnteredAt:  record.Timestamp.UTC(),
		LastSeenAt: record.Timestamp.UTC(),
	})
	if transitionError != nil {
		return transitionError
	}

	useCase.publishAndBroadcastEvent(transitionEvent)
	return nil
}

func (useCase *AlarmUseCase) Acknowledge(
	ctx context.Context,
	tagID uuid.UUID,
	actorID string,
	note *string,
) (domain.AlarmAcknowledgeResult, error) {
	result, acknowledgeError := useCase.repository.Acknowledge(
		ctx,
		tagID,
		strings.TrimSpace(actorID),
		note,
		time.Now().UTC(),
	)
	if acknowledgeError != nil {
		return domain.AlarmAcknowledgeResult{}, acknowledgeError
	}

	useCase.publishAndBroadcastEvent(result.Event)
	return result, nil
}

func (useCase *AlarmUseCase) ListAlarms(
	ctx context.Context,
	query domain.AlarmListQuery,
) (domain.AlarmListResult, error) {
	normalizedQuery := query
	if normalizedQuery.Limit <= 0 {
		normalizedQuery.Limit = defaultAlarmListLimit
	}
	if normalizedQuery.Limit > maxAlarmListLimit {
		return domain.AlarmListResult{}, fmt.Errorf("limit exceeds %d: %w", maxAlarmListLimit, domain.ErrInvalidInput)
	}
	if normalizedQuery.Offset < 0 {
		return domain.AlarmListResult{}, fmt.Errorf("offset must be non-negative: %w", domain.ErrInvalidInput)
	}

	return useCase.repository.ListAlarms(ctx, normalizedQuery)
}

func (useCase *AlarmUseCase) GetAlarm(ctx context.Context, tagID uuid.UUID) (domain.AlarmDetail, error) {
	if tagID == uuid.Nil {
		return domain.AlarmDetail{}, fmt.Errorf("tag id is required: %w", domain.ErrInvalidInput)
	}

	return useCase.repository.GetAlarm(ctx, tagID)
}

func (useCase *AlarmUseCase) GetActiveAlarmSnapshot(ctx context.Context) ([]domain.AlarmStateRecord, error) {
	if useCase.repository == nil {
		return nil, nil
	}

	return useCase.repository.ListActiveUnacked(ctx, 0)
}

func (useCase *AlarmUseCase) HandleTagConfigChanged(ctx context.Context, tagID string, operation string) {
	if useCase.repository == nil {
		return
	}

	parsedTagID, parseError := uuid.Parse(strings.TrimSpace(tagID))
	if parseError != nil {
		return
	}

	if strings.EqualFold(strings.TrimSpace(operation), "deleted") {
		useCase.setpointsMutex.Lock()
		delete(useCase.setpointsByTag, parsedTagID)
		useCase.setpointsMutex.Unlock()
		_ = useCase.repository.DeleteByTagID(ctx, parsedTagID)
		return
	}

	setpoints, exists, loadError := useCase.repository.LoadSetpoint(ctx, parsedTagID)
	if loadError != nil {
		useCase.logger.Warn(
			"failed to reload tag setpoints",
			"method",
			"AlarmUseCase.HandleTagConfigChanged",
			"tag_id",
			parsedTagID.String(),
			"error",
			loadError,
		)
		return
	}

	useCase.setpointsMutex.Lock()
	if exists {
		useCase.setpointsByTag[parsedTagID] = setpoints
	} else {
		delete(useCase.setpointsByTag, parsedTagID)
	}
	useCase.setpointsMutex.Unlock()
}

func (useCase *AlarmUseCase) RunCommLossScanner(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			threshold := time.Now().UTC().Add(-useCase.commLossTimeout)
			candidates, candidatesError := useCase.repository.ListCommLossCandidates(ctx, threshold, 200)
			if candidatesError != nil {
				useCase.logger.Warn(
					"failed to scan comm loss candidates",
					"method",
					"AlarmUseCase.RunCommLossScanner",
					"error",
					candidatesError,
				)
				continue
			}

			for _, candidate := range candidates {
				ingestError := useCase.IngestValue(ctx, domain.IngestRecord{
					TagID:     candidate.TagID,
					Timestamp: time.Now().UTC(),
					Value:     candidate.LastValue,
					Quality:   domain.QualityCommLoss,
				})
				if ingestError != nil {
					useCase.logger.Warn(
						"failed to set comm loss state",
						"method",
						"AlarmUseCase.RunCommLossScanner",
						"tag_id",
						candidate.TagID.String(),
						"error",
						ingestError,
					)
				}
			}
		}
	}
}

func (useCase *AlarmUseCase) getSetpoints(tagID uuid.UUID) domain.Setpoints {
	useCase.setpointsMutex.RLock()
	defer useCase.setpointsMutex.RUnlock()
	return useCase.setpointsByTag[tagID]
}

func (useCase *AlarmUseCase) publishAndBroadcastEvent(event domain.AlarmEvent) {
	if useCase.publisher != nil {
		go func() {
			for attempt := 0; attempt < 3; attempt++ {
				publishError := useCase.publisher.Publish(context.Background(), event)
				if publishError == nil {
					return
				}

				useCase.logger.Warn(
					"failed to publish alarm event",
					"method",
					"AlarmUseCase.publishAndBroadcastEvent",
					"tag_id",
					event.TagID.String(),
					"event_type",
					event.EventType.String(),
					"attempt",
					attempt+1,
					"error",
					publishError,
				)
				time.Sleep(200 * time.Millisecond)
			}
		}()
	}

	if useCase.realtimeBroadcaster != nil {
		useCase.realtimeBroadcaster.BroadcastAlarm(event)
	}
}
