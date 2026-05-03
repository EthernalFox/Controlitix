package realtime

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type LastValueProvider interface {
	GetLastValue(ctx context.Context, tagID uuid.UUID) (domain.IngestRecord, bool, error)
}

type TagValueEvent struct {
	TagID   uuid.UUID
	Value   *float64
	Quality domain.Quality
	Record  domain.IngestRecord
}

type HubOptions struct {
	Debouncer               *Debouncer
	LastValueProvider       LastValueProvider
	MaxSubscriptionsPerConn int
	MaxConnectionsPerUser   int
	AlarmSnapshotProvider   AlarmSnapshotProvider
	TopicResolver           *TopicResolver
	Logger                  *slog.Logger
}

type AlarmSnapshotProvider interface {
	GetActiveAlarmSnapshot(ctx context.Context) ([]domain.AlarmStateRecord, error)
}

type Hub struct {
	mutex sync.RWMutex

	subscriptions           map[uuid.UUID]map[*Connection]struct{}
	connectionSubscriptions map[*Connection]map[uuid.UUID]struct{}
	connectionGroups        map[*Connection]*connectionSubscriptionGroups
	groupTopicSubscribers   map[string]map[*Connection]struct{}

	alarmsSubscribers map[*Connection]struct{}
	connectionsByUser map[string][]*Connection
	lastByTag         map[uuid.UUID]domain.IngestRecord

	debouncer             *Debouncer
	lastValueProvider     LastValueProvider
	maxSubscriptionsPerConn int
	maxConnectionsPerUser int
	alarmSnapshotProvider AlarmSnapshotProvider
	topicResolver         *TopicResolver
	logger                *slog.Logger
}

type SubscribeResult struct {
	SubscribedTopics []string
	SubscribedTagIDs []uuid.UUID
	InvalidTopics    []string
	LimitedTopics    []string
	SubscribedAlarms bool
}

type UnsubscribeResult struct {
	UnsubscribedTopics []string
	InvalidTopics      []string
	UnsubscribedAlarms bool
}

func NewHub(options HubOptions) *Hub {
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}
	debouncer := options.Debouncer
	if debouncer == nil {
		debouncer = NewDebouncer(0, 0, logger)
	}

	return &Hub{
		subscriptions:           make(map[uuid.UUID]map[*Connection]struct{}),
		connectionSubscriptions: make(map[*Connection]map[uuid.UUID]struct{}),
		connectionGroups:        make(map[*Connection]*connectionSubscriptionGroups),
		groupTopicSubscribers:   make(map[string]map[*Connection]struct{}),
		alarmsSubscribers:       make(map[*Connection]struct{}),
		connectionsByUser:       make(map[string][]*Connection),
		lastByTag:               make(map[uuid.UUID]domain.IngestRecord),
		debouncer:               debouncer,
		lastValueProvider:       options.LastValueProvider,
		maxSubscriptionsPerConn: options.MaxSubscriptionsPerConn,
		maxConnectionsPerUser:   options.MaxConnectionsPerUser,
		alarmSnapshotProvider:   options.AlarmSnapshotProvider,
		topicResolver:           options.TopicResolver,
		logger:                  logger,
	}
}

func (hub *Hub) RegisterConnection(connection *Connection) *Connection {
	if connection == nil {
		return nil
	}

	principalSubject := strings.TrimSpace(connection.Subject())

	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	hub.connectionSubscriptions[connection] = make(map[uuid.UUID]struct{})
	hub.connectionGroups[connection] = newConnectionSubscriptionGroups()

	if principalSubject == "" {
		return nil
	}

	userConnections := append(hub.connectionsByUser[principalSubject], connection)
	hub.connectionsByUser[principalSubject] = userConnections

	if hub.maxConnectionsPerUser > 0 && len(userConnections) > hub.maxConnectionsPerUser {
		oldest := userConnections[0]
		hub.connectionsByUser[principalSubject] = userConnections[1:]
		return oldest
	}

	return nil
}

