package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
)

type auditRepositoryMock struct {
	insert func(ctx context.Context, event domain.AuditEvent) error
}

func (m *auditRepositoryMock) Insert(ctx context.Context, event domain.AuditEvent) error {
	return m.insert(ctx, event)
}

type auditProducerMock struct {
	publish func(ctx context.Context, event domain.AuditEvent) error
}

func (m *auditProducerMock) Publish(ctx context.Context, event domain.AuditEvent) error {
	return m.publish(ctx, event)
}

func TestAuditServiceRecordWritesToDBAndKafka(t *testing.T) {
	t.Parallel()

	dbCalled := false
	kafkaCalled := make(chan struct{}, 1)
	service := &AuditService{
		repo: &auditRepositoryMock{
			insert: func(ctx context.Context, event domain.AuditEvent) error {
				dbCalled = true
				return nil
			},
		},
		producer: &auditProducerMock{
			publish: func(ctx context.Context, event domain.AuditEvent) error {
				kafkaCalled <- struct{}{}
				return nil
			},
		},
	}

	err := service.Record(context.Background(), domain.AuditEvent{
		Action: "login.success",
		Result: "success",
	})
	if err != nil {
		t.Fatalf("record audit event: %v", err)
	}
	if !dbCalled {
		t.Fatal("expected audit repository insert call")
	}

	select {
	case <-kafkaCalled:
	case <-time.After(time.Second):
		t.Fatal("expected kafka publish call")
	}
}

func TestAuditServiceRecordReturnsDBError(t *testing.T) {
	t.Parallel()

	service := &AuditService{
		repo: &auditRepositoryMock{
			insert: func(ctx context.Context, event domain.AuditEvent) error {
				return errors.New("database down")
			},
		},
	}

	err := service.Record(context.Background(), domain.AuditEvent{
		Action: "login.failure",
		Result: "failure",
	})
	if err == nil {
		t.Fatal("expected error when audit repository insert fails")
	}
}
