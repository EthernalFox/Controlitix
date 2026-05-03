package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

const (
	defaultNotifierMaxAttempts    = 5
	defaultNotifierBatchLimit     = 100
	defaultNotifierBackoffBase    = 60 * time.Second
	defaultNotifierPollInterval   = 30 * time.Second
	defaultNotifierEscalationWait = 5 * time.Minute
	notifierSchedulerLockKey      = 350035
	telegramParseMode             = "MarkdownV2"
	dryRunSkipReason              = "no_bot_token"
)

type NotifierMessageRenderer interface {
	Render(
		event domain.AlarmEvent,
		routeContext domain.AlarmRouteContext,
		escalated bool,
		escalationDelay time.Duration,
	) (string, error)
}

type NotifierUseCase struct {
	chatsRepository        domain.TelegramChatsRepository
	notificationRepository domain.NotificationRepository
	notifierClient         domain.TelegramNotifierClient
	messageRenderer        NotifierMessageRenderer
	auditUseCase           *AuditUseCase

	maxAttempts     int
	backoffBase     time.Duration
	escalationDelay time.Duration
	batchLimit      int
	pollInterval    time.Duration
	schedulerLock   int64
	dryRun          bool

	logger *slog.Logger
}

type NotifierUseCaseOptions struct {
	MaxAttempts     int
	BackoffBase     time.Duration
	EscalationDelay time.Duration
	BatchLimit      int
	PollInterval    time.Duration
	SchedulerLock   int64
	DryRun          bool
}

func NewNotifierUseCase(
	chatsRepository domain.TelegramChatsRepository,
	notificationRepository domain.NotificationRepository,
	notifierClient domain.TelegramNotifierClient,
	messageRenderer NotifierMessageRenderer,
	auditUseCase *AuditUseCase,
	options NotifierUseCaseOptions,
	logger *slog.Logger,
) *NotifierUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	if options.MaxAttempts <= 0 {
		options.MaxAttempts = defaultNotifierMaxAttempts
	}
	if options.BackoffBase <= 0 {
		options.BackoffBase = defaultNotifierBackoffBase
	}
	if options.EscalationDelay <= 0 {
		options.EscalationDelay = defaultNotifierEscalationWait
	}
	if options.BatchLimit <= 0 {
		options.BatchLimit = defaultNotifierBatchLimit
	}
	if options.PollInterval <= 0 {
		options.PollInterval = defaultNotifierPollInterval
	}
	if options.SchedulerLock == 0 {
		options.SchedulerLock = notifierSchedulerLockKey
	}

	return &NotifierUseCase{
		chatsRepository:        chatsRepository,
		notificationRepository: notificationRepository,
		notifierClient:         notifierClient,
		messageRenderer:        messageRenderer,
		auditUseCase:           auditUseCase,
		maxAttempts:            options.MaxAttempts,
		backoffBase:            options.BackoffBase,
		escalationDelay:        options.EscalationDelay,
		batchLimit:             options.BatchLimit,
		pollInterval:           options.PollInterval,
		schedulerLock:          options.SchedulerLock,
		dryRun:                 options.DryRun,
		logger:                 logger,
	}
}

func (useCase *NotifierUseCase) ListChats(
	ctx context.Context,
	filter domain.TelegramChatFilter,
) (domain.TelegramChatListResult, error) {
	if useCase.chatsRepository == nil {
		return domain.TelegramChatListResult{}, errors.New("telegram chats repository is not configured")
	}
	return useCase.chatsRepository.ListChats(ctx, filter)
}