func (hub *Hub) UnregisterConnection(connection *Connection) {
	if connection == nil {
		return
	}

	hub.mutex.Lock()
	groups := hub.connectionGroups[connection]
	if groups != nil {
		for topic := range groups.groupedTopics {
			hub.removeGroupTopicSubscriberLocked(topic, connection)
		}

		for tagID := range groups.tagRefs {
			hub.removeTagSubscriptionLocked(connection, tagID)
		}
	}

	delete(hub.connectionSubscriptions, connection)
	delete(hub.connectionGroups, connection)
	delete(hub.alarmsSubscribers, connection)

	principalSubject := strings.TrimSpace(connection.Subject())
	if principalSubject != "" {
		userConnections := hub.connectionsByUser[principalSubject]
		if len(userConnections) > 0 {
			filtered := make([]*Connection, 0, len(userConnections))
			for _, existing := range userConnections {
				if existing != connection {
					filtered = append(filtered, existing)
				}
			}
			if len(filtered) == 0 {
				delete(hub.connectionsByUser, principalSubject)
			} else {
				hub.connectionsByUser[principalSubject] = filtered
			}
		}
	}
	hub.mutex.Unlock()

	hub.debouncer.StopConnection(connection)
}

func (hub *Hub) Subscribe(connection *Connection, topics []string) SubscribeResult {
	result := SubscribeResult{}
	if connection == nil {
		return result
	}

	addedTagIDs := make(map[uuid.UUID]struct{})
	for _, topic := range normalizeTopics(topics) {
		if IsAlarmsTopic(topic) {
			hub.mutex.Lock()
			hub.alarmsSubscribers[connection] = struct{}{}
			hub.mutex.Unlock()

			result.SubscribedTopics = append(result.SubscribedTopics, TopicAlarms)
			result.SubscribedAlarms = true
			continue
		}

		tagID, tagError := ParseTagTopic(topic)
		if tagError == nil {
			added, subscribed, limited := hub.subscribeDirectTag(connection, topic, tagID)
			if limited {
				result.LimitedTopics = append(result.LimitedTopics, topic)
				continue
			}
			if subscribed {
				result.SubscribedTopics = append(result.SubscribedTopics, topic)
				for _, value := range added {
					addedTagIDs[value] = struct{}{}
				}
			}
			continue
		}

		added, subscribed, limitExceeded, resolveError := hub.subscribeGroupTopic(connection, topic)
		if resolveError != nil {
			if errors.Is(resolveError, domain.ErrNotFound) {
				result.InvalidTopics = append(result.InvalidTopics, topic)
				continue
			}
			if errors.Is(resolveError, ErrTopicResolveLimit) {
				result.LimitedTopics = append(result.LimitedTopics, topic)
				continue
			}

			hub.logger.Warn(
				"failed to resolve subscription topic",
				"method",
				"Hub.Subscribe",
				"topic",
				topic,
				"error",
				resolveError,
			)
			result.InvalidTopics = append(result.InvalidTopics, topic)
			continue
		}
		if limitExceeded {
			result.LimitedTopics = append(result.LimitedTopics, topic)
			continue
		}
		if subscribed {
			result.SubscribedTopics = append(result.SubscribedTopics, topic)
			for _, value := range added {
				addedTagIDs[value] = struct{}{}
			}
		}
	}

	result.SubscribedTagIDs = mapUUIDSetToSlice(addedTagIDs)
	for _, tagID := range result.SubscribedTagIDs {
		hub.debouncer.Ensure(connection, tagID)
	}

	return result
}

func (hub *Hub) subscribeDirectTag(connection *Connection, topic string, tagID uuid.UUID) ([]uuid.UUID, bool, bool) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	groups := hub.ensureConnectionGroupsLocked(connection)
	if _, exists := groups.directTags[tagID]; exists {
		return nil, true, false
	}
	if hub.exceedsSubscriptionLimit(groups, []uuid.UUID{tagID}) {
		return nil, false, true
	}

	groups.directTags[tagID] = struct{}{}
	added := hub.incrementTagRefsLocked(connection, groups, []uuid.UUID{tagID})
	return added, true, false
}

