package usecase

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

func TestNormalize(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name     string
		reading  domain.Reading
		tag      *domain.TagSnapshot
		validate func(t *testing.T, got domain.Reading)
	}{
		{
			name: "cast int16 to int64",
			reading: domain.Reading{
				TagID:   "tag-1",
				Value:   int16(42),
				Quality: domain.QualityGood,
				ReadAt:  now,
			},
			tag: &domain.TagSnapshot{
				ID:       "tag-1",
				DataType: "int16",
			},
			validate: func(t *testing.T, got domain.Reading) {
				t.Helper()
				value, ok := got.Value.(int64)
				if !ok {
					t.Fatalf("expected int64 value, got %T", got.Value)
				}
				if value != 42 {
					t.Fatalf("expected value 42, got %d", value)
				}
			},
		},
		{
			name: "cast float32 to float64 keeps raw value",
			reading: domain.Reading{
				TagID:   "tag-1",
				Value:   float32(1.5),
				Quality: domain.QualityGood,
				ReadAt:  now,
			},
			tag: &domain.TagSnapshot{
				ID:       "tag-1",
				DataType: "float32",
			},
			validate: func(t *testing.T, got domain.Reading) {
				t.Helper()
				value, ok := got.Value.(float64)
				if !ok {
					t.Fatalf("expected float64 value, got %T", got.Value)
				}
				if math.Abs(value-1.5) > 1e-9 {
					t.Fatalf("expected value 1.5, got %f", value)
				}

				rawValue, ok := got.RawValue.(float32)
				if !ok {
					t.Fatalf("expected raw float32 value, got %T", got.RawValue)
				}
				if rawValue != float32(1.5) {
					t.Fatalf("expected raw value 1.5, got %f", rawValue)
				}
			},
		},
		{
			name: "scaling interpolation",
			reading: domain.Reading{
				TagID:   "tag-1",
				Value:   int64(2048),
				Quality: domain.QualityGood,
				ReadAt:  now,
			},
			tag: &domain.TagSnapshot{
				ID:       "tag-1",
				DataType: "int32",
				Scaling: &domain.TagScaling{
					RawMin: float64Pointer(0),
					RawMax: float64Pointer(4095),
					EngMin: float64Pointer(0),
					EngMax: float64Pointer(100),
				},
			},
			validate: func(t *testing.T, got domain.Reading) {
				t.Helper()
				value, ok := got.Value.(float64)
				if !ok {
					t.Fatalf("expected float64 value, got %T", got.Value)
				}
				if math.Abs(value-50.0) > 0.05 {
					t.Fatalf("expected value close to 50.0, got %f", value)
				}
			},
		},
		{
			name: "scaling factor and offset",
			reading: domain.Reading{
				TagID:   "tag-1",
				Value:   int64(850),
				Quality: domain.QualityGood,
				ReadAt:  now,
			},
			tag: &domain.TagSnapshot{
				ID:       "tag-1",
				DataType: "int32",
				Scaling: &domain.TagScaling{
					Factor: float64Pointer(0.1),
					Offset: float64Pointer(-40),
				},
			},
			validate: func(t *testing.T, got domain.Reading) {
				t.Helper()
				value, ok := got.Value.(float64)
				if !ok {
					t.Fatalf("expected float64 value, got %T", got.Value)
				}
				if math.Abs(value-45) > 1e-9 {
					t.Fatalf("expected value 45, got %f", value)
				}
			},
		},
		{
			name: "cast bad marks reading as bad",
			reading: domain.Reading{
				TagID:   "tag-1",
				Value:   "abc",
				Quality: domain.QualityGood,
				ReadAt:  now,
			},
			tag: &domain.TagSnapshot{
				ID:       "tag-1",
				DataType: "int16",
			},
			validate: func(t *testing.T, got domain.Reading) {
				t.Helper()
				if got.Quality != domain.QualityBad {
					t.Fatalf("expected bad quality, got %s", got.Quality)
				}
				if !strings.Contains(strings.ToLower(got.Error), "cast") {
					t.Fatalf("expected cast error, got %q", got.Error)
				}
			},
		},
		{
			name: "bad quality passes through unchanged",
			reading: domain.Reading{
				TagID:   "tag-1",
				Value:   "timeout",
				Quality: domain.QualityBad,
				Error:   "timeout",
				ReadAt:  now,
			},
			tag: &domain.TagSnapshot{
				ID:       "tag-1",
				DataType: "float64",
			},
			validate: func(t *testing.T, got domain.Reading) {
				t.Helper()
				if got.Quality != domain.QualityBad {
					t.Fatalf("expected bad quality, got %s", got.Quality)
				}
				if got.Error != "timeout" {
					t.Fatalf("expected error timeout, got %q", got.Error)
				}
				if got.Value != "timeout" {
					t.Fatalf("expected unchanged value, got %v", got.Value)
				}
			},
		},
		{
			name: "without scaling only cast is applied",
			reading: domain.Reading{
				TagID:   "tag-1",
				Value:   uint16(2048),
				Quality: domain.QualityGood,
				ReadAt:  now,
			},
			tag: &domain.TagSnapshot{
				ID:       "tag-1",
				DataType: "uint16",
				Scaling:  nil,
			},
			validate: func(t *testing.T, got domain.Reading) {
				t.Helper()
				value, ok := got.Value.(uint64)
				if !ok {
					t.Fatalf("expected uint64 value, got %T", got.Value)
				}
				if value != 2048 {
					t.Fatalf("expected value 2048, got %d", value)
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := Normalize(testCase.reading, testCase.tag)
			testCase.validate(t, got)
		})
	}
}

func float64Pointer(value float64) *float64 {
	return &value
}