func (useCase *NotifierUseCase) RouteAndSend(ctx context.Context, event domain.AlarmEvent) error {
	if useCase.chatsRepository == nil || useCase.notificationRepository == nil {
		return errors.New("notifier repositories are not configured")
	}

	routeContext, routeContextError := useCase.chatsRepository.GetAlarmRouteContext(ctx, event.TagID)
	if routeContextError != nil {
		return fmt.Errorf("resolve route context: %w", routeContextError)
	}

	severity := mapSeverity(event)
	chats, chatsError := useCase.chatsRepository.ListChatsForRouting(
		ctx,
		severity,
		routeContext.ObjectID,
	)
	if chatsError != nil {
		return fmt.Errorf("load chats for routing: %w", chatsError)
	}

	for _, chat := range chats {
		created, createError := useCase.notificationRepository.CreateNotification(ctx, event.ID, chat.ChatID)
		if createError != nil {
			return fmt.Errorf("create notification for chat %d: %w", chat.ChatID, createError)
		}
		if !created {
			continue
		}

		handleError := useCase.deliverNotification(
			ctx,
			event,
			routeContext,
			chat.ChatID,
			false,
			0,
		)
		if handleError != nil {
			useCase.logger.Warn(
				"failed to process notification delivery",
				"method",
				"NotifierUseCase.RouteAndSend",
				"alarm_event_id",
				event.ID.String(),
				"chat_id",
				chat.ChatID,
				"error",
				handleError,
			)
		}
	}

	if event.EventType == domain.AlarmEventRaised && isEscalationState(event.StateTo) {
		if createEscalationError := useCase.notificationRepository.CreateEscalation(
			ctx,
			event.ID,
			event.TagID,
			event.StateTo,
			time.Now().UTC().Add(useCase.escalationDelay),
		); createEscalationError != nil {
			useCase.logger.Warn(
				"failed to enqueue escalation",
				"method",
				"NotifierUseCase.RouteAndSend",
				"alarm_event_id",
				event.ID.String(),
				"tag_id",
				event.TagID.String(),
				"error",
				createEscalationError,
			)
		}
	}

	return nil
}

func (useCase *NotifierUseCase) RetryPending(ctx context.Context) error {
	if useCase.notificationRepository == nil {
		return errors.New("notification repository is not configured")
	}

	pending, listError := useCase.notificationRepository.ListPendingNotifications(
		ctx,
		useCase.maxAttempts,
		useCase.backoffBase,
		useCase.batchLimit,
	)
	if listError != nil {
		return fmt.Errorf("list pending notifications: %w", listError)
	}

	for _, item := range pending {
		routeContext, routeContextError := useCase.chatsRepository.GetAlarmRouteContext(
			ctx,
			item.Event.TagID,
		)
		if routeContextError != nil {
			useCase.logger.Warn(
				"failed to resolve route context for retry",
				"method",
				"NotifierUseCase.RetryPending",
				"alarm_event_id",
				item.AlarmEventID.String(),
				"tag_id",
				item.Event.TagID.String(),
				"error",
				routeContextError,
			)
			continue
		}

		if deliveryError := useCase.deliverNotification(
			ctx,
			item.Event,
			routeContext,
			item.ChatID,
			false,
			item.Attempt,
		); deliveryError != nil {
			useCase.logger.Warn(
				"failed retry delivery",
				"method",
				"NotifierUseCase.RetryPending",
				"alarm_event_id",
				item.AlarmEventID.String(),
				"chat_id",
				item.ChatID,
				"attempt",
				item.Attempt,
				"error",
				deliveryError,
			)
		}
	}

	return nil
}

