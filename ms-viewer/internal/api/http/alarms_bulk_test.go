package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
	"github.com/EthernalFox/Controlitix/shared/authctx"
)

func TestAcknowledgeBulkReturnsMultiStatus(t *testing.T) {
	tagAcked := uuid.New()
	tagNotActive := uuid.New()

	handler := NewAlarmsHandler(&stubAlarmUseCase{
		listFunc: func(_ context.Context, _ domain.AlarmListQuery) (domain.AlarmListResult, error) {
			return domain.AlarmListResult{}, nil
		},
		getFunc: func(_ context.Context, _ uuid.UUID) (domain.AlarmDetail, error) {
			return domain.AlarmDetail{}, nil
		},
		acknowledgeFunc: func(_ context.Context, _ uuid.UUID, _ string, _ *string) (domain.AlarmAcknowledgeResult, error) {
			return domain.AlarmAcknowledgeResult{}, nil
		},
		acknowledgeBulkFunc: func(_ context.Context, _ []uuid.UUID, actorID string, _ *string) (domain.AlarmBulkAcknowledgeResult, error) {
			state := domain.AlarmStateHi
			return domain.AlarmBulkAcknowledgeResult{
				Items: []domain.AlarmBulkAcknowledgeItemResult{
					{TagID: tagAcked, Status: domain.AlarmBulkAckStatusAcked, State: &state},
					{TagID: tagNotActive, Status: domain.AlarmBulkAckStatusNotActive},
				},
				AckedAt:  time.Now().UTC(),
				ActorID:  actorID,
				SuccessN: 1,
				FailedN:  1,
			}, nil
		},
	}, nil)

	router := chi.NewRouter()
	router.Post("/api/alarms/acknowledge", handler.AcknowledgeBulk)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/alarms/acknowledge",
		bytes.NewBufferString(`{"items":[{"tag_id":"`+tagAcked.String()+`"},{"tag_id":"`+tagNotActive.String()+`"}]}`),
	)
	request = request.WithContext(authctx.WithPrincipal(request.Context(), authctx.Principal{Subject: "operator-1"}))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusMultiStatus {
		t.Fatalf("expected status %d, got %d", http.StatusMultiStatus, response.Code)
	}

	var payload bulkAcknowledgeResponse
	if decodeError := json.NewDecoder(response.Body).Decode(&payload); decodeError != nil {
		t.Fatalf("decode response: %v", decodeError)
	}
	if payload.SuccessN != 1 || payload.FailedN != 1 {
		t.Fatalf("unexpected counters: success=%d failed=%d", payload.SuccessN, payload.FailedN)
	}
}

func TestAcknowledgeBulkReturnsOKOnFullSuccess(t *testing.T) {
	tagID := uuid.New()

	handler := NewAlarmsHandler(&stubAlarmUseCase{
		listFunc: func(_ context.Context, _ domain.AlarmListQuery) (domain.AlarmListResult, error) {
			return domain.AlarmListResult{}, nil
		},
		getFunc: func(_ context.Context, _ uuid.UUID) (domain.AlarmDetail, error) {
			return domain.AlarmDetail{}, nil
		},
		acknowledgeFunc: func(_ context.Context, _ uuid.UUID, _ string, _ *string) (domain.AlarmAcknowledgeResult, error) {
			return domain.AlarmAcknowledgeResult{}, nil
		},
		acknowledgeBulkFunc: func(_ context.Context, _ []uuid.UUID, actorID string, _ *string) (domain.AlarmBulkAcknowledgeResult, error) {
			state := domain.AlarmStateHi
			return domain.AlarmBulkAcknowledgeResult{
				Items: []domain.AlarmBulkAcknowledgeItemResult{
					{TagID: tagID, Status: domain.AlarmBulkAckStatusAcked, State: &state},
				},
				AckedAt:  time.Now().UTC(),
				ActorID:  actorID,
				SuccessN: 1,
				FailedN:  0,
			}, nil
		},
	}, nil)

	router := chi.NewRouter()
	router.Post("/api/alarms/acknowledge", handler.AcknowledgeBulk)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/alarms/acknowledge",
		bytes.NewBufferString(`{"items":[{"tag_id":"`+tagID.String()+`"}]}`),
	)
	request = request.WithContext(authctx.WithPrincipal(request.Context(), authctx.Principal{Subject: "operator-1"}))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}