func (hub *Hub) subscribeGroupTopic(
	connection *Connection,
	topic string,
) ([]uuid.UUID, bool, bool, error) {
	hub.mutex.RLock()
	groups := hub.connectionGroups[connection]
	if groups != nil {
		if _, exists := groups.groupedTopics[normalizeTopicKey(topic)]; exists {
			hub.mutex.RUnlock()
			return nil, true, false, nil
		}
	}
	hub.mutex.RUnlock()

	if hub.topicResolver == nil {
		return nil, false, false, errors.New("topic resolver is not configured")
	}

	resolveContext, cancelResolve := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelResolve()

	tagIDs, resolveError := hub.topicResolver.ResolveTopic(resolveContext, topic)
	if resolveError != nil {
		return nil, false, false, resolveError
	}

	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	groups = hub.ensureConnectionGroupsLocked(connection)
	normalizedTopic := normalizeTopicKey(topic)
	if _, exists := groups.groupedTopics[normalizedTopic]; exists {
		return nil, true, false, nil
	}
	if hub.exceedsSubscriptionLimit(groups, tagIDs) {
		return nil, false, true, nil
	}

	topicTagSet := make(map[uuid.UUID]struct{}, len(tagIDs))
	for _, tagID := range tagIDs {
		topicTagSet[tagID] = struct{}{}
	}
	groups.groupedTopics[normalizedTopic] = topicTagSet

	subscribers := hub.groupTopicSubscribers[normalizedTopic]
	if subscribers == nil {
		subscribers = make(map[*Connection]struct{})
		hub.groupTopicSubscribers[normalizedTopic] = subscribers
	}
	subscribers[connection] = struct{}{}

	added := hub.incrementTagRefsLocked(connection, groups, tagIDs)
	return added, true, false, nil
}

func (hub *Hub) Unsubscribe(connection *Connection, topics []string) UnsubscribeResult {
	result := UnsubscribeResult{}
	if connection == nil {
		return result
	}

	removedTagIDs := make(map[uuid.UUID]struct{})
	for _, topic := range normalizeTopics(topics) {
		if IsAlarmsTopic(topic) {
			hub.mutex.Lock()
			if _, subscribed := hub.alarmsSubscribers[connection]; subscribed {
				delete(hub.alarmsSubscribers, connection)
				result.UnsubscribedTopics = append(result.UnsubscribedTopics, TopicAlarms)
				result.UnsubscribedAlarms = true
			}
			hub.mutex.Unlock()
			continue
		}

		tagID, tagError := ParseTagTopic(topic)
		if tagError == nil {
			removed, unsubscribed := hub.unsubscribeDirectTag(connection, tagID)
			if unsubscribed {
				result.UnsubscribedTopics = append(result.UnsubscribedTopics, topic)
				for _, value := range removed {
					removedTagIDs[value] = struct{}{}
				}
			}
			continue
		}

		_, _, groupParseError := ParseGroupTopic(topic)
		if groupParseError != nil {
			result.InvalidTopics = append(result.InvalidTopics, topic)
			continue
		}

		removed, unsubscribed := hub.unsubscribeGroupTopic(connection, topic)
		if unsubscribed {
			result.UnsubscribedTopics = append(result.UnsubscribedTopics, topic)
			for _, value := range removed {
				removedTagIDs[value] = struct{}{}
			}
		}
	}

	for _, tagID := range mapUUIDSetToSlice(removedTagIDs) {
		hub.debouncer.Remove(connection, tagID)
	}

	return result
}

func (hub *Hub) unsubscribeDirectTag(connection *Connection, tagID uuid.UUID) ([]uuid.UUID, bool) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	groups := hub.connectionGroups[connection]
	if groups == nil {
		return nil, false
	}
	if _, exists := groups.directTags[tagID]; !exists {
		return nil, false
	}

	delete(groups.directTags, tagID)
	removed := hub.decrementTagRefsLocked(connection, groups, []uuid.UUID{tagID})
	return removed, true
}

func (hub *Hub) unsubscribeGroupTopic(connection *Connection, topic string) ([]uuid.UUID, bool) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	groups := hub.connectionGroups[connection]
	if groups == nil {
		return nil, false
	}

	normalizedTopic := normalizeTopicKey(topic)
	topicTags, exists := groups.groupedTopics[normalizedTopic]
	if !exists {
		return nil, false
	}

	delete(groups.groupedTopics, normalizedTopic)
	hub.removeGroupTopicSubscriberLocked(normalizedTopic, connection)
	removed := hub.decrementTagRefsLocked(connection, groups, mapUUIDKeys(topicTags))
	return removed, true
}

