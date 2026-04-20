package kafka

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

func TestSerializeTagValueFloatGood(t *testing.T) {
	payload, serializeError := serializeTagValue(domain.Reading{
		TagID:    "tag-1",
		DeviceID: "device-1",
		Value:    50.0,
		Quality:  domain.QualityGood,
		ReadAt:   time.Date(2026, 4, 18, 10, 23, 45, 123000000, time.UTC),
	})
	if serializeError != nil {
		t.Fatalf("unexpected serialize error: %v", serializeError)
	}

	var message map[string]any
	if unmarshalError := json.Unmarshal(payload, &message); unmarshalError != nil {
		t.Fatalf("unmarshal payload: %v", unmarshalError)
	}

	if message["quality"] != "good" {
		t.Fatalf("expected quality good, got %v", message["quality"])
	}

	value, ok := message["value"].(float64)
	if !ok {
		t.Fatalf("expected numeric value, got %T", message["value"])
	}
	if value != 50 {
		t.Fatalf("expected value 50, got %v", value)
	}
}

func TestSerializeTagValueBoolGood(t *testing.T) {
	payload, serializeError := serializeTagValue(domain.Reading{
		TagID:    "tag-1",
		DeviceID: "device-1",
		Value:    true,
		Quality:  domain.QualityGood,
		ReadAt:   time.Now().UTC(),
	})
	if serializeError != nil {
		t.Fatalf("unexpected serialize error: %v", serializeError)
	}

	var message map[string]any
	if unmarshalError := json.Unmarshal(payload, &message); unmarshalError != nil {
		t.Fatalf("unmarshal payload: %v", unmarshalError)
	}

	value, ok := message["value"].(bool)
	if !ok {
		t.Fatalf("expected bool value, got %T", message["value"])
	}
	if !value {
		t.Fatal("expected true value")
	}
}

func TestSerializeTagValueBad(t *testing.T) {
	payload, serializeError := serializeTagValue(domain.Reading{
		TagID:    "tag-1",
		DeviceID: "device-1",
		Value:    15.0,
		Quality:  domain.QualityBad,
		Error:    "timeout",
		ReadAt:   time.Now().UTC(),
	})
	if serializeError != nil {
		t.Fatalf("unexpected serialize error: %v", serializeError)
	}

	var message map[string]any
	if unmarshalError := json.Unmarshal(payload, &message); unmarshalError != nil {
		t.Fatalf("unmarshal payload: %v", unmarshalError)
	}

	if message["value"] != nil {
		t.Fatalf("expected nil value for bad quality, got %v", message["value"])
	}
	if message["quality"] != "bad" {
		t.Fatalf("expected quality bad, got %v", message["quality"])
	}
	if message["error"] != "timeout" {
		t.Fatalf("expected timeout error, got %v", message["error"])
	}
}
