package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AlarmState int16

var (
	ErrAlarmNotActive    = errors.New("alarm not active")
	ErrAlarmAlreadyAcked = errors.New("alarm already acked")
)

const (
	AlarmStateOK        AlarmState = 0
	AlarmStateLo        AlarmState = 1
	AlarmStateHi        AlarmState = 2
	AlarmStateLoLo      AlarmState = 3
	AlarmStateHiHi      AlarmState = 4
	AlarmStateUncertain AlarmState = 5
	AlarmStateBad       AlarmState = 6
	AlarmStateCommLoss  AlarmState = 7
	AlarmStateOffline   AlarmState = 8
)

func (state AlarmState) String() string {
	switch state {
	case AlarmStateLo:
		return "lo"
	case AlarmStateHi:
		return "hi"
	case AlarmStateLoLo:
		return "lolo"
	case AlarmStateHiHi:
		return "hihi"
	case AlarmStateUncertain:
		return "uncertain"
	case AlarmStateBad:
		return "bad"
	case AlarmStateCommLoss:
		return "comm_loss"
	case AlarmStateOffline:
		return "offline"
	default:
		return "ok"
	}
}

type AlarmEventType int16

const (
	AlarmEventRaised       AlarmEventType = 0
	AlarmEventCleared      AlarmEventType = 1
	AlarmEventAcked        AlarmEventType = 2
	AlarmEventSuppressed   AlarmEventType = 3
	AlarmEventUnsuppressed AlarmEventType = 4
)

func (eventType AlarmEventType) String() string {
	switch eventType {
	case AlarmEventCleared:
		return "cleared"
	case AlarmEventAcked:
		return "acked"
	case AlarmEventSuppressed:
		return "suppressed"
	case AlarmEventUnsuppressed:
		return "unsuppressed"
	default:
		return "raised"
	}
}

type Setpoints struct {
	LoLo *float64
	Lo   *float64
	Hi   *float64
	HiHi *float64
}

type AlarmStateRecord struct {
	TagID        uuid.UUID
	State        AlarmState
	LastValue    *float64
	LastQuality  Quality
	EnteredAt    time.Time
	LastSeenAt   time.Time
	Suppressed   bool
	UpdatedAt    time.Time
	ObjectID     uuid.UUID
	ObjectName   string
	DeviceID     uuid.UUID
	DeviceName   string
	TagName      string
	Ack          *AlarmAck
}

type AlarmAck struct {
	State   AlarmState
	ActorID string
	Note    *string
	AckedAt time.Time
}

type AlarmEvent struct {
	ID        uuid.UUID
	TagID     uuid.UUID
	EventType AlarmEventType
	StateFrom AlarmState
	StateTo   AlarmState
	Value     *float64
	Quality   Quality
	TS        time.Time
	ActorID   *string
	Note      *string
	CreatedAt time.Time
}

type AlarmListStatus string

const (
	AlarmListStatusActive  AlarmListStatus = "active"
	AlarmListStatusAcked   AlarmListStatus = "acked"
	AlarmListStatusCleared AlarmListStatus = "cleared"
)

type AlarmSeverityFilter string

const (
	AlarmSeverityInfo  AlarmSeverityFilter = "info"
	AlarmSeverityWarn  AlarmSeverityFilter = "warn"
	AlarmSeverityAlarm AlarmSeverityFilter = "alarm"
)

type AlarmListQuery struct {
	Status   AlarmListStatus
	Severity AlarmSeverityFilter
	ObjectID uuid.UUID
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

type AlarmListResult struct {
	Items  []AlarmStateRecord
	Total  int
	Limit  int
	Offset int
}

type AlarmDetail struct {
	CurrentState *AlarmStateRecord
	Events       []AlarmEvent
}

type AlarmTransitionInput struct {
	TagID      uuid.UUID
	StateFrom  AlarmState
	StateTo    AlarmState
	Value      *float64
	Quality    Quality
	TS         time.Time
	EnteredAt  time.Time
	LastSeenAt time.Time
}

type AlarmAcknowledgeResult struct {
	TagID uuid.UUID
	State AlarmState
	Ack   AlarmAck
	Event AlarmEvent
}

func ParseAlarmListStatus(value string) AlarmListStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(AlarmListStatusAcked):
		return AlarmListStatusAcked
	case string(AlarmListStatusCleared):
		return AlarmListStatusCleared
	default:
		return AlarmListStatusActive
	}
}

func ParseAlarmSeverity(value string) AlarmSeverityFilter {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "alarm":
		return AlarmSeverityAlarm
	case "info", "warn":
		return AlarmSeverityWarn
	default:
		return ""
	}
}
