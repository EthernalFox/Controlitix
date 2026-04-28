package http

import (
	"encoding/json"
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type objectListItemResponse struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	PublishedDiagramCount int    `json:"published_diagram_count"`
	FirstPublishedDiagramID string `json:"first_published_diagram_id"`
}

type objectListResponse struct {
	Items  []objectListItemResponse `json:"items"`
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
}

type diagramListItemResponse struct {
	ID            string `json:"id"`
	ObjectID      string `json:"object_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	PublishedAt   string `json:"published_at"`
	FigureCount   int    `json:"figure_count"`
	BoundTagCount int    `json:"bound_tag_count"`
}

type diagramListResponse struct {
	Items  []diagramListItemResponse `json:"items"`
	Total  int                       `json:"total"`
	Offset int                       `json:"offset"`
	Limit  int                       `json:"limit"`
}

type diagramCanvasResponse struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Background string `json:"background"`
}

type diagramTagResponse struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	DeviceID   string                `json:"device_id"`
	DeviceName string                `json:"device_name"`
	Unit       trendUnitResponse     `json:"unit"`
	DataType   trendDataTypeResponse `json:"data_type"`
}

type diagramFigureResponse struct {
	ID     string          `json:"id"`
	Type   string          `json:"type"`
	TagID  *string         `json:"tag_id"`
	Params json.RawMessage `json:"params"`
	Tag    *diagramTagResponse `json:"tag,omitempty"`
}

type diagramResponse struct {
	ID          string                `json:"id"`
	ObjectID    string                `json:"object_id"`
	ObjectName  string                `json:"object_name"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	PublishedAt string                `json:"published_at"`
	Canvas      diagramCanvasResponse `json:"canvas"`
	Figures     []diagramFigureResponse `json:"figures"`
}

type diagramSnapshotValueResponse struct {
	TagID string   `json:"tag_id"`
	TS    string   `json:"ts"`
	V     *float64 `json:"v"`
	Q     string   `json:"q"`
}

type diagramSnapshotResponse struct {
	DiagramID     string                         `json:"diagram_id"`
	TS            string                         `json:"ts"`
	Values        []diagramSnapshotValueResponse `json:"values"`
	MissingTagIDs []string                       `json:"missing_tag_ids"`
}

func mapObjectListResponse(result domain.ObjectListResult) objectListResponse {
	items := make([]objectListItemResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, objectListItemResponse{
			ID:                    item.ID.String(),
			Name:                  item.Name,
			Description:           item.Description,
			PublishedDiagramCount: item.PublishedDiagramCount,
			FirstPublishedDiagramID: item.FirstPublishedDiagramID.String(),
		})
	}

	return objectListResponse{
		Items:  items,
		Total:  result.Total,
		Offset: result.Offset,
		Limit:  result.Limit,
	}
}

func mapDiagramListResponse(result domain.DiagramListResult) diagramListResponse {
	items := make([]diagramListItemResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, diagramListItemResponse{
			ID:            item.ID.String(),
			ObjectID:      item.ObjectID.String(),
			Name:          item.Name,
			Description:   item.Description,
			PublishedAt:   item.PublishedAt.UTC().Format(time.RFC3339Nano),
			FigureCount:   item.FigureCount,
			BoundTagCount: item.BoundTagCount,
		})
	}

	return diagramListResponse{
		Items:  items,
		Total:  result.Total,
		Offset: result.Offset,
		Limit:  result.Limit,
	}
}

func mapDiagramResponse(diagram domain.Diagram) diagramResponse {
	figures := make([]diagramFigureResponse, 0, len(diagram.Figures))
	for _, figure := range diagram.Figures {
		var tagID *string
		if figure.TagID != nil {
			tagIDValue := figure.TagID.String()
			tagID = &tagIDValue
		}

		var tag *diagramTagResponse
		if figure.Tag != nil {
			tag = &diagramTagResponse{
				ID:         figure.Tag.ID.String(),
				Name:       figure.Tag.Name,
				DeviceID:   figure.Tag.DeviceID.String(),
				DeviceName: figure.Tag.DeviceName,
				Unit: trendUnitResponse{
					ID:       figure.Tag.Unit.ID,
					Name:     figure.Tag.Unit.Name,
					Symbol:   figure.Tag.Unit.Symbol,
					Category: figure.Tag.Unit.Category,
				},
				DataType: trendDataTypeResponse{
					ID:   figure.Tag.DataType.ID,
					Name: figure.Tag.DataType.Name,
				},
			}
		}

		figures = append(figures, diagramFigureResponse{
			ID:     figure.ID.String(),
			Type:   string(figure.Type),
			TagID:  tagID,
			Params: figure.Params,
			Tag:    tag,
		})
	}

	return diagramResponse{
		ID:          diagram.ID.String(),
		ObjectID:    diagram.ObjectID.String(),
		ObjectName:  diagram.ObjectName,
		Name:        diagram.Name,
		Description: diagram.Description,
		PublishedAt: diagram.PublishedAt.UTC().Format(time.RFC3339Nano),
		Canvas: diagramCanvasResponse{
			Width:      diagram.Canvas.Width,
			Height:     diagram.Canvas.Height,
			Background: diagram.Canvas.Background,
		},
		Figures: figures,
	}
}

func mapDiagramSnapshotResponse(snapshot domain.DiagramSnapshot) diagramSnapshotResponse {
	values := make([]diagramSnapshotValueResponse, 0, len(snapshot.Values))
	for _, record := range snapshot.Values {
		values = append(values, diagramSnapshotValueResponse{
			TagID: record.TagID.String(),
			TS:    record.Timestamp.UTC().Format(time.RFC3339Nano),
			V:     record.Value,
			Q:     string(record.Quality),
		})
	}

	missingTagIDs := make([]string, 0, len(snapshot.MissingTagIDs))
	for _, tagID := range snapshot.MissingTagIDs {
		missingTagIDs = append(missingTagIDs, tagID.String())
	}

	return diagramSnapshotResponse{
		DiagramID:     snapshot.DiagramID.String(),
		TS:            snapshot.TS.UTC().Format(time.RFC3339Nano),
		Values:        values,
		MissingTagIDs: missingTagIDs,
	}
}
