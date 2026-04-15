package domain

import (
	"encoding/json"
	"time"
)

type Figure struct {
	ID         string
	DiagramID  string
	TagID      *string
	FigureType string
	Parameters json.RawMessage
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