func (hub *Hub) Publish(event TagValueEvent) {
	if event.TagID == uuid.Nil {
		return
	}

	record := event.Record
	if record.TagID == uuid.Nil {
		return
	}

	hub.mutex.Lock()
	existing, exists := hub.lastByTag[event.TagID]
	if !exists || record.Timestamp.After(existing.Timestamp) {
		hub.lastByTag[event.TagID] = record
	}

	connections := make([]*Connection, 0)
	for connection := range hub.subscriptions[event.TagID] {
		connections = append(connections, connection)
	}
	hub.mutex.Unlock()

	for _, connection := range connections {
		hub.debouncer.Publish(connection, record)
	}
}

func (hub *Hub) LoadLatestRecords(
	ctx context.Context,
	tagIDs []uuid.UUID,
) map[uuid.UUID]domain.IngestRecord {
	result := make(map[uuid.UUID]domain.IngestRecord)
	missing := make([]uuid.UUID, 0)

	hub.mutex.RLock()
	for _, tagID := range tagIDs {
		if tagID == uuid.Nil {
			continue
		}

		record, exists := hub.lastByTag[tagID]
		if exists {
			result[tagID] = record
			continue
		}
		missing = append(missing, tagID)
	}
	hub.mutex.RUnlock()

	if hub.lastValueProvider == nil || len(missing) == 0 {
		return result
	}

	for _, tagID := range missing {
		record, found, readError := hub.lastValueProvider.GetLastValue(ctx, tagID)
		if readError != nil {
			hub.logger.Warn(
				"failed to read snapshot from cache",
				"method",
				"Hub.LoadLatestRecords",
				"tag_id",
				tagID.String(),
				"error",
				readError,
			)
			continue
		}
		if !found {
			continue
		}

		result[tagID] = record
		hub.mutex.Lock()
		existing, exists := hub.lastByTag[tagID]
		if !exists || record.Timestamp.After(existing.Timestamp) {
			hub.lastByTag[tagID] = record
		}
		hub.mutex.Unlock()
	}

	return result
}

func (hub *Hub) SubscriptionCount(connection *Connection) int {
	if connection == nil {
		return 0
	}

	hub.mutex.RLock()
	defer hub.mutex.RUnlock()

	groups := hub.connectionGroups[connection]
	if groups == nil {
		return 0
	}
	return len(groups.tagRefs)
}

func (hub *Hub) HandleConfigChanged(event domain.ConfigChangedEvent) {
	if hub.topicResolver == nil {
		return
	}

	affectedTopics := hub.topicResolver.InvalidateByConfigEvent(event)
	if len(affectedTopics) == 0 {
		return
	}

	for _, topic := range affectedTopics {
		hub.reconcileGroupTopic(topic)
	}
}

func (hub *Hub) BroadcastConfigChanged(event domain.ConfigChangedEvent) {
	hub.mutex.RLock()
	connections := make([]*Connection, 0, len(hub.connectionSubscriptions))
	for connection := range hub.connectionSubscriptions {
		connections = append(connections, connection)
	}
	hub.mutex.RUnlock()

	for _, connection := range connections {
		connection.enqueueConfigChanged(event)
	}
}

func (hub *Hub) BroadcastAlarm(event domain.AlarmEvent) {
	hub.mutex.RLock()
	connections := make([]*Connection, 0, len(hub.alarmsSubscribers))
	for connection := range hub.alarmsSubscribers {
		connections = append(connections, connection)
	}
	hub.mutex.RUnlock()

	for _, connection := range connections {
		connection.enqueueAlarm(event)
	}
}

func (hub *Hub) BroadcastAlarmBatch(events []domain.AlarmEvent) {
	if len(events) == 0 {
		return
	}

	hub.mutex.RLock()
	connections := make([]*Connection, 0, len(hub.alarmsSubscribers))
	for connection := range hub.alarmsSubscribers {
		connections = append(connections, connection)
	}
	hub.mutex.RUnlock()

	for _, connection := range connections {
		connection.enqueueAlarmsBatch(events)
	}
}

func (hub *Hub) LoadActiveAlarmSnapshot(
	ctx context.Context,
) ([]domain.AlarmStateRecord, error) {
	if hub.alarmSnapshotProvider == nil {
		return nil, nil
	}

	return hub.alarmSnapshotProvider.GetActiveAlarmSnapshot(ctx)
}

