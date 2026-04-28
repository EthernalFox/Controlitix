package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

func floatPtr(value float64) *float64 {
	return &value
}

type fakeAlarmRepository struct {
	stateByTag   map[uuid.UUID]domain.AlarmStateRecord
	setpoints    map[uuid.UUID]domain.Setpoints
	events       []domain.AlarmEvent
	acks         map[uuid.UUID]domain.AlarmAck
}

func (repository *fakeAlarmRepository) LoadSetpoints(_ context.Context) (map[uuid.UUID]domain.Setpoints, error) {
	return repository.setpoints, nil
}

func (repository *fakeAlarmRepository) LoadSetpoint(
	_ context.Context,
	tagID uuid.UUID,
) (domain.Setpoints, bool, error) {
	setpoints, exists := repository.setpoints[tagID]
	return setpoints, exists, nil
}

func (repository *fakeAlarmRepository) GetState(
	_ context.Context,
	tagID uuid.UUID,
) (domain.AlarmStateRecord, bool, error) {
	record, exists := repository.stateByTag[tagID]
	return record, exists, nil
}

func (repository *fakeAlarmRepository) UpsertState(_ context.Context, record domain.AlarmStateRecord) error {
	repository.stateByTag[record.TagID] = record
	return nil
}

func (repository *fakeAlarmRepository) ApplyTransition(
	_ context.Context,
	input domain.AlarmTransitionInput,
) (domain.AlarmEvent, error) {
	record := domain.AlarmStateRecord{
		TagID:       input.TagID,
		State:       input.StateTo,
		LastValue:   input.Value,
		LastQuality: input.Quality,
		EnteredAt:   input.EnteredAt,
		LastSeenAt:  input.LastSeenAt,
	}
	repository.stateByTag[input.TagID] = record

	event := domain.AlarmEvent{
		ID:        uuid.New(),
		TagID:     input.TagID,
		StateFrom: input.StateFrom,
		StateTo:   input.StateTo,
		Value:     input.Value,
		Quality:   input.Quality,
		TS:        input.TS,
		EventType: domain.AlarmEventRaised,
	}
	if input.StateTo == domain.AlarmStateOK {
		event.EventType = domain.AlarmEventCleared
	}
	repository.events = append(repository.events, event)
	return event, nil
}

func (repository *fakeAlarmRepository) Acknowledge(
	_ context.Context,
	tagID uuid.UUID,
	actorID string,
	note *string,
	ts time.Time,
) (domain.AlarmAcknowledgeResult, error) {
	record := repository.stateByTag[tagID]
	if record.State == domain.AlarmStateOK {
		return domain.AlarmAcknowledgeResult{}, domain.ErrAlarmNotActive
	}
	ack := domain.AlarmAck{
		State:   record.State,
		ActorID: actorID,
		Note:    note,
		AckedAt: ts,
	}
	repository.acks[tagID] = ack
	event := domain.AlarmEvent{
		ID:        uuid.New(),
		TagID:     tagID,
		EventType: domain.AlarmEventAcked,
		StateFrom: record.State,
		StateTo:   record.State,
		Quality:   record.LastQuality,
		TS:        ts,
	}
	repository.events = append(repository.events, event)

	return domain.AlarmAcknowledgeResult{
		TagID: tagID,
		State: record.State,
		Ack:   ack,
		Event: event,
	}, nil
}

func (repository *fakeAlarmRepository) ListAlarms(
	_ context.Context,
	query domain.AlarmListQuery,
) (domain.AlarmListResult, error) {
	items := make([]domain.AlarmStateRecord, 0, len(repository.stateByTag))
	for _, record := range repository.stateByTag {
		if query.Status == domain.AlarmListStatusActive && record.State == domain.AlarmStateOK {
			continue
		}
		items = append(items, record)
	}
	return domain.AlarmListResult{Items: items, Total: len(items), Limit: query.Limit, Offset: query.Offset}, nil
}

func (repository *fakeAlarmRepository) GetAlarm(
	_ context.Context,
	tagID uuid.UUID,
) (domain.AlarmDetail, error) {
	record := repository.stateByTag[tagID]
	return domain.AlarmDetail{CurrentState: &record, Events: repository.events}, nil
}

func (repository *fakeAlarmRepository) ListActiveUnacked(
	_ context.Context,
	limit int,
) ([]domain.AlarmStateRecord, error) {
	items := make([]domain.AlarmStateRecord, 0)
	for _, record := range repository.stateByTag {
		if record.State != domain.AlarmStateOK {
			items = append(items, record)
		}
		if len(items) >= limit {
			break
		}
	}
	return items, nil
}

func (repository *fakeAlarmRepository) ListCommLossCandidates(
	_ context.Context,
	_ time.Time,
	_ int,
) ([]domain.AlarmStateRecord, error) {
	return nil, nil
}

func (repository *fakeAlarmRepository) DeleteByTagID(_ context.Context, tagID uuid.UUID) error {
	delete(repository.stateByTag, tagID)
	delete(repository.setpoints, tagID)
	return nil
}

type fakeAlarmPublisher struct {
	events []domain.AlarmEvent
}

func (publisher *fakeAlarmPublisher) Publish(_ context.Context, event domain.AlarmEvent) error {
	publisher.events = append(publisher.events, event)
	return nil
}

type fakeAlarmBroadcaster struct {
	events []domain.AlarmEvent
}

func (broadcaster *fakeAlarmBroadcaster) BroadcastAlarm(event domain.AlarmEvent) {
	broadcaster.events = append(broadcaster.events, event)
}

func TestAlarmUseCaseIngestValueCreatesTransition(t *testing.T) {
	tagID := uuid.New()
	repository := &fakeAlarmRepository{
		stateByTag: map[uuid.UUID]domain.AlarmStateRecord{},
		setpoints: map[uuid.UUID]domain.Setpoints{
			tagID: {
				Hi: floatPtr(80),
			},
		},
		acks: map[uuid.UUID]domain.AlarmAck{},
	}
	publisher := &fakeAlarmPublisher{}
	broadcaster := &fakeAlarmBroadcaster{}
	useCase := NewAlarmUseCase(repository, publisher, broadcaster, 2, time.Minute, nil)

	initializeError := useCase.Initialize(context.Background())
	if initializeError != nil {
		t.Fatalf("Initialize returned error: %v", initializeError)
	}

	value := 85.0
	ingestError := useCase.IngestValue(context.Background(), domain.IngestRecord{
		TagID:     tagID,
		Timestamp: time.Now().UTC(),
		Value:     &value,
		Quality:   domain.QualityOK,
	})
	if ingestError != nil {
		t.Fatalf("IngestValue returned error: %v", ingestError)
	}

	currentState := repository.stateByTag[tagID]
	if currentState.State != domain.AlarmStateHi {
		t.Fatalf("expected state hi, got %s", currentState.State.String())
	}
	if len(repository.events) != 1 {
		t.Fatalf("expected one transition event, got %d", len(repository.events))
	}
}
