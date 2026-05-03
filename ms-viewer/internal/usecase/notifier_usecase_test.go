package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type fakeNotifierChatsRepository struct {
	listResult    domain.TelegramChatListResult
	routeChats    []domain.TelegramChat
	adminChats    []domain.TelegramChat
	routeContext  domain.AlarmRouteContext
	listChatsCall int
}

func (repository *fakeNotifierChatsRepository) ListChats(
	_ context.Context,
	_ domain.TelegramChatFilter,
) (domain.TelegramChatListResult, error) {
	repository.listChatsCall++
	return repository.listResult, nil
}

func (repository *fakeNotifierChatsRepository) ListChatsForRouting(
	_ context.Context,
	_ int16,
	_ uuid.UUID,
) ([]domain.TelegramChat, error) {
	return repository.routeChats, nil
}

func (repository *fakeNotifierChatsRepository) ListAdminChatsForObject(
	_ context.Context,
	_ uuid.UUID,
) ([]domain.TelegramChat, error) {
	return repository.adminChats, nil
}

func (repository *fakeNotifierChatsRepository) GetAlarmRouteContext(
	_ context.Context,
	_ uuid.UUID,
) (domain.AlarmRouteContext, error) {
	return repository.routeContext, nil
}

type fakeNotificationRepository struct {
	created           map[string]bool
	skipped           []string
	sent              []string
	failed            []string
	pendingRetry      []string
	escalations       int
	dueEscalations    []domain.EscalationRecord
	alarmByID         map[uuid.UUID]domain.AlarmEvent
	lockCallCount     int
	lockAcquireResult bool
}

func newFakeNotificationRepository() *fakeNotificationRepository {
	return &fakeNotificationRepository{
		created:           make(map[string]bool),
		alarmByID:         make(map[uuid.UUID]domain.AlarmEvent),
		lockAcquireResult: true,
	}
}

func (repository *fakeNotificationRepository) CreateNotification(
	_ context.Context,
	alarmEventID uuid.UUID,
	chatID int64,
) (bool, error) {
	key := fmt.Sprintf("%s:%d", alarmEventID.String(), chatID)
	if repository.created[key] {
		return false, nil
	}
	repository.created[key] = true
	return true, nil
}

func (repository *fakeNotificationRepository) MarkNotificationSent(
	_ context.Context,
	alarmEventID uuid.UUID,
	chatID int64,
) error {
	repository.sent = append(repository.sent, fmt.Sprintf("%s:%d", alarmEventID.String(), chatID))
	return nil
}

func (repository *fakeNotificationRepository) MarkNotificationPendingRetry(
	_ context.Context,
	alarmEventID uuid.UUID,
	chatID int64,
	_ string,
	_ *time.Duration,
) error {
	repository.pendingRetry = append(repository.pendingRetry, fmt.Sprintf("%s:%d", alarmEventID.String(), chatID))
	return nil
}

func (repository *fakeNotificationRepository) MarkNotificationFailed(
	_ context.Context,
	alarmEventID uuid.UUID,
	chatID int64,
	_ string,
) error {
	repository.failed = append(repository.failed, fmt.Sprintf("%s:%d", alarmEventID.String(), chatID))
	return nil
}

func (repository *fakeNotificationRepository) MarkNotificationSkipped(
	_ context.Context,
	alarmEventID uuid.UUID,
	chatID int64,
	reason string,
) error {
	repository.skipped = append(
		repository.skipped,
		fmt.Sprintf("%s:%d:%s", alarmEventID.String(), chatID, reason),
	)
	return nil
}

func (repository *fakeNotificationRepository) ListPendingNotifications(
	_ context.Context,
	_ int,
	_ time.Duration,
	_ int,
) ([]domain.PendingNotification, error) {
	return nil, nil
}

func (repository *fakeNotificationRepository) CreateEscalation(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	_ domain.AlarmState,
	_ time.Time,
) error {
	repository.escalations++
	return nil
}

func (repository *fakeNotificationRepository) ListDueEscalations(
	_ context.Context,
	_ int,
) ([]domain.EscalationRecord, error) {
	return repository.dueEscalations, nil
}

func (repository *fakeNotificationRepository) MarkEscalationProcessed(
	_ context.Context,
	_ uuid.UUID,
	_ bool,
) error {
	return nil
}

func (repository *fakeNotificationRepository) IsTagAcknowledgedForState(
	_ context.Context,
	_ uuid.UUID,
	_ domain.AlarmState,
) (bool, error) {
	return false, nil
}

func (repository *fakeNotificationRepository) GetAlarmEventByID(
	_ context.Context,
	alarmEventID uuid.UUID,
) (domain.AlarmEvent, error) {
	return repository.alarmByID[alarmEventID], nil
}

func (repository *fakeNotificationRepository) WithSchedulerLock(
	ctx context.Context,
	_ int64,
	callback func(context.Context) error,
) (bool, error) {
	repository.lockCallCount++
	if !repository.lockAcquireResult {
		return false, nil
	}
	return true, callback(ctx)
}

type fakeNotifierClient struct {
	sent int
}

func (client *fakeNotifierClient) SendMessage(
	_ context.Context,
	_ int64,
	_ string,
	_ string,
) error {
	client.sent++
	return nil
}

type fakeRenderer struct{}

func (renderer *fakeRenderer) Render(
	_ domain.AlarmEvent,
	_ domain.AlarmRouteContext,
	_ bool,
	_ time.Duration,
) (string, error) {
	return "message", nil
}

func TestNotifierUseCaseRouteAndSendDryRunAndDedup(t *testing.T) {
	eventID := uuid.New()
	tagID := uuid.New()
	chatID := int64(123456)

	chatsRepository := &fakeNotifierChatsRepository{
		routeChats: []domain.TelegramChat{{ChatID: chatID}},
		routeContext: domain.AlarmRouteContext{
			TagID:      tagID,
			TagName:    "boil_1.t_(out)",
			DeviceID:   uuid.New(),
			DeviceName: "Boiler",
			ObjectID:   uuid.New(),
			ObjectName: "Boiler room",
		},
	}
	notificationRepository := newFakeNotificationRepository()
	notifier := NewNotifierUseCase(
		chatsRepository,
		notificationRepository,
		&fakeNotifierClient{},
		&fakeRenderer{},
		nil,
		NotifierUseCaseOptions{DryRun: true},
		nil,
	)

	event := domain.AlarmEvent{
		ID:        eventID,
		TagID:     tagID,
		EventType: domain.AlarmEventRaised,
		StateTo:   domain.AlarmStateHi,
		TS:        time.Now().UTC(),
	}
	if routeError := notifier.RouteAndSend(context.Background(), event); routeError != nil {
		t.Fatalf("RouteAndSend returned error: %v", routeError)
	}
	if routeError := notifier.RouteAndSend(context.Background(), event); routeError != nil {
		t.Fatalf("RouteAndSend duplicate returned error: %v", routeError)
	}

	if len(notificationRepository.skipped) != 1 {
		t.Fatalf("expected one skipped notification, got %d", len(notificationRepository.skipped))
	}
}
