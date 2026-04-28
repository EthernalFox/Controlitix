package realtime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

func TestDebouncerEmitsAtMostOneValuePerWindow(t *testing.T) {
	debouncer := NewDebouncer(100*time.Millisecond, time.Minute, nil)
	connection := newTestConnection()
	tagID := uuid.New()

	baseTimestamp := time.Now().UTC()
	for i := 0; i < 10; i++ {
		value := float64(i)
		debouncer.Publish(connection, domain.IngestRecord{
			TagID:     tagID,
			Timestamp: baseTimestamp.Add(time.Duration(i) * time.Millisecond),
			Value:     &value,
			Quality:   domain.QualityOK,
		})
	}

	valueMessages := 0
	deadline := time.After(260 * time.Millisecond)
	for {
		select {
		case message := <-connection.outbox:
			var envelope struct {
				T string `json:"t"`
			}
			if unmarshalError := json.Unmarshal(message.payload, &envelope); unmarshalError != nil {
				t.Fatalf("failed to decode message: %v", unmarshalError)
			}
			if envelope.T == ServerTypeValue {
				valueMessages++
			}
		case <-deadline:
			if valueMessages > 2 {
				t.Fatalf("expected at most 2 value messages for debounce window, got %d", valueMessages)
			}
			if valueMessages == 0 {
				t.Fatal("expected at least one value message")
			}
			debouncer.StopConnection(connection)
			return
		}
	}
}

func TestDebouncerSendsHeartbeatSnapshot(t *testing.T) {
	debouncer := NewDebouncer(100*time.Millisecond, 120*time.Millisecond, nil)
	connection := newTestConnection()
	tagID := uuid.New()
	value := 10.5

	record := domain.IngestRecord{
		TagID:     tagID,
		Timestamp: time.Now().UTC(),
		Value:     &value,
		Quality:   domain.QualityOK,
	}

	debouncer.Touch(connection, record)

	select {
	case message := <-connection.outbox:
		var payload struct {
			T      string `json:"t"`
			Reason string `json:"reason"`
		}
		if unmarshalError := json.Unmarshal(message.payload, &payload); unmarshalError != nil {
			t.Fatalf("failed to decode message: %v", unmarshalError)
		}
		if payload.T != ServerTypeSnapshot {
			t.Fatalf("expected snapshot heartbeat, got %s", payload.T)
		}
		if payload.Reason != string(SnapshotReasonHeartbeat) {
			t.Fatalf("expected heartbeat reason, got %s", payload.Reason)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for heartbeat snapshot")
	}

	debouncer.StopConnection(connection)
}