func (hub *Hub) UnsubscribeTagFromAll(tagID uuid.UUID) {
	if tagID == uuid.Nil {
		return
	}

	hub.mutex.Lock()
	tagSubscriptions := hub.subscriptions[tagID]
	if len(tagSubscriptions) == 0 {
		hub.mutex.Unlock()
		return
	}

	connections := make([]*Connection, 0, len(tagSubscriptions))
	for connection := range tagSubscriptions {
		connections = append(connections, connection)

		groups := hub.connectionGroups[connection]
		if groups == nil {
			continue
		}
		if _, direct := groups.directTags[tagID]; direct {
			delete(groups.directTags, tagID)
		}
		for topic, topicTags := range groups.groupedTopics {
			if _, exists := topicTags[tagID]; !exists {
				continue
			}
			delete(topicTags, tagID)
			if len(topicTags) == 0 {
				delete(groups.groupedTopics, topic)
				hub.removeGroupTopicSubscriberLocked(topic, connection)
			}
		}

		hub.decrementTagRefForSingleLocked(connection, groups, tagID)
	}
	hub.mutex.Unlock()

	for _, connection := range connections {
		hub.debouncer.Remove(connection, tagID)
	}
}

func (hub *Hub) reconcileGroupTopic(topic string) {
	if hub.topicResolver == nil {
		return
	}

	normalizedTopic := normalizeTopicKey(topic)
	if normalizedTopic == "" {
		return
	}

	hub.mutex.RLock()
	subscribersSet := hub.groupTopicSubscribers[normalizedTopic]
	if len(subscribersSet) == 0 {
		hub.mutex.RUnlock()
		return
	}
	subscribers := make([]*Connection, 0, len(subscribersSet))
	for connection := range subscribersSet {
		subscribers = append(subscribers, connection)
	}
	hub.mutex.RUnlock()

	resolveContext, cancelResolve := context.WithTimeout(context.Background(), 5*time.Second)
	resolvedTagIDs, resolveError := hub.topicResolver.ResolveTopic(resolveContext, normalizedTopic)
	cancelResolve()

	resolvedSet := make(map[uuid.UUID]struct{}, len(resolvedTagIDs))
	if resolveError == nil {
		for _, tagID := range resolvedTagIDs {
			resolvedSet[tagID] = struct{}{}
		}
	}

	for _, connection := range subscribers {
		addedTagIDs := make([]uuid.UUID, 0)
		removedTagIDs := make([]uuid.UUID, 0)

		hub.mutex.Lock()
		groups := hub.connectionGroups[connection]
		if groups == nil {
			hub.mutex.Unlock()
			continue
		}

		oldSet, exists := groups.groupedTopics[normalizedTopic]
		if !exists {
			hub.mutex.Unlock()
			continue
		}

		if resolveError != nil {
			removedTagIDs = append(removedTagIDs, hub.decrementTagRefsLocked(connection, groups, mapUUIDKeys(oldSet))...)
			delete(groups.groupedTopics, normalizedTopic)
			hub.removeGroupTopicSubscriberLocked(normalizedTopic, connection)
		} else {
			removed := make([]uuid.UUID, 0)
			for tagID := range oldSet {
				if _, stillExists := resolvedSet[tagID]; stillExists {
					continue
				}
				removed = append(removed, tagID)
			}
			added := make([]uuid.UUID, 0)
			for tagID := range resolvedSet {
				if _, alreadyExists := oldSet[tagID]; alreadyExists {
					continue
				}
				added = append(added, tagID)
			}

			if hub.exceedsSubscriptionLimit(groups, added) {
				available := hub.maxSubscriptionsPerConn - len(groups.tagRefs)
				if available < 0 {
					available = 0
				}
				trimmed := make([]uuid.UUID, 0, available)
				for _, tagID := range added {
					if available == 0 {
						break
					}
					trimmed = append(trimmed, tagID)
					available--
				}
				added = trimmed
			}

			removedTagIDs = append(removedTagIDs, hub.decrementTagRefsLocked(connection, groups, removed)...)
			addedTagIDs = append(addedTagIDs, hub.incrementTagRefsLocked(connection, groups, added)...)

			nextSet := make(map[uuid.UUID]struct{}, len(resolvedSet))
			for tagID := range resolvedSet {
				nextSet[tagID] = struct{}{}
			}
			groups.groupedTopics[normalizedTopic] = nextSet
		}
		hub.mutex.Unlock()

		for _, tagID := range addedTagIDs {
			hub.debouncer.Ensure(connection, tagID)
		}
		for _, tagID := range removedTagIDs {
			hub.debouncer.Remove(connection, tagID)
		}

		if len(addedTagIDs) > 0 {
			connection.enqueueSubscribeSnapshots(addedTagIDs)
		}
		if len(addedTagIDs) > 0 || len(removedTagIDs) > 0 {
			connection.enqueueTopicsChanged(normalizedTopic, len(addedTagIDs), len(removedTagIDs))
		}
	}
}

