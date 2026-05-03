package usecase

import (
	"context"
	"testing"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/auditctx"
	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
	"github.com/EthernalFox/Controlitix/shared/authctx"
)

type auditProducerStub struct {
	records []domain.AuditRecord
}

func (stub *auditProducerStub) Publish(_ context.Context, record domain.AuditRecord) {
	stub.records = append(stub.records, record)
}

func TestAuditUseCaseRecordWithPrincipal(t *testing.T) {
	producer := &auditProducerStub{}
	useCase := NewAuditUseCase(producer, nil)

	ctx := context.Background()
	ctx = auditctx.WithRequestMetadata(ctx, auditctx.RequestMetadata{
		RequestID: "req-1",
		IP:        "10.0.0.1",
		UserAgent: "ua",
	})
	ctx = authctx.WithPrincipal(ctx, authctx.Principal{
		Subject:  "user-1",
		Username: "operator-1",
		Roles:    []string{"operator"},
	})

	useCase.Record(ctx, domain.AuditEvent{
		Action: "alarm.acknowledged",
		Target: domain.AuditTarget{Type: "alarm", ID: "tag-1"},
		Result: domain.AuditResultSuccess,
	})

	if len(producer.records) != 1 {
		t.Fatalf("expected one record")
	}
	record := producer.records[0]
	if record.ActorID == nil || *record.ActorID != "user-1" {
		t.Fatalf("unexpected actor id: %#v", record.ActorID)
	}
	if record.ActorUsername != "operator-1" {
		t.Fatalf("unexpected actor username: %s", record.ActorUsername)
	}
	if record.Details["request_id"] != "req-1" {
		t.Fatalf("expected request_id in details")
	}
}

func TestAuditUseCaseRecordSystem(t *testing.T) {
	producer := &auditProducerStub{}
	useCase := NewAuditUseCase(producer, nil)

	useCase.Record(context.Background(), domain.AuditEvent{
		Action: "alarm.raised",
		Target: domain.AuditTarget{Type: "alarm", ID: "tag-1"},
		Result: domain.AuditResultSuccess,
	})

	if len(producer.records) != 1 {
		t.Fatalf("expected one record")
	}
	record := producer.records[0]
	if record.ActorID != nil {
		t.Fatalf("expected nil actor id for system event")
	}
	if record.ActorUsername != "system" {
		t.Fatalf("expected system actor username")
	}
}
