package http

import (
	"encoding/json"
	"time"
)

type monitoringObjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type monitoringObjectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type diagramRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type diagramResponse struct {
	ID          string     `json:"id"`
	ObjectID    string     `json:"object_id"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type createFiguresRequest struct {
	Figures []figureRequest `json:"figures"`
}

type figureRequest struct {
	FigureType string          `json:"type"`
	TagID      *string         `json:"tag_id"`
	Parameters json.RawMessage `json:"params"`
}

type updateFigureRequest struct {
	FigureType *string          `json:"type"`
	TagID      *string          `json:"tag_id"`
	Parameters *json.RawMessage `json:"params"`
}

type figureResponse struct {
	ID         string          `json:"id"`
	DiagramID  string          `json:"diagram_id"`
	TagID      *string         `json:"tag_id"`
	FigureType string          `json:"type"`
	Parameters json.RawMessage `json:"params"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type createDeviceRequest struct {
	ObjectID    *string         `json:"object_id"`
	TypeID      int             `json:"type_id"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Settings    json.RawMessage `json:"settings"`
}

type updateDeviceRequest struct {
	TypeID      *int    `json:"type_id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type assignDeviceRequest struct {
	ObjectID *string `json:"object_id"`
}

type deviceParamsRequest struct {
	Settings json.RawMessage `json:"settings"`
}

type deviceResponse struct {
	ID          string          `json:"id"`
	ObjectID    *string         `json:"object_id"`
	TypeID      int             `json:"type_id"`
	TypeName    string          `json:"type_name"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Settings    json.RawMessage `json:"settings"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type deviceTypeResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type errorResponse struct {
	Message string `json:"message"`
}
