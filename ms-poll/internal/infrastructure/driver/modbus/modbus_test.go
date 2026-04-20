package modbus

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

func TestParseTagAddress(t *testing.T) {
	tests := []struct {
		name             string
		rawAddress       json.RawMessage
		dataType         string
		expectedType     RegisterType
		expectedOffset   uint16
		expectedLength   uint16
		expectParseError bool
	}{
		{
			name:           "holding bool",
			rawAddress:     rawAddress("holding", 40001, intPointer(1)),
			dataType:       "bool",
			expectedType:   RegHolding,
			expectedOffset: 0,
			expectedLength: 1,
		},
		{
			name:           "input uint8",
			rawAddress:     rawAddress("input", 30002, intPointer(1)),
			dataType:       "uint8",
			expectedType:   RegInput,
			expectedOffset: 1,
			expectedLength: 1,
		},
		{
			name:           "coil bool",
			rawAddress:     rawAddress("coil", 1, intPointer(1)),
			dataType:       "bool",
			expectedType:   RegCoil,
			expectedOffset: 0,
			expectedLength: 1,
		},
		{
			name:           "discrete bool",
			rawAddress:     rawAddress("discrete", 10001, intPointer(1)),
			dataType:       "bool",
			expectedType:   RegDiscrete,
			expectedOffset: 0,
			expectedLength: 1,
		},
		{
			name:           "holding int32 length 2",
			rawAddress:     rawAddress("holding", 40020, intPointer(2)),
			dataType:       "int32",
			expectedType:   RegHolding,
			expectedOffset: 19,
			expectedLength: 2,
		},
		{
			name:           "input float64 length 4",
			rawAddress:     rawAddress("input", 30010, intPointer(4)),
			dataType:       "float64",
			expectedType:   RegInput,
			expectedOffset: 9,
			expectedLength: 4,
		},
		{
			name:           "holding string with explicit length",
			rawAddress:     rawAddress("holding", 40030, intPointer(8)),
			dataType:       "string",
			expectedType:   RegHolding,
			expectedOffset: 29,
			expectedLength: 8,
		},
		{
			name:             "coil with non bool data_type is invalid",
			rawAddress:       rawAddress("coil", 1, intPointer(1)),
			dataType:         "uint16",
			expectParseError: true,
		},
		{
			name:             "holding string without length is invalid",
			rawAddress:       rawAddress("holding", 40030, nil),
			dataType:         "string",
			expectParseError: true,
		},
		{
			name:             "holding with out of range address",
			rawAddress:       rawAddress("holding", 50000, intPointer(1)),
			dataType:         "uint16",
			expectParseError: true,
		},
		{
			name:             "unsupported register type",
			rawAddress:       rawAddress("custom", 40001, intPointer(1)),
			dataType:         "uint16",
			expectParseError: true,
		},
		{
			name:             "invalid length for float32",
			rawAddress:       rawAddress("holding", 40050, intPointer(1)),
			dataType:         "float32",
			expectParseError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			parsedAddress, parseError := ParseTagAddress(
				testCase.rawAddress,
				testCase.dataType,
			)

			if testCase.expectParseError {
				if parseError == nil {
					t.Fatal("expected parse error, got nil")
				}
				return
			}

			if parseError != nil {
				t.Fatalf("unexpected parse error: %v", parseError)
			}

			if parsedAddress.RegisterType != testCase.expectedType {
				t.Fatalf(
					"unexpected register type: expected %s, got %s",
					testCase.expectedType,
					parsedAddress.RegisterType,
				)
			}
			if parsedAddress.Offset != testCase.expectedOffset {
				t.Fatalf(
					"unexpected offset: expected %d, got %d",
					testCase.expectedOffset,
					parsedAddress.Offset,
				)
			}
			if parsedAddress.Length != testCase.expectedLength {
				t.Fatalf(
					"unexpected length: expected %d, got %d",
					testCase.expectedLength,
					parsedAddress.Length,
				)
			}
		})
	}
}

func TestDecodeRegisters(t *testing.T) {
	tests := []struct {
		name          string
		buffer        []byte
		dataType      string
		expectedValue any
		floatDelta    float64
	}{
		{
			name:          "uint16",
			buffer:        []byte{0x41, 0x48},
			dataType:      "uint16",
			expectedValue: uint64(16712),
		},
		{
			name:          "float32",
			buffer:        []byte{0x40, 0x49, 0x0F, 0xDB},
			dataType:      "float32",
			expectedValue: 3.14159,
			floatDelta:    1e-5,
		},
		{
			name:          "float64",
			buffer:        []byte{0x40, 0x09, 0x21, 0xFB, 0x54, 0x44, 0x2D, 0x18},
			dataType:      "float64",
			expectedValue: 3.14159265358979,
			floatDelta:    1e-12,
		},
		{
			name:          "int16 negative one",
			buffer:        []byte{0xFF, 0xFF},
			dataType:      "int16",
			expectedValue: int64(-1),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			decodedValue, decodeError := DecodeRegisters(
				testCase.buffer,
				testCase.dataType,
			)
			if decodeError != nil {
				t.Fatalf("unexpected decode error: %v", decodeError)
			}

			switch expectedValue := testCase.expectedValue.(type) {
			case float64:
				value, isFloat := decodedValue.(float64)
				if !isFloat {
					t.Fatalf("decoded value type mismatch: got %T", decodedValue)
				}
				if math.Abs(value-expectedValue) > testCase.floatDelta {
					t.Fatalf(
						"decoded float mismatch: expected %f got %f",
						expectedValue,
						value,
					)
				}
			default:
				if decodedValue != testCase.expectedValue {
					t.Fatalf(
						"decoded value mismatch: expected %v got %v",
						testCase.expectedValue,
						decodedValue,
					)
				}
			}
		})
	}
}

