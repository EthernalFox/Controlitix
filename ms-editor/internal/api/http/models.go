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

type bulkUpsertFiguresRequest struct {
	Figures []bulkFigureItemRequest `json:"figures"`
}

type figureRequest struct {
	FigureType string          `json:"type"`
	TagID      *string         `json:"tag_id"`
	Parameters json.RawMessage `json:"params"`
}

type bulkFigureItemRequest struct {
	ID         *string         `json:"id,omitempty"`
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

type bulkUpsertFiguresResponse struct {
	Figures []figureResponse `json:"figures"`
	Summary bulkSummary      `json:"summary"`
}

type bulkSummary struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Deleted int `json:"deleted"`
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

type createTagRequest struct {
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Params      *tagParamsInput `json:"params"`
	Setpoints   *setpointsInput `json:"setpoints"`
	Scaling     *scalingInput   `json:"scaling"`
}

type tagParamsInput struct {
	DataTypeID int             `json:"data_type_id"`
	UnitID     *int            `json:"unit_id"`
	Address    json.RawMessage `json:"address"`
}

type setpointsInput struct {
	LoLo *float64 `json:"lolo"`
	Lo   *float64 `json:"lo"`
	Hi   *float64 `json:"hi"`
	HiHi *float64 `json:"hihi"`
}

type scalingInput struct {
	RawMin *float64 `json:"raw_min"`
	RawMax *float64 `json:"raw_max"`
	EngMin *float64 `json:"eng_min"`
	EngMax *float64 `json:"eng_max"`
	Factor *float64 `json:"factor"`
	Offset *float64 `json:"offset"`
}

type updateTagRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type tagResponse struct {
	ID          string             `json:"id"`
	DeviceID    string             `json:"device_id"`
	Name        string             `json:"name"`
	Description *string            `json:"description"`
	Params      *tagParamsResponse `json:"params"`
	Setpoints   *setpointsInput    `json:"setpoints"`
	Scaling     *scalingInput      `json:"scaling"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type tagParamsResponse struct {
	ID         string          `json:"id"`
	DataTypeID int             `json:"data_type_id"`
	UnitID     *int            `json:"unit_id"`
	Address    json.RawMessage `json:"address"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type dataTypeResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type unitResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Category string `json:"category"`
}

type listResponse struct {
	Items  any `json:"items"`
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type problemResponse struct {
	Type     string       `json:"type"`
	Title    string       `json:"title"`
	Status   int          `json:"status"`
	Detail   string       `json:"detail,omitempty"`
	Instance string       `json:"instance,omitempty"`
	Errors   []fieldError `json:"errors,omitempty"`
}

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
