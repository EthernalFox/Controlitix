package realtime

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type fakeHubTopicResolverRepository struct {
	topics map[string][]uuid.UUID
}

func (repository *fakeHubTopicResolverRepository) ResolveDiagramTagIDs(
	_ context.Context,
	diagramID uuid.UUID,
	_ int,
) ([]uuid.UUID, error) {
	return append(
		[]uuid.UUID(nil),
		repository.topics["diagram:"+diagramID.String()]...,
	), nil
}

func (repository *fakeHubTopicResolverRepository) ResolveObjectTagIDs(
	_ context.Context,
	objectID uuid.UUID,
	_ int,
) ([]uuid.UUID, error) {
	return append(
		[]uuid.UUID(nil),
		repository.topics["object:"+objectID.String()]...,
	), nil
}

func TestHubSubscriptionRefcountWithDirectAndGroupedTopics(t *testing.T) {
	diagramID := uuid.New()
	objectID := uuid.New()
	sharedTagID := uuid.New()
	diagramTagID := uuid.New()
	objectTagID := uuid.New()

	repository := &fakeHubTopicResolverRepository{
		topics: map[string][]uuid.UUID{
			"diagram:" + diagramID.String(): {sharedTagID, diagramTagID},
			"object:" + objectID.String():   {sharedTagID, objectTagID},
		},
	}
	topicResolver := NewTopicResolver(repository, NewMembershipCache(0), 10, nil)
	hub := NewHub(HubOptions{
		Debouncer:               NewDebouncer(0, time.Minute, nil),
		MaxSubscriptionsPerConn: 10,
		TopicResolver:           topicResolver,
	})
	connection := newTestConnection()
	hub.RegisterConnection(connection)

	hub.Subscribe(connection, []string{"tag:" + sharedTagID.String()})
	hub.Subscribe(connection, []string{"diagram:" + diagramID.String()})
	hub.Subscribe(connection, []string{"object:" + objectID.String()})

	if count := hub.SubscriptionCount(connection); count != 3 {
		t.Fatalf("expected 3 unique subscriptions, got %d", count)
	}

	hub.Unsubscribe(connection, []string{"tag:" + sharedTagID.String()})
	drainOutbox(connection.outbox)
	publishTag(hub, sharedTagID)
	if got := waitMessageType(t, connection.outbox, time.Second); got != ServerTypeValue {
		t.Fatalf("expected value message for shared tag after direct unsubscribe, got %s", got)
	}

	hub.Unsubscribe(connection, []string{"diagram:" + diagramID.String()})
	drainOutbox(connection.outbox)
	publishTag(hub, sharedTagID)
	if got := waitMessageType(t, connection.outbox, time.Second); got != ServerTypeValue {
		t.Fatalf("expected value message for shared tag after diagram unsubscribe, got %s", got)
	}

	hub.Unsubscribe(connection, []string{"object:" + objectID.String()})
	drainOutbox(connection.outbox)
	publishTag(hub, sharedTagID)

	select {
	case <-connection.outbox:
		t.Fatal("unexpected message after unsubscribing from all shared tag paths")
	case <-time.After(200 * time.Millisecond):
	}
}

func publishTag(hub *Hub, tagID uuid.UUID) {
	value := 10.0
	hub.Publish(TagValueEvent{
		TagID: tagID,
		Record: domain.IngestRecord{
			TagID:     tagID,
			Timestamp: time.Now().UTC(),
			Value:     &value,
			Quality:   domain.QualityOK,
		},
	})
}

func drainOutbox(outbox chan outboundMessage) {
	for {
		select {
		case <-outbox:
		default:
			return
		}
	}
}

