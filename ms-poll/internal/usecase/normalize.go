package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

func Normalize(reading domain.Reading, tag *domain.TagSnapshot) domain.Reading {
	if reading.Quality != domain.QualityGood {
		return reading
	}

	normalized := reading
	normalized.RawValue = reading.Value

	castedValue, castError := castByDataType(reading.Value, tag)
	if castError != nil {
		normalized.Quality = domain.QualityBad
		normalized.Error = fmt.Sprintf("cast: %v", castError)
		return normalized
	}
	normalized.Value = castedValue

	if tag == nil || tag.Scaling == nil {
		return normalized
	}

	numericValue, isNumeric := numericToFloat64(normalized.Value)
	if !isNumeric {
		return normalized
	}

	if shouldInterpolate(tag.Scaling) {
		rawMin := *tag.Scaling.RawMin
		rawMax := *tag.Scaling.RawMax
		engMin := *tag.Scaling.EngMin
		engMax := *tag.Scaling.EngMax

		if rawMax != rawMin {
			numericValue = (numericValue-rawMin)*(engMax-engMin)/(rawMax-rawMin) + engMin
		}
	}

	if tag.Scaling.Factor != nil {
		numericValue = numericValue * *tag.Scaling.Factor
	}

	if tag.Scaling.Offset != nil {
		numericValue = numericValue + *tag.Scaling.Offset
	}

	normalized.Value = numericValue
	return normalized
}

type NormalizingPublisher struct {
	next   TagValuesPublisher
	store  *SnapshotStore
	logger *slog.Logger

	warnedTags sync.Map
}

func NewNormalizingPublisher(
	next TagValuesPublisher,
	store *SnapshotStore,
	logger *slog.Logger,
) *NormalizingPublisher {
	if logger == nil {
		logger = slog.Default()
	}

	return &NormalizingPublisher{
		next:   next,
		store:  store,
		logger: logger,
	}
}

func (publisher *NormalizingPublisher) Publish(
	ctx context.Context,
	readings []domain.Reading,
) error {
	if len(readings) == 0 {
		return nil
	}
	if publisher.next == nil {
		return errors.New("next publisher is not configured")
	}
	if publisher.store == nil {
		return errors.New("snapshot store is not configured")
	}

	snapshot := publisher.store.Get()
	if snapshot == nil || len(snapshot.Tags) == 0 {
		return nil
	}

	normalizedReadings := make([]domain.Reading, 0, len(readings))
	for _, reading := range readings {
		tagSnapshot, exists := snapshot.Tags[reading.TagID]
		if !exists || tagSnapshot == nil {
			continue
		}

		publisher.warnZeroDivisionOnce(tagSnapshot)
		normalizedReadings = append(
			normalizedReadings,
			Normalize(reading, tagSnapshot),
		)
	}

	if len(normalizedReadings) == 0 {
		return nil
	}

	return publisher.next.Publish(ctx, normalizedReadings)
}

func (publisher *NormalizingPublisher) warnZeroDivisionOnce(tag *domain.TagSnapshot) {
	if tag == nil || tag.Scaling == nil {
		return
	}
	if !shouldInterpolate(tag.Scaling) {
		return
	}
	if *tag.Scaling.RawMin != *tag.Scaling.RawMax {
		return
	}

	if _, loaded := publisher.warnedTags.LoadOrStore(tag.ID, struct{}{}); loaded {
		return
	}

	publisher.logger.Warn(
		"skip interpolation due to invalid scaling range",
		"method",
		"Normalize",
		"tag_id",
		tag.ID,
		"raw_min",
		*tag.Scaling.RawMin,
		"raw_max",
		*tag.Scaling.RawMax,
	)
}

func shouldInterpolate(scaling *domain.TagScaling) bool {
	if scaling == nil {
		return false
	}

	return scaling.RawMin != nil &&
		scaling.RawMax != nil &&
		scaling.EngMin != nil &&
		scaling.EngMax != nil
}

func castByDataType(value any, tag *domain.TagSnapshot) (any, error) {
	if tag == nil {
		return value, nil
	}

	switch strings.ToLower(strings.TrimSpace(tag.DataType)) {
	case "bool":
		return toBoolValue(value)
	case "int8", "int16", "int32":
		intValue, conversionError := toInt64Value(value)
		if conversionError != nil {
			return nil, conversionError
		}

		switch strings.ToLower(strings.TrimSpace(tag.DataType)) {
		case "int8":
			if intValue < math.MinInt8 || intValue > math.MaxInt8 {
				return nil, fmt.Errorf("value %d is out of range for int8", intValue)
			}
			return int64(int8(intValue)), nil
		case "int16":
			if intValue < math.MinInt16 || intValue > math.MaxInt16 {
				return nil, fmt.Errorf("value %d is out of range for int16", intValue)
			}
			return int64(int16(intValue)), nil
		default:
			if intValue < math.MinInt32 || intValue > math.MaxInt32 {
				return nil, fmt.Errorf("value %d is out of range for int32", intValue)
			}
			return int64(int32(intValue)), nil
		}
	case "uint8", "uint16", "uint32":
		uintValue, conversionError := toUint64Value(value)
		if conversionError != nil {
			return nil, conversionError
		}

		switch strings.ToLower(strings.TrimSpace(tag.DataType)) {
		case "uint8":
			if uintValue > math.MaxUint8 {
				return nil, fmt.Errorf("value %d is out of range for uint8", uintValue)
			}
			return uint64(uint8(uintValue)), nil
		case "uint16":
			if uintValue > math.MaxUint16 {
				return nil, fmt.Errorf("value %d is out of range for uint16", uintValue)
			}
			return uint64(uint16(uintValue)), nil
		default:
			if uintValue > math.MaxUint32 {
				return nil, fmt.Errorf("value %d is out of range for uint32", uintValue)
			}
			return uint64(uint32(uintValue)), nil
		}
	case "float32", "float64":
		return toFloat64Value(value)
	case "string":
		return fmt.Sprintf("%v", value), nil
	default:
		return nil, fmt.Errorf("unsupported data_type: %s", tag.DataType)
	}
}

