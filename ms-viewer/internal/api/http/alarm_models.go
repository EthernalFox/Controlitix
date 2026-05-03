package http

import (
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type alarmsListResponse struct {
	Items  []alarmRecordResponse `json:"items"`
	Total  int                   `json:"total"`
	Offset int                   `json:"offset"`
	Limit  int                   `json:"limit"`
}

type alarmRecordResponse struct {
	TagID      string               `json:"tag_id"`
	TagName    string               `json:"tag_name"`
	DeviceID   string               `json:"device_id"`
	DeviceName string               `json:"device_name"`
	ObjectID   string               `json:"object_id"`
	ObjectName string               `json:"object_name"`
	State      string               `json:"state"`
	Value      *float64             `json:"value"`
	Quality    string               `json:"quality"`
	EnteredAt  string               `json:"entered_at"`
	LastSeenAt string               `json:"last_seen_at"`
	Acked      bool                 `json:"acked"`
	Ack        *alarmAckDTOResponse `json:"ack"`
}

type alarmAckDTOResponse struct {
	ActorID string  `json:"actor_id"`
	AckedAt string  `json:"acked_at"`
	Note    *string `json:"note"`
}

type alarmEventResponse struct {
	ID        string   `json:"id"`
	EventType string   `json:"event_type"`
	TagID     string   `json:"tag_id"`
	StateFrom string   `json:"state_from"`
	StateTo   string   `json:"state_to"`
	Value     *float64 `json:"value"`
	Quality   string   `json:"quality"`
	TS        string   `json:"ts"`
	ActorID   *string  `json:"actor_id"`
	Note      *string  `json:"note"`
}

type alarmDetailResponse struct {
	Current *alarmRecordResponse `json:"current"`
	Events  []alarmEventResponse `json:"events"`
}

type acknowledgeRequest struct {
	Note *string `json:"note"`
}

type bulkAcknowledgeRequestItem struct {
	TagID string `json:"tag_id"`
}

type bulkAcknowledgeRequest struct {
	Items []bulkAcknowledgeRequestItem `json:"items"`
	Note  *string                      `json:"note,omitempty"`
}

type bulkAcknowledgeItemResponse struct {
	TagID  string  `json:"tag_id"`
	Status string  `json:"status"`
	State  *string `json:"state,omitempty"`
}

type bulkAcknowledgeResponse struct {
	Items    []bulkAcknowledgeItemResponse `json:"items"`
	AckedAt  string                        `json:"acked_at"`
	ActorID  string                        `json:"actor_id"`
	SuccessN int                           `json:"success_n"`
	FailedN  int                           `json:"failed_n"`
}

func mapAlarmsListResponse(result domain.AlarmListResult) alarmsListResponse {
	items := make([]alarmRecordResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, mapAlarmRecordResponse(item))
	}

	return alarmsListResponse{
		Items:  items,
		Total:  result.Total,
		Offset: result.Offset,
		Limit:  result.Limit,
	}
}

func mapAlarmRecordResponse(record domain.AlarmStateRecord) alarmRecordResponse {
	var ack *alarmAckDTOResponse
	if record.Ack != nil {
		ack = &alarmAckDTOResponse{
			ActorID: record.Ack.ActorID,
			AckedAt: record.Ack.AckedAt.UTC().Format(time.RFC3339Nano),
			Note:    record.Ack.Note,
		}
	}

	return alarmRecordResponse{
		TagID:      record.TagID.String(),
		TagName:    record.TagName,
		DeviceID:   record.DeviceID.String(),
		DeviceName: record.DeviceName,
		ObjectID:   record.ObjectID.String(),
		ObjectName: record.ObjectName,
		State:      record.State.String(),
		Value:      record.LastValue,
		Quality:    string(record.LastQuality),
		EnteredAt:  record.EnteredAt.UTC().Format(time.RFC3339Nano),
		LastSeenAt: record.LastSeenAt.UTC().Format(time.RFC3339Nano),
		Acked:      record.Ack != nil,
		Ack:        ack,
	}
}

func mapAlarmDetailResponse(detail domain.AlarmDetail) alarmDetailResponse {
	events := make([]alarmEventResponse, 0, len(detail.Events))
	for _, event := range detail.Events {
		events = append(events, alarmEventResponse{
			ID:        event.ID.String(),
			EventType: event.EventType.String(),
			TagID:     event.TagID.String(),
			StateFrom: event.StateFrom.String(),
			StateTo:   event.StateTo.String(),
			Value:     event.Value,
			Quality:   string(event.Quality),
			TS:        event.TS.UTC().Format(time.RFC3339Nano),
			ActorID:   event.ActorID,
			Note:      event.Note,
		})
	}

	var current *alarmRecordResponse
	if detail.CurrentState != nil {
		response := mapAlarmRecordResponse(*detail.CurrentState)
		current = &response
	}

	return alarmDetailResponse{
		Current: current,
		Events:  events,
	}
}

func mapBulkAcknowledgeResponse(result domain.AlarmBulkAcknowledgeResult) bulkAcknowledgeResponse {
	items := make([]bulkAcknowledgeItemResponse, 0, len(result.Items))
	for _, item := range result.Items {
		var state *string
		if item.State != nil {
			value := item.State.String()
			state = &value
		}

		items = append(items, bulkAcknowledgeItemResponse{
			TagID:  item.TagID.String(),
			Status: string(item.Status),
			State:  state,
		})
	}

	return bulkAcknowledgeResponse{
		Items:    items,
		AckedAt:  result.AckedAt.UTC().Format(time.RFC3339Nano),
		ActorID:  result.ActorID,
		SuccessN: result.SuccessN,
		FailedN:  result.FailedN,
	}
}
