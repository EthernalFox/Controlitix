package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
	kafkainfra "github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/kafka"
	"github.com/EthernalFox/Controlitix/ms-auth/internal/infrastructure/repository"
)

const kafkaPublishTimeout = 2 * time.Second

type auditRepository interface {
	Insert(ctx context.Context, event domain.AuditEvent) error
}

type auditProducer interface {
	Publish(ctx context.Context, event domain.AuditEvent) error
}

type AuditService struct {
	repo     auditRepository
	producer auditProducer
	logger   *slog.Logger
}

func NewAuditService(
	repo *repository.AuditRepository,
	producer *kafkainfra.AuditProducer,
	logger *slog.Logger,
) *AuditService {
	if logger == nil {
		logger = slog.Default()
	}

	return &AuditService{
		repo:     repo,
		producer: producer,
		logger:   logger,
	}
}

func (service *AuditService) Record(ctx context.Context, event domain.AuditEvent) error {
	logger := service.logger
	if logger == nil {
		logger = slog.Default()
	}

	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}
	if event.Metadata == nil {
		event.Metadata = map[string]any{}
	}

	if err := service.repo.Insert(ctx, event); err != nil {
		return fmt.Errorf("insert audit event into database: %w", err)
	}

	if service.producer != nil {
		go func(copyEvent domain.AuditEvent) {
			publishContext, cancelPublish := context.WithTimeout(
				context.Background(),
				kafkaPublishTimeout,
			)
			defer cancelPublish()

			if publishError := service.producer.Publish(publishContext, copyEvent); publishError != nil {
				logger.Warn(
					"failed to publish audit event to kafka",
					"method",
					"AuditService.Record",
					"action",
					copyEvent.Action,
					"result",
					copyEvent.Result,
					"error",
					publishError,
				)
			}
		}(event)
	}

	return nil
}