func numericToFloat64(value any) (float64, bool) {
	switch convertedValue := value.(type) {
	case int:
		return float64(convertedValue), true
	case int8:
		return float64(convertedValue), true
	case int16:
		return float64(convertedValue), true
	case int32:
		return float64(convertedValue), true
	case int64:
		return float64(convertedValue), true
	case uint:
		return float64(convertedValue), true
	case uint8:
		return float64(convertedValue), true
	case uint16:
		return float64(convertedValue), true
	case uint32:
		return float64(convertedValue), true
	case uint64:
		return float64(convertedValue), true
	case float32:
		return float64(convertedValue), true
	case float64:
		return convertedValue, true
	default:
		return 0, false
	}
}

func toBoolValue(value any) (bool, error) {
	switch convertedValue := value.(type) {
	case bool:
		return convertedValue, nil
	case string:
		trimmedValue := strings.TrimSpace(convertedValue)
		if strings.EqualFold(trimmedValue, "true") {
			return true, nil
		}
		if strings.EqualFold(trimmedValue, "false") {
			return false, nil
		}

		floatValue, parseError := strconv.ParseFloat(trimmedValue, 64)
		if parseError != nil {
			return false, fmt.Errorf("cannot cast string %q to bool", convertedValue)
		}
		return floatValue != 0, nil
	default:
		numericValue, isNumeric := numericToFloat64(value)
		if !isNumeric {
			return false, fmt.Errorf("cannot cast %T to bool", value)
		}
		return numericValue != 0, nil
	}
}

func toInt64Value(value any) (int64, error) {
	switch convertedValue := value.(type) {
	case int:
		return int64(convertedValue), nil
	case int8:
		return int64(convertedValue), nil
	case int16:
		return int64(convertedValue), nil
	case int32:
		return int64(convertedValue), nil
	case int64:
		return convertedValue, nil
	case uint:
		if uint64(convertedValue) > math.MaxInt64 {
			return 0, fmt.Errorf("value %d is out of int64 range", convertedValue)
		}
		return int64(convertedValue), nil
	case uint8:
		return int64(convertedValue), nil
	case uint16:
		return int64(convertedValue), nil
	case uint32:
		return int64(convertedValue), nil
	case uint64:
		if convertedValue > math.MaxInt64 {
			return 0, fmt.Errorf("value %d is out of int64 range", convertedValue)
		}
		return int64(convertedValue), nil
	case float32:
		return int64(convertedValue), nil
	case float64:
		return int64(convertedValue), nil
	case string:
		parsedValue, parseError := strconv.ParseInt(strings.TrimSpace(convertedValue), 10, 64)
		if parseError != nil {
			return 0, fmt.Errorf("parse int64 from string: %w", parseError)
		}
		return parsedValue, nil
	default:
		return 0, fmt.Errorf("cannot cast %T to int64", value)
	}
}

func toUint64Value(value any) (uint64, error) {
	switch convertedValue := value.(type) {
	case int:
		if convertedValue < 0 {
			return 0, fmt.Errorf("negative value %d cannot be cast to uint64", convertedValue)
		}
		return uint64(convertedValue), nil
	case int8:
		if convertedValue < 0 {
			return 0, fmt.Errorf("negative value %d cannot be cast to uint64", convertedValue)
		}
		return uint64(convertedValue), nil
	case int16:
		if convertedValue < 0 {
			return 0, fmt.Errorf("negative value %d cannot be cast to uint64", convertedValue)
		}
		return uint64(convertedValue), nil
	case int32:
		if convertedValue < 0 {
			return 0, fmt.Errorf("negative value %d cannot be cast to uint64", convertedValue)
		}
		return uint64(convertedValue), nil
	case int64:
		if convertedValue < 0 {
			return 0, fmt.Errorf("negative value %d cannot be cast to uint64", convertedValue)
		}
		return uint64(convertedValue), nil
	case uint:
		return uint64(convertedValue), nil
	case uint8:
		return uint64(convertedValue), nil
	case uint16:
		return uint64(convertedValue), nil
	case uint32:
		return uint64(convertedValue), nil
	case uint64:
		return convertedValue, nil
	case float32:
		if convertedValue < 0 {
			return 0, fmt.Errorf("negative value %f cannot be cast to uint64", convertedValue)
		}
		return uint64(convertedValue), nil
	case float64:
		if convertedValue < 0 {
			return 0, fmt.Errorf("negative value %f cannot be cast to uint64", convertedValue)
		}
		return uint64(convertedValue), nil
	case string:
		parsedValue, parseError := strconv.ParseUint(strings.TrimSpace(convertedValue), 10, 64)
		if parseError != nil {
			return 0, fmt.Errorf("parse uint64 from string: %w", parseError)
		}
		return parsedValue, nil
	default:
		return 0, fmt.Errorf("cannot cast %T to uint64", value)
	}
}

func toFloat64Value(value any) (float64, error) {
	if numericValue, isNumeric := numericToFloat64(value); isNumeric {
		return numericValue, nil
	}

	stringValue, isString := value.(string)
	if !isString {
		return 0, fmt.Errorf("cannot cast %T to float64", value)
	}

	parsedValue, parseError := strconv.ParseFloat(strings.TrimSpace(stringValue), 64)
	if parseError != nil {
		return 0, fmt.Errorf("parse float64 from string: %w", parseError)
	}

	return parsedValue, nil
}
