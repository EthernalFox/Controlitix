package usecase

import (
	"log/slog"
	"reflect"
)

const snapshotSubscriptionBufferSize = 64

type SnapshotEvent struct {
	Kind     string
	DeviceID string
	TagID    string
}

func (store *SnapshotStore) Subscribe() <-chan SnapshotEvent {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	subscriber := make(chan SnapshotEvent, snapshotSubscriptionBufferSize)
	subscriberKey := channelPointer(subscriber)
	store.subscribers[subscriberKey] = subscriber

	return subscriber
}

func (store *SnapshotStore) Unsubscribe(subscriber <-chan SnapshotEvent) {
	if subscriber == nil {
		return
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	subscriberKey := channelPointer(subscriber)
	internalChannel, exists := store.subscribers[subscriberKey]
	if !exists {
		return
	}

	delete(store.subscribers, subscriberKey)
	close(internalChannel)
}

func (store *SnapshotStore) publishEventLocked(event SnapshotEvent) {
	for subscriberKey, subscriber := range store.subscribers {
		select {
		case subscriber <- event:
		default:
			if store.logger != nil {
				store.logger.Warn(
					"snapshot event dropped",
					"kind",
					event.Kind,
					"device_id",
					event.DeviceID,
					"tag_id",
					event.TagID,
					"subscriber",
					subscriberKey,
				)
			}
		}
	}
}

func channelPointer(channel any) uintptr {
	return reflect.ValueOf(channel).Pointer()
}

func defaultSnapshotStoreLogger(logger *slog.Logger) *slog.Logger {
	if logger != nil {
		return logger
	}

	return slog.Default()
}
