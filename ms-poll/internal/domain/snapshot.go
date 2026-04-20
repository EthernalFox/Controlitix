package domain

import (
	"encoding/json"
	"time"
)

type DeviceSnapshot struct {
	ID        string
	ObjectID  *string
	TypeID    int
	TypeName  string
	Name      string
	Settings  json.RawMessage
	UpdatedAt time.Time
}

type TagSnapshot struct {
	ID         string
	DeviceID   string
	Name       string
	DataType   string
	UnitSymbol *string
	Address    json.RawMessage
	Scaling    *TagScaling
	Setpoints  *TagSetpoints
	UpdatedAt  time.Time
}

type TagScaling struct {
	RawMin *float64
	RawMax *float64
	EngMin *float64
	EngMax *float64
	Factor *float64
	Offset *float64
}

type TagSetpoints struct {
	LoLo *float64
	Lo   *float64
	Hi   *float64
	HiHi *float64
}

type Snapshot struct {
	Devices      map[string]*DeviceSnapshot
	Tags         map[string]*TagSnapshot
	TagsByDevice map[string][]*TagSnapshot
	LoadedAt     time.Time
}
