package usecase

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type bulkPublisherSpy struct {
	mutex      sync.Mutex
	batchCalls int
	batchSize  int
	called     chan struct{}
}

func (spy *bulkPublisherSpy) Publish(_ context.Context, _ domain.AlarmEvent) error {
	return nil
}

func (spy *bulkPublisherSpy) PublishBatch(_ context.Context, events []domain.AlarmEvent) error {
	spy.mutex.Lock()
	spy.batchCalls++
	spy.batchSize = len(events)
	spy.mutex.Unlock()

	select {
	case spy.called <- struct{}{}:
	default:
	}

	return nil
}

func (spy *bulkPublisherSpy) snapshot() (int, int) {
	spy.mutex.Lock()
	defer spy.mutex.Unlock()
	return spy.batchCalls, spy.batchSize
}

type bulkBroadcasterSpy struct {
	mutex      sync.Mutex
	batchCalls int
	batchSize  int
}

func (spy *bulkBroadcasterSpy) BroadcastAlarm(_ domain.AlarmEvent) {}

func (spy *bulkBroadcasterSpy) BroadcastAlarmBatch(events []domain.AlarmEvent) {
	spy.mutex.Lock()
	defer spy.mutex.Unlock()
	spy.batchCalls++
	spy.batchSize = len(events)
}

func (spy *bulkBroadcasterSpy) snapshot() (int, int) {
	spy.mutex.Lock()
	defer spy.mutex.Unlock()
	return spy.batchCalls, spy.batchSize
}

func TestAlarmUseCaseAcknowledgeBulkMixedResult(t *testing.T) {
	tagAcked := uuid.New()
	tagNotActive := uuid.New()
	tagMissing := uuid.New()

	repository := &fakeAlarmRepository{
		stateByTag: map[uuid.UUID]domain.AlarmStateRecord{
			tagAcked: {
				TagID:       tagAcked,
				State:       domain.AlarmStateHi,
				LastQuality: domain.QualityOK,
			},
			tagNotActive: {
				TagID:       tagNotActive,
				State:       domain.AlarmStateOK,
				LastQuality: domain.QualityOK,
			},
		},
		setpoints: map[uuid.UUID]domain.Setpoints{},
		acks:      map[uuid.UUID]domain.AlarmAck{},
	}
	publisher := &bulkPublisherSpy{called: make(chan struct{}, 1)}
	broadcaster := &bulkBroadcasterSpy{}
	useCase := NewAlarmUseCase(repository, publisher, broadcaster, nil, 2, time.Minute, nil)

	result, acknowledgeError := useCase.AcknowledgeBulk(
		context.Background(),
		[]uuid.UUID{tagAcked, tagNotActive, tagMissing},
		"operator-1",
		nil,
	)
	if acknowledgeError != nil {
		t.Fatalf("AcknowledgeBulk returned error: %v", acknowledgeError)
	}
	if result.SuccessN != 1 || result.FailedN != 2 {
		t.Fatalf("unexpected counters: success=%d failed=%d", result.SuccessN, result.FailedN)
	}
	if len(result.Items) != 3 {
		t.Fatalf("expected 3 item results, got %d", len(result.Items))
	}
	if result.Items[0].Status != domain.AlarmBulkAckStatusAcked {
		t.Fatalf("expected first item status acked, got %s", result.Items[0].Status)
	}
	if result.Items[1].Status != domain.AlarmBulkAckStatusNotActive {
		t.Fatalf("expected second item status not_active, got %s", result.Items[1].Status)
	}
	if result.Items[2].Status != domain.AlarmBulkAckStatusNotFound {
		t.Fatalf("expected third item status not_found, got %s", result.Items[2].Status)
	}

	select {
	case <-publisher.called:
	case <-time.After(2 * time.Second):
		t.Fatal("expected one publisher batch call")
	}

	publisherCalls, publisherBatchSize := publisher.snapshot()
	if publisherCalls != 1 || publisherBatchSize != 1 {
		t.Fatalf("unexpected publisher stats: calls=%d size=%d", publisherCalls, publisherBatchSize)
	}

	broadcasterCalls, broadcasterBatchSize := broadcaster.snapshot()
	if broadcasterCalls != 1 || broadcasterBatchSize != 1 {
		t.Fatalf("unexpected broadcaster stats: calls=%d size=%d", broadcasterCalls, broadcasterBatchSize)
	}
}

func TestAlarmUseCaseAcknowledgeBulkIdempotent(t *testing.T) {
	tagID := uuid.New()
	repository := &fakeAlarmRepository{
		stateByTag: map[uuid.UUID]domain.AlarmStateRecord{
			tagID: {
				TagID:       tagID,
				State:       domain.AlarmStateHi,
				LastQuality: domain.QualityOK,
			},
		},
		setpoints: map[uuid.UUID]domain.Setpoints{},
		acks:      map[uuid.UUID]domain.AlarmAck{},
	}
	publisher := &bulkPublisherSpy{called: make(chan struct{}, 2)}
	broadcaster := &bulkBroadcasterSpy{}
	useCase := NewAlarmUseCase(repository, publisher, broadcaster, nil, 2, time.Minute, nil)

	firstResult, firstError := useCase.AcknowledgeBulk(
		context.Background(),
		[]uuid.UUID{tagID},
		"operator-1",
		nil,
	)
	if firstError != nil {
		t.Fatalf("first AcknowledgeBulk returned error: %v", firstError)
	}
	if firstResult.SuccessN != 1 || firstResult.FailedN != 0 {
		t.Fatalf("unexpected first counters: success=%d failed=%d", firstResult.SuccessN, firstResult.FailedN)
	}

	select {
	case <-publisher.called:
	case <-time.After(2 * time.Second):
		t.Fatal("expected first publisher batch call")
	}

	secondResult, secondError := useCase.AcknowledgeBulk(
		context.Background(),
		[]uuid.UUID{tagID},
		"operator-1",
		nil,
	)
	if secondError != nil {
		t.Fatalf("second AcknowledgeBulk returned error: %v", secondError)
	}
	if secondResult.SuccessN != 0 || secondResult.FailedN != 1 {
		t.Fatalf("unexpected second counters: success=%d failed=%d", secondResult.SuccessN, secondResult.FailedN)
	}
	if secondResult.Items[0].Status != domain.AlarmBulkAckStatusAlreadyAcked {
		t.Fatalf("expected already_acked status, got %s", secondResult.Items[0].Status)
	}

	time.Sleep(200 * time.Millisecond)
	publisherCalls, publisherBatchSize := publisher.snapshot()
	if publisherCalls != 1 || publisherBatchSize != 1 {
		t.Fatalf("unexpected publisher stats after second call: calls=%d size=%d", publisherCalls, publisherBatchSize)
	}

	broadcasterCalls, broadcasterBatchSize := broadcaster.snapshot()
	if broadcasterCalls != 1 || broadcasterBatchSize != 1 {
		t.Fatalf("unexpected broadcaster stats after second call: calls=%d size=%d", broadcasterCalls, broadcasterBatchSize)
	}
}
