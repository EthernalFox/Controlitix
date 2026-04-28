package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type fakeTopicMembershipRepository struct {
	diagramByID map[uuid.UUID][]uuid.UUID
	objectByID  map[uuid.UUID][]uuid.UUID
	errByID     map[uuid.UUID]error

	diagramCalls int
	objectCalls  int
}

func (repository *fakeTopicMembershipRepository) ResolveDiagramTagIDs(
	_ context.Context,
	diagramID uuid.UUID,
	_ int,
) ([]uuid.UUID, error) {
	repository.diagramCalls++
	if err := repository.errByID[diagramID]; err != nil {
		return nil, err
	}

	return append([]uuid.UUID(nil), repository.diagramByID[diagramID]...), nil
}

func (repository *fakeTopicMembershipRepository) ResolveObjectTagIDs(
	_ context.Context,
	objectID uuid.UUID,
	_ int,
) ([]uuid.UUID, error) {
	repository.objectCalls++
	if err := repository.errByID[objectID]; err != nil {
		return nil, err
	}

	return append([]uuid.UUID(nil), repository.objectByID[objectID]...), nil
}

func TestTopicResolverResolveDiagramAndObjectTopics(t *testing.T) {
	diagramID := uuid.New()
	objectID := uuid.New()
	firstTagID := uuid.New()
	secondTagID := uuid.New()
	thirdTagID := uuid.New()

	repository := &fakeTopicMembershipRepository{
		diagramByID: map[uuid.UUID][]uuid.UUID{
			diagramID: {firstTagID, secondTagID, firstTagID},
		},
		objectByID: map[uuid.UUID][]uuid.UUID{
			objectID: {secondTagID, thirdTagID},
		},
		errByID: map[uuid.UUID]error{},
	}
	resolver := NewTopicResolver(repository, NewMembershipCache(0), 10, nil)

	diagramTopic := "diagram:" + diagramID.String()
	objectTopic := "object:" + objectID.String()

	diagramTags, diagramError := resolver.ResolveTopic(context.Background(), diagramTopic)
	if diagramError != nil {
		t.Fatalf("ResolveTopic diagram returned error: %v", diagramError)
	}
	assertUUIDSetEqual(t, diagramTags, []uuid.UUID{firstTagID, secondTagID})

	objectTags, objectError := resolver.ResolveTopic(context.Background(), objectTopic)
	if objectError != nil {
		t.Fatalf("ResolveTopic object returned error: %v", objectError)
	}
	assertUUIDSetEqual(t, objectTags, []uuid.UUID{secondTagID, thirdTagID})
}

func TestTopicResolverUsesCache(t *testing.T) {
	diagramID := uuid.New()
	tagID := uuid.New()

	repository := &fakeTopicMembershipRepository{
		diagramByID: map[uuid.UUID][]uuid.UUID{
			diagramID: {tagID},
		},
		objectByID: map[uuid.UUID][]uuid.UUID{},
		errByID:    map[uuid.UUID]error{},
	}
	resolver := NewTopicResolver(repository, NewMembershipCache(0), 10, nil)
	topic := "diagram:" + diagramID.String()

	_, firstError := resolver.ResolveTopic(context.Background(), topic)
	if firstError != nil {
		t.Fatalf("first ResolveTopic returned error: %v", firstError)
	}
	_, secondError := resolver.ResolveTopic(context.Background(), topic)
	if secondError != nil {
		t.Fatalf("second ResolveTopic returned error: %v", secondError)
	}

	if repository.diagramCalls != 1 {
		t.Fatalf("expected one repository call, got %d", repository.diagramCalls)
	}
}

func TestTopicResolverLimitExceeded(t *testing.T) {
	objectID := uuid.New()
	repository := &fakeTopicMembershipRepository{
		diagramByID: map[uuid.UUID][]uuid.UUID{},
		objectByID: map[uuid.UUID][]uuid.UUID{
			objectID: {uuid.New(), uuid.New(), uuid.New()},
		},
		errByID: map[uuid.UUID]error{},
	}
	resolver := NewTopicResolver(repository, NewMembershipCache(0), 2, nil)

	_, resolveError := resolver.ResolveTopic(
		context.Background(),
		"object:"+objectID.String(),
	)
	if !errors.Is(resolveError, ErrTopicResolveLimit) {
		t.Fatalf("expected ErrTopicResolveLimit, got %v", resolveError)
	}
}

func TestTopicResolverNotFound(t *testing.T) {
	diagramID := uuid.New()
	repository := &fakeTopicMembershipRepository{
		diagramByID: map[uuid.UUID][]uuid.UUID{},
		objectByID:  map[uuid.UUID][]uuid.UUID{},
		errByID: map[uuid.UUID]error{
			diagramID: domain.ErrNotFound,
		},
	}
	resolver := NewTopicResolver(repository, NewMembershipCache(0), 10, nil)

	_, resolveError := resolver.ResolveTopic(
		context.Background(),
		"diagram:"+diagramID.String(),
	)
	if !errors.Is(resolveError, domain.ErrNotFound) {
		t.Fatalf("expected domain.ErrNotFound, got %v", resolveError)
	}
}

func TestTopicResolverInvalidateByConfigEvent(t *testing.T) {
	diagramID := uuid.New()
	tagID := uuid.New()
	repository := &fakeTopicMembershipRepository{
		diagramByID: map[uuid.UUID][]uuid.UUID{
			diagramID: {tagID},
		},
		objectByID: map[uuid.UUID][]uuid.UUID{},
		errByID:    map[uuid.UUID]error{},
	}
	resolver := NewTopicResolver(repository, NewMembershipCache(0), 10, nil)

	topic := "diagram:" + diagramID.String()
	_, resolveError := resolver.ResolveTopic(context.Background(), topic)
	if resolveError != nil {
		t.Fatalf("ResolveTopic returned error: %v", resolveError)
	}

	payload, _ := json.Marshal(map[string]any{
		"diagram_id": diagramID.String(),
	})
	affected := resolver.InvalidateByConfigEvent(domain.ConfigChangedEvent{
		EntityType: "figure",
		EntityID:   uuid.NewString(),
		Operation:  "updated",
		Payload:    payload,
	})

	if len(affected) != 1 || affected[0] != topic {
		t.Fatalf("unexpected affected topics: %v", affected)
	}
}

func assertUUIDSetEqual(t *testing.T, actual []uuid.UUID, expected []uuid.UUID) {
	t.Helper()

	actualSet := make(map[uuid.UUID]struct{}, len(actual))
	for _, tagID := range actual {
		actualSet[tagID] = struct{}{}
	}
	expectedSet := make(map[uuid.UUID]struct{}, len(expected))
	for _, tagID := range expected {
		expectedSet[tagID] = struct{}{}
	}

	if len(actualSet) != len(expectedSet) {
		t.Fatalf("set size mismatch: actual=%d expected=%d", len(actualSet), len(expectedSet))
	}
	for tagID := range expectedSet {
		if _, exists := actualSet[tagID]; !exists {
			t.Fatalf("missing expected tag id: %s", tagID.String())
		}
	}
}