func TestDecodeBits(t *testing.T) {
	buffer := []byte{0b00001010}
	if !DecodeBits(buffer, 1) {
		t.Fatal("expected bit index 1 to be true")
	}
	if DecodeBits(buffer, 0) {
		t.Fatal("expected bit index 0 to be false")
	}
}

func TestBuildBatches(t *testing.T) {
	t.Run("three contiguous holding tags in one batch", func(t *testing.T) {
		tags := []*domain.TagSnapshot{
			testTag("tag-1", "uint16", rawAddress("holding", 40001, intPointer(1))),
			testTag("tag-2", "uint16", rawAddress("holding", 40002, intPointer(1))),
			testTag("tag-3", "uint16", rawAddress("holding", 40003, intPointer(1))),
		}

		batches, buildError := BuildBatches(tags)
		if buildError != nil {
			t.Fatalf("unexpected build error: %v", buildError)
		}
		if len(batches) != 1 {
			t.Fatalf("expected one batch, got %d", len(batches))
		}
		if batches[0].Quantity != 3 {
			t.Fatalf("expected quantity 3, got %d", batches[0].Quantity)
		}
	})

	t.Run("holding tags with large gap split into two batches", func(t *testing.T) {
		tags := []*domain.TagSnapshot{
			testTag("tag-1", "uint16", rawAddress("holding", 40001, intPointer(1))),
			testTag("tag-2", "uint16", rawAddress("holding", 40100, intPointer(1))),
		}

		batches, buildError := BuildBatches(tags)
		if buildError != nil {
			t.Fatalf("unexpected build error: %v", buildError)
		}
		if len(batches) != 2 {
			t.Fatalf("expected two batches, got %d", len(batches))
		}
	})

	t.Run("130 contiguous tags split by max quantity", func(t *testing.T) {
		tags := make([]*domain.TagSnapshot, 0, 130)
		for index := 0; index < 130; index++ {
			tags = append(tags, testTag(
				fmt.Sprintf("tag-%d", index),
				"uint16",
				rawAddress("holding", 40001+index, intPointer(1)),
			))
		}

		batches, buildError := BuildBatches(tags)
		if buildError != nil {
			t.Fatalf("unexpected build error: %v", buildError)
		}
		if len(batches) != 2 {
			t.Fatalf("expected two batches, got %d", len(batches))
		}
		for _, batch := range batches {
			if batch.Quantity > 125 {
				t.Fatalf("batch quantity exceeds 125: %d", batch.Quantity)
			}
		}
	})

	t.Run("holding and coil tags produce separate batches", func(t *testing.T) {
		tags := []*domain.TagSnapshot{
			testTag("tag-1", "uint16", rawAddress("holding", 40001, intPointer(1))),
			testTag("tag-2", "bool", rawAddress("coil", 1, intPointer(1))),
		}

		batches, buildError := BuildBatches(tags)
		if buildError != nil {
			t.Fatalf("unexpected build error: %v", buildError)
		}
		if len(batches) != 2 {
			t.Fatalf("expected two batches, got %d", len(batches))
		}
	})

	t.Run("invalid address returns error", func(t *testing.T) {
		tags := []*domain.TagSnapshot{
			testTag("tag-1", "uint16", rawAddress("holding", 60001, intPointer(1))),
		}

		_, buildError := BuildBatches(tags)
		if buildError == nil {
			t.Fatal("expected build error for invalid address")
		}
	})
}

func rawAddress(
	registerType string,
	address int,
	length *int,
) json.RawMessage {
	if length == nil {
		return json.RawMessage(fmt.Sprintf(
			`{"register_type":"%s","address":%d}`,
			registerType,
			address,
		))
	}

	return json.RawMessage(fmt.Sprintf(
		`{"register_type":"%s","address":%d,"length":%d}`,
		registerType,
		address,
		*length,
	))
}

func testTag(
	id string,
	dataType string,
	address json.RawMessage,
) *domain.TagSnapshot {
	return &domain.TagSnapshot{
		ID:       id,
		DeviceID: "device-1",
		DataType: dataType,
		Address:  address,
	}
}

func intPointer(value int) *int {
	return &value
}
