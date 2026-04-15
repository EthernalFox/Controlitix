package domain

import (
	"context"
	"encoding/json"
)

type MonitoringObjectUpdate struct {
	Name        *string
	Description *string
}

type DiagramUpdate struct {
	Name        *string
	Description *string
}

type FigureUpdate struct {
	TagID      *string
	FigureType *string
	Parameters *json.RawMessage
}

type MonitoringObjectRepository interface {
	CreateMonitoringObject(
		ctx context.Context,
		monitoringObject MonitoringObject,
	) (MonitoringObject, error)
	GetMonitoringObject(
		ctx context.Context,
		monitoringObjectID string,
	) (MonitoringObject, error)
	UpdateMonitoringObject(
		ctx context.Context,
		monitoringObjectID string,
		update MonitoringObjectUpdate,
	) (MonitoringObject, error)
	DeleteMonitoringObject(
		ctx context.Context,
		monitoringObjectID string,
	) error
	ListMonitoringObjects(
		ctx context.Context,
	) ([]MonitoringObject, error)
}

type DiagramRepository interface {
	CreateDiagram(
		ctx context.Context,
		diagram Diagram,
	) (Diagram, error)
	GetDiagram(
		ctx context.Context,
		diagramID string,
	) (Diagram, error)
	UpdateDiagram(
		ctx context.Context,
		diagramID string,
		update DiagramUpdate,
	) (Diagram, error)
	DeleteDiagram(
		ctx context.Context,
		diagramID string,
	) error
	PublishDiagram(
		ctx context.Context,
		diagramID string,
	) (Diagram, error)
	ListDiagramsByMonitoringObject(
		ctx context.Context,
		monitoringObjectID string,
	) ([]Diagram, error)
}

type FigureRepository interface {
	CreateFigures(
		ctx context.Context,
		diagramID string,
		figures []Figure,
	) ([]Figure, error)
	UpdateFigure(
		ctx context.Context,
		figureID string,
		update FigureUpdate,
	) (Figure, error)
	DeleteFigure(
		ctx context.Context,
		figureID string,
	) error
	ListFiguresByDiagram(
		ctx context.Context,
		diagramID string,
	) ([]Figure, error)
}
