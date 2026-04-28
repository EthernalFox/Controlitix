package realtime

import (
	"encoding/json"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

func TestHubSubscribePublishAndUnsubscribe(t *testing.T) {
	debouncer := NewDebouncer(20*time.Millisecond, time.Minute, nil)
	hub := NewHub(HubOptions{
		Debouncer:               debouncer,
		MaxSubscriptionsPerConn: 10,
		MaxConnectionsPerUser:   5,
	})

	connection := newTestConnection()
	hub.RegisterConnection(connection)

	tagID := uuid.New()
	topic := "tag:" + tagID.String()

	subscribeResult := hub.Subscribe(connection, []string{topic})
	if len(subscribeResult.InvalidTopics) != 0 {
		t.Fatalf("expected no invalid topics, got %v", subscribeResult.InvalidTopics)
	}
	if len(subscribeResult.SubscribedTopics) != 1 {
		t.Fatalf("expected one subscribed topic, got %d", len(subscribeResult.SubscribedTopics))
	}

	now := time.Now().UTC()
	value := 42.5
	hub.Publish(TagValueEvent{
		TagID: tagID,
		Record: domain.IngestRecord{
			TagID:     tagID,
			Timestamp: now,
			Value:     &value,
			Quality:   domain.QualityOK,
		},
	})

	messageType := waitMessageType(t, connection.outbox, time.Second)
	if messageType != ServerTypeValue {
		t.Fatalf("expected value message, got %s", messageType)
	}

	unsubscribeResult := hub.Unsubscribe(connection, []string{topic})
	if len(unsubscribeResult.UnsubscribedTopics) != 1 {
		t.Fatalf("expected one unsubscribed topic, got %d", len(unsubscribeResult.UnsubscribedTopics))
	}

	hub.Publish(TagValueEvent{
		TagID: tagID,
		Record: domain.IngestRecord{
			TagID:     tagID,
			Timestamp: now.Add(time.Second),
			Value:     &value,
			Quality:   domain.QualityOK,
		},
	})

	select {
	case <-connection.outbox:
		t.Fatal("unexpected message after unsubscribe")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestHubUnregisterConnectionClearsDebouncerState(t *testing.T) {
	debouncer := NewDebouncer(20*time.Millisecond, time.Minute, nil)
	hub := NewHub(HubOptions{Debouncer: debouncer})
	connection := newTestConnection()
	hub.RegisterConnection(connection)

	tagIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	topics := make([]string, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		topics = append(topics, "tag:"+tagID.String())
	}

	hub.Subscribe(connection, topics)
	hub.UnregisterConnection(connection)

	if debouncer.StateCount() != 0 {
		t.Fatalf("expected no debouncer states, got %d", debouncer.StateCount())
	}
	if count := hub.SubscriptionCount(connection); count != 0 {
		t.Fatalf("expected no subscriptions for connection, got %d", count)
	}
}

func TestHubConcurrentSubscribeAndPublish(t *testing.T) {
	debouncer := NewDebouncer(5*time.Millisecond, time.Minute, nil)
	hub := NewHub(HubOptions{
		Debouncer:               debouncer,
		MaxSubscriptionsPerConn: 200,
	})

	const subscribers = 25
	tagID := uuid.New()
	topic := "tag:" + tagID.String()

	connections := make([]*Connection, 0, subscribers)
	for i := 0; i < subscribers; i++ {
		connection := newTestConnection()
		hub.RegisterConnection(connection)
		connections = append(connections, connection)
	}

	var subscribeGroup sync.WaitGroup
	for _, connection := range connections {
		subscribeGroup.Add(1)
		go func(connection *Connection) {
			defer subscribeGroup.Done()
			hub.Subscribe(connection, []string{topic})
		}(connection)
	}
	subscribeGroup.Wait()

	var publishGroup sync.WaitGroup
	for i := 0; i < 100; i++ {
		publishGroup.Add(1)
		go func(index int) {
			defer publishGroup.Done()
			value := float64(index)
			hub.Publish(TagValueEvent{
				TagID: tagID,
				Record: domain.IngestRecord{
					TagID:     tagID,
					Timestamp: time.Now().UTC(),
					Value:     &value,
					Quality:   domain.QualityOK,
				},
			})
		}(i)
	}
	publishGroup.Wait()

	for _, connection := range connections {
		hub.UnregisterConnection(connection)
	}
}

func TestHubBroadcastConfigChanged(t *testing.T) {
	hub := NewHub(HubOptions{Debouncer: NewDebouncer(20 * time.Millisecond, time.Minute, nil)})
	first := newTestConnection()
	second := newTestConnection()
	hub.RegisterConnection(first)
	hub.RegisterConnection(second)

	hub.BroadcastConfigChanged(domain.ConfigChangedEvent{
		EntityType: "diagram",
		EntityID:   uuid.NewString(),
		Operation:  "published",
		Timestamp:  time.Now().UTC(),
	})

	if messageType := waitMessageType(t, first.outbox, time.Second); messageType != ServerTypeConfigChanged {
		t.Fatalf("expected config.changed for first connection, got %s", messageType)
	}
	if messageType := waitMessageType(t, second.outbox, time.Second); messageType != ServerTypeConfigChanged {
		t.Fatalf("expected config.changed for second connection, got %s", messageType)
	}
}

func TestHubUnsubscribeTagFromAll(t *testing.T) {
	hub := NewHub(HubOptions{Debouncer: NewDebouncer(20 * time.Millisecond, time.Minute, nil)})
	first := newTestConnection()
	second := newTestConnection()
	hub.RegisterConnection(first)
	hub.RegisterConnection(second)

	tagID := uuid.New()
	topic := "tag:" + tagID.String()

	hub.Subscribe(first, []string{topic})
	hub.Subscribe(second, []string{topic})
	hub.UnsubscribeTagFromAll(tagID)

	value := 1.0
	hub.Publish(TagValueEvent{
		TagID: tagID,
		Record: domain.IngestRecord{
			TagID:     tagID,
			Timestamp: time.Now().UTC(),
			Value:     &value,
			Quality:   domain.QualityOK,
		},
	})

	select {
	case <-first.outbox:
		t.Fatal("unexpected message for first connection after unsubscribe-all")
	case <-time.After(200 * time.Millisecond):
	}
	select {
	case <-second.outbox:
		t.Fatal("unexpected message for second connection after unsubscribe-all")
	case <-time.After(200 * time.Millisecond):
	}
}

func newTestConnection() *Connection {
	return &Connection{
		outbox:   make(chan outboundMessage, 64),
		done:     make(chan struct{}),
		openedAt: time.Now().UTC(),
		logger:   slog.Default(),
	}
}

func waitMessageType(
	testingContext *testing.T,
	messages <-chan outboundMessage,
	timeout time.Duration,
) string {
	testingContext.Helper()

	select {
	case message := <-messages:
		var envelope struct {
			T string `json:"t"`
		}
		if unmarshalError := json.Unmarshal(message.payload, &envelope); unmarshalError != nil {
			testingContext.Fatalf("failed to decode message: %v", unmarshalError)
		}
		return envelope.T
	case <-time.After(timeout):
		testingContext.Fatal("timed out waiting for realtime message")
		return ""
	}
}
