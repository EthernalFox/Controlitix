package realtime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

var (
	ErrTopicResolveLimit = errors.New("topic resolve limit exceeded")
)

type TopicMembershipRepository interface {
	ResolveDiagramTagIDs(ctx context.Context, diagramID uuid.UUID, limit int) ([]uuid.UUID, error)
	ResolveObjectTagIDs(ctx context.Context, objectID uuid.UUID, limit int) ([]uuid.UUID, error)
}

type TopicResolver struct {
	repository    TopicMembershipRepository
	cache         *MembershipCache
	resolveLimit  int
	logger        *slog.Logger
}

func NewTopicResolver(
	repository TopicMembershipRepository,
	cache *MembershipCache,
	resolveLimit int,
	logger *slog.Logger,
) *TopicResolver {
	if logger == nil {
		logger = slog.Default()
	}
	if resolveLimit <= 0 {
		resolveLimit = 500
	}
	if cache == nil {
		cache = NewMembershipCache(0)
	}

	return &TopicResolver{
		repository:   repository,
		cache:        cache,
		resolveLimit: resolveLimit,
		logger:       logger,
	}
}

func (resolver *TopicResolver) ResolveTopic(ctx context.Context, topic string) ([]uuid.UUID, error) {
	if resolver == nil {
		return nil, fmt.Errorf("resolver is not configured")
	}

	topic = normalizeTopicKey(topic)
	if topic == "" {
		return nil, fmt.Errorf("topic is empty")
	}

	prefix, entityID, parseError := ParseGroupTopic(topic)
	if parseError != nil {
		return nil, parseError
	}

	if cached, exists := resolver.cache.Get(topic); exists {
		return cached, nil
	}

	if resolver.repository == nil {
		return nil, fmt.Errorf("topic membership repository is not configured")
	}

	var (
		tagIDs []uuid.UUID
		resolveError error
	)

	switch prefix {
	case TopicPrefixDiagram:
		tagIDs, resolveError = resolver.repository.ResolveDiagramTagIDs(ctx, entityID, resolver.resolveLimit+1)
	case TopicPrefixObject:
		tagIDs, resolveError = resolver.repository.ResolveObjectTagIDs(ctx, entityID, resolver.resolveLimit+1)
	default:
		resolveError = fmt.Errorf("unsupported topic prefix")
	}
	if resolveError != nil {
		if errors.Is(resolveError, domain.ErrNotFound) {
			return nil, resolveError
		}
		return nil, fmt.Errorf("resolve topic membership: %w", resolveError)
	}

	if len(tagIDs) > resolver.resolveLimit {
		return nil, ErrTopicResolveLimit
	}

	resolved := uniqueUUIDs(tagIDs)
	resolver.cache.Set(topic, resolved)
	return resolved, nil
}

func (resolver *TopicResolver) Invalidate(entityType string, entityID string) []string {
	if resolver == nil || resolver.cache == nil {
		return nil
	}

	return resolver.cache.Invalidate(entityType, entityID)
}

func (resolver *TopicResolver) InvalidateByTag(tagID string) []string {
	if resolver == nil || resolver.cache == nil {
		return nil
	}

	return resolver.cache.InvalidateByTag(tagID)
}

func (resolver *TopicResolver) InvalidateByConfigEvent(event domain.ConfigChangedEvent) []string {
	if resolver == nil {
		return nil
	}

	entityType := strings.ToLower(strings.TrimSpace(event.EntityType))
	entityID := strings.TrimSpace(event.EntityID)
	if entityType == "" {
		return nil
	}

	switch entityType {
	case "diagram":
		return resolver.Invalidate(TopicPrefixDiagram, entityID)
	case "figure":
		payloadDiagramID := extractStringFromPayload(event.Payload, "diagram_id")
		if payloadDiagramID == "" {
			return nil
		}
		return resolver.Invalidate(TopicPrefixDiagram, payloadDiagramID)
	case "object":
		return resolver.Invalidate(TopicPrefixObject, entityID)
	case "tag":
		return resolver.InvalidateByTag(entityID)
	case "device":
		payloadObjectID := extractStringFromPayload(event.Payload, "object_id")
		if payloadObjectID == "" {
			return nil
		}
		return resolver.Invalidate(TopicPrefixObject, payloadObjectID)
	default:
		return nil
	}
}