func (useCase *NotifierUseCase) ProcessEscalations(ctx context.Context) error {
	if useCase.notificationRepository == nil {
		return errors.New("notification repository is not configured")
	}

	dueEscalations, listError := useCase.notificationRepository.ListDueEscalations(ctx, useCase.batchLimit)
	if listError != nil {
		return fmt.Errorf("list due escalations: %w", listError)
	}

	for _, escalation := range dueEscalations {
		acked, ackError := useCase.notificationRepository.IsTagAcknowledgedForState(
			ctx,
			escalation.TagID,
			escalation.State,
		)
		if ackError != nil {
			useCase.logger.Warn(
				"failed to check alarm acknowledgement",
				"method",
				"NotifierUseCase.ProcessEscalations",
				"escalation_id",
				escalation.ID.String(),
				"tag_id",
				escalation.TagID.String(),
				"error",
				ackError,
			)
			continue
		}

		if acked {
			if markError := useCase.notificationRepository.MarkEscalationProcessed(ctx, escalation.ID, true); markError != nil {
				useCase.logger.Warn(
					"failed to mark skipped escalation",
					"method",
					"NotifierUseCase.ProcessEscalations",
					"escalation_id",
					escalation.ID.String(),
					"error",
					markError,
				)
			}
			continue
		}

		event, eventError := useCase.notificationRepository.GetAlarmEventByID(ctx, escalation.AlarmEventID)
		if eventError != nil {
			useCase.logger.Warn(
				"failed to load escalation event",
				"method",
				"NotifierUseCase.ProcessEscalations",
				"escalation_id",
				escalation.ID.String(),
				"alarm_event_id",
				escalation.AlarmEventID.String(),
				"error",
				eventError,
			)
			_ = useCase.notificationRepository.MarkEscalationProcessed(ctx, escalation.ID, true)
			continue
		}

		routeContext, routeContextError := useCase.chatsRepository.GetAlarmRouteContext(ctx, event.TagID)
		if routeContextError != nil {
			useCase.logger.Warn(
				"failed to resolve escalation route context",
				"method",
				"NotifierUseCase.ProcessEscalations",
				"escalation_id",
				escalation.ID.String(),
				"tag_id",
				event.TagID.String(),
				"error",
				routeContextError,
			)
			continue
		}

		adminChats, chatsError := useCase.chatsRepository.ListAdminChatsForObject(
			ctx,
			routeContext.ObjectID,
		)
		if chatsError != nil {
			useCase.logger.Warn(
				"failed to resolve admin chats",
				"method",
				"NotifierUseCase.ProcessEscalations",
				"escalation_id",
				escalation.ID.String(),
				"object_id",
				routeContext.ObjectID.String(),
				"error",
				chatsError,
			)
			continue
		}

		for _, chat := range adminChats {
			created, createError := useCase.notificationRepository.CreateNotification(
				ctx,
				event.ID,
				chat.ChatID,
			)
			if createError != nil {
				useCase.logger.Warn(
					"failed to create escalation notification",
					"method",
					"NotifierUseCase.ProcessEscalations",
					"alarm_event_id",
					event.ID.String(),
					"chat_id",
					chat.ChatID,
					"error",
					createError,
				)
				continue
			}
			if !created {
				continue
			}

			if deliveryError := useCase.deliverNotification(
				ctx,
				event,
				routeContext,
				chat.ChatID,
				true,
				0,
			); deliveryError != nil {
				useCase.logger.Warn(
					"failed escalation delivery",
					"method",
					"NotifierUseCase.ProcessEscalations",
					"alarm_event_id",
					event.ID.String(),
					"chat_id",
					chat.ChatID,
					"error",
					deliveryError,
				)
			}
		}

		if markError := useCase.notificationRepository.MarkEscalationProcessed(ctx, escalation.ID, false); markError != nil {
			useCase.logger.Warn(
				"failed to mark escalation processed",
				"method",
				"NotifierUseCase.ProcessEscalations",
				"escalation_id",
				escalation.ID.String(),
				"error",
				markError,
			)
		}
	}

	return nil
}

func (useCase *NotifierUseCase) RunScheduler(ctx context.Context) error {
	ticker := time.NewTicker(useCase.pollInterval)
	defer ticker.Stop()

	if tickError := useCase.runScheduledWork(ctx); tickError != nil {
		useCase.logger.Warn(
			"notifier scheduler tick failed",
			"method",
			"NotifierUseCase.RunScheduler",
			"error",
			tickError,
		)
	}

	for {
		select {
		case <-ctx.Done():
			shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = useCase.runScheduledWork(shutdownContext)
			cancel()
			return nil
		case <-ticker.C:
			if tickError := useCase.runScheduledWork(ctx); tickError != nil {
				useCase.logger.Warn(
					"notifier scheduler tick failed",
					"method",
					"NotifierUseCase.RunScheduler",
					"error",
					tickError,
				)
			}
		}
	}
}

func (useCase *NotifierUseCase) runScheduledWork(ctx context.Context) error {
	if useCase.notificationRepository == nil {
		return errors.New("notification repository is not configured")
	}

	owner, lockError := useCase.notificationRepository.WithSchedulerLock(
		ctx,
		useCase.schedulerLock,
		func(callbackContext context.Context) error {
			if retryError := useCase.RetryPending(callbackContext); retryError != nil {
				return retryError
			}
			if escalationError := useCase.ProcessEscalations(callbackContext); escalationError != nil {
				return escalationError
			}
			return nil
		},
	)
	if lockError != nil {
		return lockError
	}

	useCase.logger.Debug(
		"notifier scheduler tick",
		"method",
		"NotifierUseCase.runScheduledWork",
		"owner",
		owner,
	)
	return nil
}

