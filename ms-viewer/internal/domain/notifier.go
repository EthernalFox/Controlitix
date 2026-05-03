package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type NotificationStatus int16

const (
	NotificationStatusPending NotificationStatus = 0
	NotificationStatusSent    NotificationStatus = 1
	NotificationStatusFailed  NotificationStatus = 2
	NotificationStatusSkipped NotificationStatus = 3
)

type TelegramChat struct {
	ID          uuid.UUID
	ChatID      int64
	Title       string
	UserID      *uuid.UUID
	Role        *string
	ObjectID    *uuid.UUID
	SeverityMin int16
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TelegramChatFilter struct {
	Enabled  *bool
	Role     string
	ObjectID uuid.UUID
}

type TelegramChatListResult struct {
	Items []TelegramChat
	Total int
}

type AlarmRouteContext struct {
	TagID      uuid.UUID
	TagName    string
	DeviceID   uuid.UUID
	DeviceName string
	ObjectID   uuid.UUID
	ObjectName string
	UnitSymbol string
}

type PendingNotification struct {
	NotificationID uuid.UUID
	AlarmEventID   uuid.UUID
	ChatID         int64
	Attempt        int16
	UpdatedAt      time.Time
	Event          AlarmEvent
	LastError      *string
}

type EscalationRecord struct {
	ID           uuid.UUID
	AlarmEventID uuid.UUID
	TagID        uuid.UUID
	State        AlarmState
	FireAt       time.Time
	Processed    bool
	Skipped      bool
	CreatedAt    time.Time
}

type RateLimitError struct {
	RetryAfter time.Duration
	Cause      error
}

func (rateLimitError *RateLimitError) Error() string {
	if rateLimitError == nil {
		return "rate limit"
	}
	if rateLimitError.Cause == nil {
		return fmt.Sprintf("rate limit: retry after %s", rateLimitError.RetryAfter)
	}
	return fmt.Sprintf("rate limit: retry after %s: %v", rateLimitError.RetryAfter, rateLimitError.Cause)
}

func (rateLimitError *RateLimitError) Unwrap() error {
	if rateLimitError == nil {
		return nil
	}
	return rateLimitError.Cause
}

type PermanentSendError struct {
	Cause error
}

func (permanentError *PermanentSendError) Error() string {
	if permanentError == nil {
		return "permanent send error"
	}
	if permanentError.Cause == nil {
		return "permanent send error"
	}
	return fmt.Sprintf("permanent send error: %v", permanentError.Cause)
}

func (permanentError *PermanentSendError) Unwrap() error {
	if permanentError == nil {
		return nil
	}
	return permanentError.Cause
}

func IsRateLimitError(err error) (*RateLimitError, bool) {
	var typedError *RateLimitError
	if errors.As(err, &typedError) {
		return typedError, true
	}
	return nil, false
}

func IsPermanentSendError(err error) bool {
	var typedError *PermanentSendError
	return errors.As(err, &typedError)
}

type TelegramChatsRepository interface {
	ListChats(ctx context.Context, filter TelegramChatFilter) (TelegramChatListResult, error)
	ListChatsForRouting(ctx context.Context, severityMin int16, objectID uuid.UUID) ([]TelegramChat, error)
	ListAdminChatsForObject(ctx context.Context, objectID uuid.UUID) ([]TelegramChat, error)
	GetAlarmRouteContext(ctx context.Context, tagID uuid.UUID) (AlarmRouteContext, error)
}

type NotificationRepository interface {
	CreateNotification(ctx context.Context, alarmEventID uuid.UUID, chatID int64) (bool, error)
	MarkNotificationSent(ctx context.Context, alarmEventID uuid.UUID, chatID int64) error
	MarkNotificationPendingRetry(
		ctx context.Context,
		alarmEventID uuid.UUID,
		chatID int64,
		lastError string,
		retryAfter *time.Duration,
	) error
	MarkNotificationFailed(
		ctx context.Context,
		alarmEventID uuid.UUID,
		chatID int64,
		lastError string,
	) error
	MarkNotificationSkipped(
		ctx context.Context,
		alarmEventID uuid.UUID,
		chatID int64,
		reason string,
	) error
	ListPendingNotifications(
		ctx context.Context,
		maxAttempts int,
		backoffBase time.Duration,
		limit int,
	) ([]PendingNotification, error)
	CreateEscalation(
		ctx context.Context,
		alarmEventID uuid.UUID,
		tagID uuid.UUID,
		state AlarmState,
		fireAt time.Time,
	) error
	ListDueEscalations(ctx context.Context, limit int) ([]EscalationRecord, error)
	MarkEscalationProcessed(ctx context.Context, escalationID uuid.UUID, skipped bool) error
	IsTagAcknowledgedForState(
		ctx context.Context,
		tagID uuid.UUID,
		state AlarmState,
	) (bool, error)
	GetAlarmEventByID(ctx context.Context, alarmEventID uuid.UUID) (AlarmEvent, error)
	WithSchedulerLock(
		ctx context.Context,
		lockKey int64,
		callback func(context.Context) error,
	) (bool, error)
}

type TelegramNotifierClient interface {
	SendMessage(
		ctx context.Context,
		chatID int64,
		text string,
		parseMode string,
	) error
}