func (hub *Hub) ensureConnectionGroupsLocked(connection *Connection) *connectionSubscriptionGroups {
	groups := hub.connectionGroups[connection]
	if groups == nil {
		groups = newConnectionSubscriptionGroups()
		hub.connectionGroups[connection] = groups
	}
	if hub.connectionSubscriptions[connection] == nil {
		hub.connectionSubscriptions[connection] = make(map[uuid.UUID]struct{})
	}
	return groups
}

func (hub *Hub) exceedsSubscriptionLimit(groups *connectionSubscriptionGroups, candidateTagIDs []uuid.UUID) bool {
	if hub.maxSubscriptionsPerConn <= 0 {
		return false
	}
	if groups == nil {
		return false
	}

	uniqueToAdd := 0
	for _, tagID := range candidateTagIDs {
		if tagID == uuid.Nil {
			continue
		}
		if groups.tagRefs[tagID] > 0 {
			continue
		}
		uniqueToAdd++
	}

	return len(groups.tagRefs)+uniqueToAdd > hub.maxSubscriptionsPerConn
}

func (hub *Hub) incrementTagRefsLocked(
	connection *Connection,
	groups *connectionSubscriptionGroups,
	tagIDs []uuid.UUID,
) []uuid.UUID {
	added := make([]uuid.UUID, 0)
	for _, tagID := range uniqueUUIDs(tagIDs) {
		current := groups.tagRefs[tagID]
		groups.tagRefs[tagID] = current + 1
		if current > 0 {
			continue
		}

		hub.connectionSubscriptions[connection][tagID] = struct{}{}
		subscribers := hub.subscriptions[tagID]
		if subscribers == nil {
			subscribers = make(map[*Connection]struct{})
			hub.subscriptions[tagID] = subscribers
		}
		subscribers[connection] = struct{}{}
		added = append(added, tagID)
	}

	return added
}

func (hub *Hub) decrementTagRefsLocked(
	connection *Connection,
	groups *connectionSubscriptionGroups,
	tagIDs []uuid.UUID,
) []uuid.UUID {
	removed := make([]uuid.UUID, 0)
	for _, tagID := range uniqueUUIDs(tagIDs) {
		if hub.decrementTagRefForSingleLocked(connection, groups, tagID) {
			removed = append(removed, tagID)
		}
	}
	return removed
}

func (hub *Hub) decrementTagRefForSingleLocked(
	connection *Connection,
	groups *connectionSubscriptionGroups,
	tagID uuid.UUID,
) bool {
	current := groups.tagRefs[tagID]
	if current <= 0 {
		return false
	}
	if current > 1 {
		groups.tagRefs[tagID] = current - 1
		return false
	}

	delete(groups.tagRefs, tagID)
	delete(hub.connectionSubscriptions[connection], tagID)
	hub.removeTagSubscriptionLocked(connection, tagID)
	return true
}

func (hub *Hub) removeTagSubscriptionLocked(connection *Connection, tagID uuid.UUID) {
	tagSubscriptions := hub.subscriptions[tagID]
	if tagSubscriptions == nil {
		return
	}
	delete(tagSubscriptions, connection)
	if len(tagSubscriptions) == 0 {
		delete(hub.subscriptions, tagID)
	}
}

func (hub *Hub) removeGroupTopicSubscriberLocked(topic string, connection *Connection) {
	subscribers := hub.groupTopicSubscribers[topic]
	if subscribers == nil {
		return
	}
	delete(subscribers, connection)
	if len(subscribers) == 0 {
		delete(hub.groupTopicSubscribers, topic)
	}
}

func mapUUIDSetToSlice(values map[uuid.UUID]struct{}) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	return result
}

func mapUUIDKeys(values map[uuid.UUID]struct{}) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	return result
}
