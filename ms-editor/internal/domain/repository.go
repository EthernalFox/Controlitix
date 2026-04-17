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
	FigureType *FigureType
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
		query ObjectListQuery,
	) (ListResult[MonitoringObject], error)
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
	ListDiagrams(
		ctx context.Context,
		query DiagramListQuery,
	) (ListResult[Diagram], error)
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
	ListFigures(
		ctx context.Context,
		query FigureListQuery,
	) (ListResult[Figure], error)
}

type DeviceRepository interface {
	CreateDevice(
		ctx context.Context,
		device Device,
	) (Device, error)
	GetDevice(
		ctx context.Context,
		deviceID string,
	) (DeviceWithParams, error)
	UpdateDevice(
		ctx context.Context,
		deviceID string,
		update DeviceUpdate,
	) (Device, error)
	DeleteDevice(
		ctx context.Context,
		deviceID string,
	) error
	ListDevices(
		ctx context.Context,
		query DeviceListQuery,
	) (ListResult[Device], error)
	AssignDeviceToObject(
		ctx context.Context,
		deviceID string,
		objectID *string,
	) (Device, error)
}

type DeviceParamsRepository interface {
	UpsertDeviceParams(
		ctx context.Context,
		deviceID string,
		settings json.RawMessage,
	) (DeviceParams, error)
	GetDeviceParams(
		ctx context.Context,
		deviceID string,
	) (DeviceParams, error)
}

type DeviceTypeRepository interface {
	ListDeviceTypes(
		ctx context.Context,
	) ([]DeviceType, error)
	GetDeviceType(
		ctx context.Context,
		typeID int,
	) (DeviceType, error)
}

type TagRepository interface {
	CreateTag(
		ctx context.Context,
		tag Tag,
	) (Tag, error)
	GetTag(
		ctx context.Context,
		tagID string,
	) (TagFull, error)
	UpdateTag(
		ctx context.Context,
		tagID string,
		update TagUpdate,
	) (Tag, error)
	DeleteTag(
		ctx context.Context,
		tagID string,
	) error
	ListTags(
		ctx context.Context,
		query TagListQuery,
	) (ListResult[Tag], error)
}

type TagParamsRepository interface {
	UpsertTagParams(
		ctx context.Context,
		tagID string,
		params TagParams,
	) (TagParams, error)
	GetTagParams(
		ctx context.Context,
		tagID string,
	) (TagParams, error)
}

type TagSetpointsRepository interface {
	UpsertTagSetpoints(
		ctx context.Context,
		paramID string,
		setpoints TagSetpoints,
	) (TagSetpoints, error)
	GetTagSetpoints(
		ctx context.Context,
		paramID string,
	) (TagSetpoints, error)
	DeleteTagSetpoints(
		ctx context.Context,
		paramID string,
	) error
}

type TagScalingRepository interface {
	UpsertTagScaling(
		ctx context.Context,
		paramID string,
		scaling TagScaling,
	) (TagScaling, error)
	GetTagScaling(
		ctx context.Context,
		paramID string,
	) (TagScaling, error)
	DeleteTagScaling(
		ctx context.Context,
		paramID string,
	) error
}

type ReferenceRepository interface {
	ListDataTypes(
		ctx context.Context,
	) ([]DataType, error)
	ListUnits(
		ctx context.Context,
	) ([]Unit, error)
	ListUnitsByCategory(
		ctx context.Context,
		category string,
	) ([]Unit, error)
}
