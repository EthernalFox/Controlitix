package modbus

import (
	"fmt"
	"sort"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

const (
	maxRegisterBatchQuantity = 125
	maxBitBatchQuantity      = 2000
)

type ReadRequest struct {
	RegisterType RegisterType
	StartOffset  uint16
	Quantity     uint16
	Tags         []*tagSlot
}

type tagSlot struct {
	Tag    *domain.TagSnapshot
	Offset uint16
	Length uint16
}

type parsedTag struct {
	tag     *domain.TagSnapshot
	address ParsedAddress
}

func BuildBatches(tags []*domain.TagSnapshot) ([]ReadRequest, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	parsedByType := make(map[RegisterType][]parsedTag)
	for _, tag := range tags {
		if tag == nil {
			return nil, fmt.Errorf("tag is nil")
		}

		parsedAddress, parseError := ParseTagAddress(tag.Address, tag.DataType)
		if parseError != nil {
			return nil, fmt.Errorf(
				"parse address for tag %s: %w",
				tag.ID,
				parseError,
			)
		}

		parsedByType[parsedAddress.RegisterType] = append(
			parsedByType[parsedAddress.RegisterType],
			parsedTag{
				tag:     tag,
				address: parsedAddress,
			},
		)
	}

	registerOrder := []RegisterType{
		RegHolding,
		RegInput,
		RegCoil,
		RegDiscrete,
	}

	batches := make([]ReadRequest, 0)
	for _, registerType := range registerOrder {
		parsedTags := parsedByType[registerType]
		if len(parsedTags) == 0 {
			continue
		}

		sort.Slice(parsedTags, func(left, right int) bool {
			return parsedTags[left].address.Offset < parsedTags[right].address.Offset
		})

		typeBatches, buildError := buildBatchesByRegisterType(
			registerType,
			parsedTags,
		)
		if buildError != nil {
			return nil, buildError
		}

		batches = append(batches, typeBatches...)
	}

	return batches, nil
}

func buildBatchesByRegisterType(
	registerType RegisterType,
	parsedTags []parsedTag,
) ([]ReadRequest, error) {
	maxQuantity := maxQuantityFor(registerType)
	if maxQuantity <= 0 {
		return nil, fmt.Errorf(
			"unsupported register_type for batching: %s",
			registerType,
		)
	}

	batches := make([]ReadRequest, 0)
	var currentBatch *ReadRequest

	for _, currentTag := range parsedTags {
		if currentTag.address.Length == 0 {
			return nil, fmt.Errorf(
				"tag %s has zero length address",
				currentTag.tag.ID,
			)
		}

		if currentTag.address.Length > maxQuantity {
			return nil, fmt.Errorf(
				"tag %s length %d exceeds max quantity %d for register_type %s",
				currentTag.tag.ID,
				currentTag.address.Length,
				maxQuantity,
				registerType,
			)
		}

		if currentBatch == nil {
			currentBatch = startBatch(registerType, currentTag)
			continue
		}

		currentStart := uint32(currentBatch.StartOffset)
		currentEnd := currentStart + uint32(currentBatch.Quantity)

		tagStart := uint32(currentTag.address.Offset)
		tagEnd := tagStart + uint32(currentTag.address.Length)

		if tagStart > currentEnd {
			batches = append(batches, *currentBatch)
			currentBatch = startBatch(registerType, currentTag)
			continue
		}

		nextEnd := maxUint32(currentEnd, tagEnd)
		nextQuantity := nextEnd - currentStart
		if nextQuantity > uint32(maxQuantity) {
			batches = append(batches, *currentBatch)
			currentBatch = startBatch(registerType, currentTag)
			continue
		}

		currentBatch.Quantity = uint16(nextQuantity)
		currentBatch.Tags = append(currentBatch.Tags, &tagSlot{
			Tag:    currentTag.tag,
			Offset: uint16(tagStart - currentStart),
			Length: currentTag.address.Length,
		})
	}

	if currentBatch != nil {
		batches = append(batches, *currentBatch)
	}

	return batches, nil
}

func startBatch(registerType RegisterType, tag parsedTag) *ReadRequest {
	return &ReadRequest{
		RegisterType: registerType,
		StartOffset:  tag.address.Offset,
		Quantity:     tag.address.Length,
		Tags: []*tagSlot{
			{
				Tag:    tag.tag,
				Offset: 0,
				Length: tag.address.Length,
			},
		},
	}
}

func maxQuantityFor(registerType RegisterType) uint16 {
	switch registerType {
	case RegHolding, RegInput:
		return maxRegisterBatchQuantity
	case RegCoil, RegDiscrete:
		return maxBitBatchQuantity
	default:
		return 0
	}
}

func maxUint32(left, right uint32) uint32 {
	if left > right {
		return left
	}
	return right
}
