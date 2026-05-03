package usecase

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/auditctx"
	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
	"github.com/EthernalFox/Controlitix/shared/authctx"
)

type AuditProducer interface {
	Publish(ctx context.Context, record domain.AuditRecord)
}

type AuditUseCase struct {
	producer AuditProducer
	logger   *slog.Logger
}

func NewAuditUseCase(producer AuditProducer, logger *slog.Logger) *AuditUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuditUseCase{producer: producer, logger: logger}
}

func (useCase *AuditUseCase) Record(ctx context.Context, event domain.AuditEvent) {
	if useCase == nil || useCase.producer == nil {
		return
	}

	principal, ok := authctx.FromContext(ctx)
	metadata := auditctx.RequestMetadataFromContext(ctx)

	result := strings.TrimSpace(event.Result)
	if result == "" {
		result = domain.AuditResultSuccess
	}

	record := domain.AuditRecord{
		Version:   1,
		Timestamp: time.Now().UTC(),
		Service:   "ms-viewer",
		Action:    strings.TrimSpace(event.Action),
		Target:    event.Target,
		Details:   copyDetails(event.Details),
		Result:    result,
		IP:        strings.TrimSpace(metadata.IP),
		UserAgent: strings.TrimSpace(metadata.UserAgent),
	}

	if requestID := strings.TrimSpace(metadata.RequestID); requestID != "" {
		if record.Details == nil {
			record.Details = make(map[string]any, 1)
		}
		record.Details["request_id"] = requestID
	}

	if ok {
		record.ActorRoles = append([]string(nil), principal.Roles...)
		subject := strings.TrimSpace(principal.Subject)
		if subject != "" {
			record.ActorID = &subject
		}
		username := strings.TrimSpace(principal.Username)
		if username == "" {
			username = subject
		}
		if username == "" {
			username = "unknown"
		}
		record.ActorUsername = username
	} else {
		record.ActorUsername = "system"
	}

	useCase.producer.Publish(ctx, record)
	useCase.logger.Info(
		"audit event queued",
		"method",
		"AuditUseCase.Record",
		"action",
		record.Action,
		"actor_id",
		record.ActorID,
		"result",
		record.Result,
	)
}

func copyDetails(source map[string]any) map[string]any {
	if len(source) == 0 {
		return nil
	}
	copyMap := make(map[string]any, len(source))
	for key, value := range source {
		copyMap[key] = value
	}
	return copyMap
}
