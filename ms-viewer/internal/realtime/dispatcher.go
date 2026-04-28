package realtime

import "github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"

type Dispatcher struct {
	hub *Hub
}

func NewDispatcher(hub *Hub) *Dispatcher {
	return &Dispatcher{hub: hub}
}

func (dispatcher *Dispatcher) Publish(record domain.IngestRecord) {
	if dispatcher == nil || dispatcher.hub == nil {
		return
	}

	dispatcher.hub.Publish(TagValueEvent{
		TagID:   record.TagID,
		Value:   record.Value,
		Quality: record.Quality,
		Record:  record,
	})
}
