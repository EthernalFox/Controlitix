package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
	"github.com/EthernalFox/Controlitix/shared/authctx"
)

type stubAlarmUseCase struct {
	listFunc            func(ctx context.Context, query domain.AlarmListQuery) (domain.AlarmListResult, error)
	getFunc             func(ctx context.Context, tagID uuid.UUID) (domain.AlarmDetail, error)
	acknowledgeFunc     func(ctx context.Context, tagID uuid.UUID, actorID string, note *string) (domain.AlarmAcknowledgeResult, error)
	acknowledgeBulkFunc func(ctx context.Context, tagIDs []uuid.UUID, actorID string, note *string) (domain.AlarmBulkAcknowledgeResult, error)
}

func (stub *stubAlarmUseCase) ListAlarms(
	ctx context.Context,
	query domain.AlarmListQuery,
) (domain.AlarmListResult, error) {
	return stub.listFunc(ctx, query)
}

func (stub *stubAlarmUseCase) GetAlarm(
	ctx context.Context,
	tagID uuid.UUID,
) (domain.AlarmDetail, error) {
	return stub.getFunc(ctx, tagID)
}

func (stub *stubAlarmUseCase) Acknowledge(
	ctx context.Context,
	tagID uuid.UUID,
	actorID string,
	note *string,
) (domain.AlarmAcknowledgeResult, error) {
	return stub.acknowledgeFunc(ctx, tagID, actorID, note)
}

func (stub *stubAlarmUseCase) AcknowledgeBulk(
	ctx context.Context,
	tagIDs []uuid.UUID,
	actorID string,
	note *string,
) (domain.AlarmBulkAcknowledgeResult, error) {
	if stub.acknowledgeBulkFunc == nil {
		return domain.AlarmBulkAcknowledgeResult{}, nil
	}
	return stub.acknowledgeBulkFunc(ctx, tagIDs, actorID, note)
}

func TestGetAlarmsParsesFilters(t *testing.T) {
	captured := domain.AlarmListQuery{}
	objectID := uuid.New()
	handler := NewAlarmsHandler(&stubAlarmUseCase{
		listFunc: func(_ context.Context, query domain.AlarmListQuery) (domain.AlarmListResult, error) {
			captured = query
			return domain.AlarmListResult{Items: []domain.AlarmStateRecord{}, Total: 0, Limit: query.Limit, Offset: query.Offset}, nil
		},
		getFunc: func(_ context.Context, _ uuid.UUID) (domain.AlarmDetail, error) {
			return domain.AlarmDetail{}, nil
		},
		acknowledgeFunc: func(_ context.Context, _ uuid.UUID, _ string, _ *string) (domain.AlarmAcknowledgeResult, error) {
			return domain.AlarmAcknowledgeResult{}, nil
		},
	}, nil)

	router := chi.NewRouter()
	router.Get("/api/alarms", handler.GetAlarms)

	request := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/alarms?status=acked&severity=alarm&object_id=%s&limit=100&offset=20", objectID.String()),
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if captured.Status != domain.AlarmListStatusAcked {
		t.Fatalf("expected status acked, got %s", captured.Status)
	}
	if captured.Severity != domain.AlarmSeverityAlarm {
		t.Fatalf("expected severity alarm, got %s", captured.Severity)
	}
	if captured.ObjectID != objectID {
		t.Fatalf("expected object id %s, got %s", objectID.String(), captured.ObjectID.String())
	}
	if captured.Limit != 100 || captured.Offset != 20 {
		t.Fatalf("unexpected pagination: limit=%d offset=%d", captured.Limit, captured.Offset)
	}
}

func TestAcknowledgeReturnsConflictForAlreadyAcked(t *testing.T) {
	tagID := uuid.New()
	handler := NewAlarmsHandler(&stubAlarmUseCase{
		listFunc: func(_ context.Context, _ domain.AlarmListQuery) (domain.AlarmListResult, error) {
			return domain.AlarmListResult{}, nil
		},
		getFunc: func(_ context.Context, _ uuid.UUID) (domain.AlarmDetail, error) {
			return domain.AlarmDetail{}, nil
		},
		acknowledgeFunc: func(_ context.Context, _ uuid.UUID, _ string, _ *string) (domain.AlarmAcknowledgeResult, error) {
			return domain.AlarmAcknowledgeResult{}, domain.ErrAlarmAlreadyAcked
		},
	}, nil)

	router := chi.NewRouter()
	router.Post("/api/alarms/{tagId}/acknowledge", handler.Acknowledge)

	request := httptest.NewRequest(http.MethodPost, "/api/alarms/"+tagID.String()+"/acknowledge", bytes.NewBufferString(`{}`))
	request = request.WithContext(authctx.WithPrincipal(request.Context(), authctx.Principal{
		Subject: "user-1",
	}))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, response.Code)
	}

	var payload Problem
	if decodeError := json.NewDecoder(response.Body).Decode(&payload); decodeError != nil {
		t.Fatalf("decode problem: %v", decodeError)
	}
	if payload.Type != "/errors/alarms/already-acked" {
		t.Fatalf("expected /errors/alarms/already-acked, got %s", payload.Type)
	}
}

func TestAcknowledgeUsesPrincipalSubject(t *testing.T) {
	tagID := uuid.New()
	var capturedActorID string
	handler := NewAlarmsHandler(&stubAlarmUseCase{
		listFunc: func(_ context.Context, _ domain.AlarmListQuery) (domain.AlarmListResult, error) {
			return domain.AlarmListResult{}, nil
		},
		getFunc: func(_ context.Context, _ uuid.UUID) (domain.AlarmDetail, error) {
			return domain.AlarmDetail{}, nil
		},
		acknowledgeFunc: func(
			_ context.Context,
			receivedTagID uuid.UUID,
			actorID string,
			note *string,
		) (domain.AlarmAcknowledgeResult, error) {
			capturedActorID = actorID
			noteValue := "ack"
			if note != nil {
				noteValue = *note
			}
			eventActorID := actorID
			return domain.AlarmAcknowledgeResult{
				TagID: receivedTagID,
				State: domain.AlarmStateHi,
				Ack: domain.AlarmAck{
					State:   domain.AlarmStateHi,
					ActorID: actorID,
					Note:    &noteValue,
					AckedAt: time.Now().UTC(),
				},
				Event: domain.AlarmEvent{
					TagID:   receivedTagID,
					ActorID: &eventActorID,
				},
			}, nil
		},
	}, nil)

	router := chi.NewRouter()
	router.Post("/api/alarms/{tagId}/acknowledge", handler.Acknowledge)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/alarms/"+tagID.String()+"/acknowledge",
		bytes.NewBufferString(`{"note":"manual ack"}`),
	)
	request = request.WithContext(authctx.WithPrincipal(request.Context(), authctx.Principal{
		Subject: "operator-1",
	}))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if capturedActorID != "operator-1" {
		t.Fatalf("expected actor_id operator-1, got %s", capturedActorID)
	}
}