func (useCase *NotifierUseCase) deliverNotification(
	ctx context.Context,
	event domain.AlarmEvent,
	routeContext domain.AlarmRouteContext,
	chatID int64,
	escalated bool,
	currentAttempt int16,
) error {
	createdAt := time.Now()
	if useCase.dryRun {
		if markError := useCase.notificationRepository.MarkNotificationSkipped(
			ctx,
			event.ID,
			chatID,
			dryRunSkipReason,
		); markError != nil {
			return markError
		}
		useCase.logger.Info(
			"notifier.skipped",
			"method",
			"NotifierUseCase.deliverNotification",
			"chat_id",
			chatID,
			"alarm_event_id",
			event.ID.String(),
			"reason",
			dryRunSkipReason,
		)
		if useCase.auditUseCase != nil {
			useCase.auditUseCase.Record(context.Background(), domain.AuditEvent{
				Action: "notifier.dispatched",
				Target: domain.AuditTarget{
					Type: "alarm_notification",
					ID:   event.ID.String(),
				},
				Details: map[string]any{
					"chat_id": chatID,
					"attempt": currentAttempt + 1,
				},
				Result: domain.AuditResultSuccess,
			})
		}
		return nil
	}

	if useCase.notifierClient == nil {
		return errors.New("telegram notifier client is not configured")
	}
	if useCase.messageRenderer == nil {
		return errors.New("notifier message renderer is not configured")
	}

	messageText, renderError := useCase.messageRenderer.Render(
		event,
		routeContext,
		escalated,
		useCase.escalationDelay,
	)
	if renderError != nil {
		return fmt.Errorf("render notification message: %w", renderError)
	}

	sendError := useCase.notifierClient.SendMessage(
		ctx,
		chatID,
		messageText,
		telegramParseMode,
	)
	if sendError == nil {
		if markError := useCase.notificationRepository.MarkNotificationSent(ctx, event.ID, chatID); markError != nil {
			return markError
		}
		useCase.logger.Info(
			"notifier.sent",
			"method",
			"NotifierUseCase.deliverNotification",
			"chat_id",
			chatID,
			"alarm_event_id",
			event.ID.String(),
			"attempt",
			currentAttempt+1,
			"latency_ms",
			time.Since(createdAt).Milliseconds(),
		)
		return nil
	}

	errorText := strings.TrimSpace(sendError.Error())
	nextAttempt := int(currentAttempt) + 1
	if nextAttempt >= useCase.maxAttempts || domain.IsPermanentSendError(sendError) {
		if markError := useCase.notificationRepository.MarkNotificationFailed(
			ctx,
			event.ID,
			chatID,
			errorText,
		); markError != nil {
			return markError
		}
		useCase.logger.Warn(
			"notifier.failed",
			"method",
			"NotifierUseCase.deliverNotification",
			"chat_id",
			chatID,
			"alarm_event_id",
			event.ID.String(),
			"attempt",
			nextAttempt,
			"error",
			errorText,
		)
		return nil
	}

	var retryAfter *time.Duration
	if rateLimitError, ok := domain.IsRateLimitError(sendError); ok {
		retryAfter = &rateLimitError.RetryAfter
	}

	if markError := useCase.notificationRepository.MarkNotificationPendingRetry(
		ctx,
		event.ID,
		chatID,
		errorText,
		retryAfter,
	); markError != nil {
		return markError
	}

	useCase.logger.Warn(
		"notifier.failed",
		"method",
		"NotifierUseCase.deliverNotification",
		"chat_id",
		chatID,
		"alarm_event_id",
		event.ID.String(),
		"attempt",
		nextAttempt,
		"error",
		errorText,
	)

	return nil
}

func mapSeverity(event domain.AlarmEvent) int16 {
	switch event.EventType {
	case domain.AlarmEventAcked, domain.AlarmEventCleared:
		return 1
	}

	switch event.StateTo {
	case domain.AlarmStateHi, domain.AlarmStateLo:
		return 2
	case domain.AlarmStateHiHi, domain.AlarmStateLoLo:
		return 4
	case domain.AlarmStateCommLoss, domain.AlarmStateOffline, domain.AlarmStateBad, domain.AlarmStateUncertain:
		return 3
	default:
		return 1
	}
}

func isEscalationState(state domain.AlarmState) bool {
	return state == domain.AlarmStateHiHi || state == domain.AlarmStateLoLo
}
