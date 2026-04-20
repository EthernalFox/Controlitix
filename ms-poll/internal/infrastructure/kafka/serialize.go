package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

type tagValueMessage struct {
	TagID    string      `json:"tag_id"`
	DeviceID string      `json:"device_id"`
	Value    any         `json:"value"`
	Quality  string      `json:"quality"`
	Error    string      `json:"error"`
	TS       string      `json:"ts"`
}

type rawReadingMessage struct {
	TagID     string `json:"tag_id"`
	DeviceID  string `json:"device_id"`
	RawValue  string `json:"raw_value"`
	ValueType string `json:"value_type"`
	Quality   string `json:"quality"`
	TS        string `json:"ts"`
}

func serializeTagValue(reading domain.Reading) ([]byte, error) {
	message := tagValueMessage{
		TagID:    reading.TagID,
		DeviceID: reading.DeviceID,
		Value:    reading.Value,
		Quality:  string(reading.Quality),
		Error:    reading.Error,
		TS:       reading.ReadAt.UTC().Format(time.RFC3339Nano),
	}
	if reading.Quality != domain.QualityGood {
		message.Value = nil
	}

	payload, marshalError := json.Marshal(message)
	if marshalError != nil {
		return nil, fmt.Errorf("marshal tag value message: %w", marshalError)
	}

	return payload, nil
}

func serializeRawReading(reading domain.Reading) ([]byte, error) {
	message := rawReadingMessage{
		TagID:     reading.TagID,
		DeviceID:  reading.DeviceID,
		RawValue:  fmt.Sprintf("%v", reading.RawValue),
		ValueType: detectValueType(reading.RawValue),
		Quality:   string(reading.Quality),
		TS:        reading.ReadAt.UTC().Format(time.RFC3339Nano),
	}

	payload, marshalError := json.Marshal(message)
	if marshalError != nil {
		return nil, fmt.Errorf("marshal raw reading message: %w", marshalError)
	}

	return payload, nil
}

func detectValueType(value any) string {
	switch value.(type) {
	case bool:
		return "bool"
	case int8:
		return "int8"
	case int16:
		return "int16"
	case int32:
		return "int32"
	case int64, int:
		return "int64"
	case uint8:
		return "uint8"
	case uint16:
		return "uint16"
	case uint32:
		return "uint32"
	case uint64, uint:
		return "uint64"
	case float32:
		return "float32"
	case float64:
		return "float64"
	case string:
		return "string"
	default:
		return "unknown"
	}
}
