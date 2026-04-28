package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type FigureType string

const (
	FigureTypeRect    FigureType = "rect"
	FigureTypeCircle  FigureType = "circle"
	FigureTypeEllipse FigureType = "ellipse"
	FigureTypeWedge   FigureType = "wedge"
	FigureTypeLine    FigureType = "line"
	FigureTypeImage   FigureType = "image"
	FigureTypeText    FigureType = "text"
	FigureTypeRing    FigureType = "ring"
	FigureTypeArc     FigureType = "arc"
	FigureTypeTag     FigureType = "tag"
	FigureTypePath    FigureType = "path"
)

type DiagramCanvas struct {
	Width      int
	Height     int
	Background string
}

type DiagramListQuery struct {
	ObjectID uuid.UUID
	Limit    int
	Offset   int
}

type DiagramListItem struct {
	ID            uuid.UUID
	ObjectID      uuid.UUID
	Name          string
	Description   string
	PublishedAt   time.Time
	FigureCount   int
	BoundTagCount int
}

type DiagramListResult struct {
	Items  []DiagramListItem
	Total  int
	Limit  int
	Offset int
}

type ObjectListQuery struct {
	Limit  int
	Offset int
}

type ObjectListItem struct {
	ID                     uuid.UUID
	Name                   string
	Description            string
	PublishedDiagramCount  int
	FirstPublishedDiagramID uuid.UUID
}

type ObjectListResult struct {
	Items  []ObjectListItem
	Total  int
	Limit  int
	Offset int
}

type TagBrief struct {
	ID         uuid.UUID
	Name       string
	DeviceID   uuid.UUID
	DeviceName string
	Unit       Unit
	DataType   DataType
}

type DiagramFigure struct {
	ID     uuid.UUID
	Type   FigureType
	TagID  *uuid.UUID
	Params json.RawMessage
	Tag    *TagBrief
}

type Diagram struct {
	ID          uuid.UUID
	ObjectID    uuid.UUID
	ObjectName  string
	Name        string
	Description string
	PublishedAt time.Time
	Canvas      DiagramCanvas
	Figures     []DiagramFigure
}

type DiagramSnapshot struct {
	DiagramID      uuid.UUID
	TS             time.Time
	Values         []IngestRecord
	MissingTagIDs  []uuid.UUID
}
