package domain

import (
	"encoding/json"
	"strings"
	"time"
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

var SupportedFigureTypes = []FigureType{
	FigureTypeRect,
	FigureTypeCircle,
	FigureTypeEllipse,
	FigureTypeWedge,
	FigureTypeLine,
	FigureTypeImage,
	FigureTypeText,
	FigureTypeRing,
	FigureTypeArc,
	FigureTypeTag,
	FigureTypePath,
}

func ParseFigureType(value string) (FigureType, bool) {
	normalizedValue := FigureType(strings.TrimSpace(strings.ToLower(value)))
	for _, figureType := range SupportedFigureTypes {
		if figureType == normalizedValue {
			return figureType, true
		}
	}

	return "", false
}

type Figure struct {
	ID         string
	DiagramID  string
	TagID      *string
	FigureType FigureType
	Parameters json.RawMessage
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
