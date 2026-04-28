package http

import (
	"fmt"
	"time"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type trendPointResponse struct {
	TS string   `json:"ts"`
	V  *float64 `json:"v"`
	Q  string   `json:"q"`
}

type trendUnitResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Category string `json:"category"`
}

type trendDataTypeResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type trendSeriesResponse struct {
	TagID      string               `json:"tag_id"`
	TagName    string               `json:"tag_name"`
	DeviceID   string               `json:"device_id"`
	DeviceName string               `json:"device_name"`
	Unit       trendUnitResponse    `json:"unit"`
	DataType   trendDataTypeResponse `json:"data_type"`
	From       string               `json:"from"`
	To         string               `json:"to"`
	Step       string               `json:"step"`
	Agg        string               `json:"agg"`
	Source     string               `json:"source"`
	Points     []trendPointResponse `json:"points"`
}

type trendBatchErrorResponse struct {
	TagID   string  `json:"tag_id"`
	Problem Problem `json:"problem"`
}

type trendsBatchResponse struct {
	Series []trendSeriesResponse     `json:"series"`
	Errors []trendBatchErrorResponse `json:"errors,omitempty"`
}

type tagsItemResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	UnitSymbol string `json:"unit_symbol"`
}

type tagsResponse struct {
	Items []tagsItemResponse `json:"items"`
	Total int                `json:"total"`
}

func mapTrendSeriesResponse(series domain.TrendSeries) trendSeriesResponse {
	points := make([]trendPointResponse, 0, len(series.Points))
	for _, point := range series.Points {
		points = append(points, trendPointResponse{
			TS: point.Timestamp.UTC().Format(time.RFC3339Nano),
			V:  point.Value,
			Q:  string(point.Quality),
		})
	}

	return trendSeriesResponse{
		TagID:      series.TagID.String(),
		TagName:    series.TagName,
		DeviceID:   series.DeviceID.String(),
		DeviceName: series.DeviceName,
		Unit: trendUnitResponse{
			ID:       series.Unit.ID,
			Name:     series.Unit.Name,
			Symbol:   series.Unit.Symbol,
			Category: series.Unit.Category,
		},
		DataType: trendDataTypeResponse{
			ID:   series.DataType.ID,
			Name: series.DataType.Name,
		},
		From:   series.From.UTC().Format(time.RFC3339Nano),
		To:     series.To.UTC().Format(time.RFC3339Nano),
		Step:   formatStep(series.Step),
		Agg:    string(series.Aggregator),
		Source: string(series.Source),
		Points: points,
	}
}

func mapTrendsBatchResponse(result domain.TrendBatchResult) trendsBatchResponse {
	series := make([]trendSeriesResponse, 0, len(result.Series))
	for _, item := range result.Series {
		series = append(series, mapTrendSeriesResponse(item))
	}

	responseErrors := make([]trendBatchErrorResponse, 0, len(result.Errors))
	for _, item := range result.Errors {
		responseErrors = append(responseErrors, trendBatchErrorResponse{
			TagID: item.TagID.String(),
			Problem: Problem{
				Type:   item.Problem.Type,
				Title:  item.Problem.Title,
				Status: item.Problem.Status,
				Detail: item.Problem.Detail,
			},
		})
	}

	return trendsBatchResponse{
		Series: series,
		Errors: responseErrors,
	}
}

func mapTagsResponse(result domain.TagSearchResult) tagsResponse {
	items := make([]tagsItemResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, tagsItemResponse{
			ID:         item.ID.String(),
			Name:       item.Name,
			DeviceID:   item.DeviceID.String(),
			DeviceName: item.DeviceName,
			UnitSymbol: item.UnitSymbol,
		})
	}

	return tagsResponse{
		Items: items,
		Total: result.Total,
	}
}

func formatStep(step time.Duration) string {
	if step <= 0 {
		return "0s"
	}
	if step%time.Hour == 0 {
		return fmt.Sprintf("%dh", int64(step/time.Hour))
	}
	if step%time.Minute == 0 {
		return fmt.Sprintf("%dm", int64(step/time.Minute))
	}
	if step%time.Second == 0 {
		return fmt.Sprintf("%ds", int64(step/time.Second))
	}
	return step.String()
}
