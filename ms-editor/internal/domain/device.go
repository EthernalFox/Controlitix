package domain

import (
	"encoding/json"
	"time"
)

type DeviceType struct {
	ID   int
	Name string
}

type Device struct {
	ID          string
	ObjectID    *string
	TypeID      int
	TypeName    string
	Name        string
	Description *string
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DeviceParams struct {
	DeviceID  string
	Settings  json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DeviceWithParams struct {
	Device Device
	Params *DeviceParams
}

type DeviceUpdate struct {
	ObjectID    **string
	TypeID      *int
	Name        *string
	Description *string
}
