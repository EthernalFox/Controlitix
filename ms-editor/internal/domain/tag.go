package domain

import (
	"encoding/json"
	"time"
)

type Tag struct {
	ID          string
	DeviceID    string
	Name        string
	Description *string
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TagParams struct {
	ID         string
	TagID      string
	DataTypeID int
	UnitID     *int
	Address    json.RawMessage
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type TagSetpoints struct {
	ParamID   string
	LoLo      *float64
	Lo        *float64
	Hi        *float64
	HiHi      *float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TagScaling struct {
	ParamID   string
	RawMin    *float64
	RawMax    *float64
	EngMin    *float64
	EngMax    *float64
	Factor    *float64
	Offset    *float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TagFull struct {
	Tag       Tag
	Params    *TagParams
	Setpoints *TagSetpoints
	Scaling   *TagScaling
}

type TagUpdate struct {
	Name        *string
	Description *string
}

type TagParamsUpdate struct {
	DataTypeID *int
	UnitID     *int
	Address    *json.RawMessage
}
