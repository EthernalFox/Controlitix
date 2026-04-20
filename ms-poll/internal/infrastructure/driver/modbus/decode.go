package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

func DecodeRegisters(buf []byte, dataType string) (any, error) {
	normalizedDataType := strings.ToLower(strings.TrimSpace(dataType))

	switch normalizedDataType {
	case "bool":
		registerValue, decodeError := decodeUint16(buf)
		if decodeError != nil {
			return nil, decodeError
		}
		return (registerValue & 1) == 1, nil
	case "int8":
		registerValue, decodeError := decodeUint16(buf)
		if decodeError != nil {
			return nil, decodeError
		}
		return int64(int8(byte(registerValue))), nil
	case "uint8":
		registerValue, decodeError := decodeUint16(buf)
		if decodeError != nil {
			return nil, decodeError
		}
		return uint64(uint8(registerValue)), nil
	case "int16":
		registerValue, decodeError := decodeUint16(buf)
		if decodeError != nil {
			return nil, decodeError
		}
		return int64(int16(registerValue)), nil
	case "uint16":
		registerValue, decodeError := decodeUint16(buf)
		if decodeError != nil {
			return nil, decodeError
		}
		return uint64(registerValue), nil
	case "int32":
		value, decodeError := decodeUint32(buf)
		if decodeError != nil {
			return nil, decodeError
		}
		return int64(int32(value)), nil
	case "uint32":
		value, decodeError := decodeUint32(buf)
		if decodeError != nil {
			return nil, decodeError
		}
		return uint64(value), nil
	case "float32":
		value, decodeError := decodeUint32(buf)
		if decodeError != nil {
			return nil, decodeError
		}
		return float64(math.Float32frombits(value)), nil
	case "float64":
		value, decodeError := decodeUint64(buf)
		if decodeError != nil {
			return nil, decodeError
		}
		return math.Float64frombits(value), nil
	case "string":
		trimmedValue := strings.TrimRight(string(buf), "\x00")
		return trimmedValue, nil
	default:
		return nil, fmt.Errorf("unsupported data_type: %s", dataType)
	}
}

func DecodeBits(buf []byte, bitIndex int) bool {
	if bitIndex < 0 {
		return false
	}

	byteIndex := bitIndex / 8
	if byteIndex >= len(buf) {
		return false
	}

	bitOffset := uint(bitIndex % 8)
	return ((buf[byteIndex] >> bitOffset) & 1) == 1
}

func decodeUint16(buf []byte) (uint16, error) {
	if len(buf) < 2 {
		return 0, fmt.Errorf("not enough bytes: expected 2, got %d", len(buf))
	}
	return binary.BigEndian.Uint16(buf[:2]), nil
}

func decodeUint32(buf []byte) (uint32, error) {
	if len(buf) < 4 {
		return 0, fmt.Errorf("not enough bytes: expected 4, got %d", len(buf))
	}
	return binary.BigEndian.Uint32(buf[:4]), nil
}

func decodeUint64(buf []byte) (uint64, error) {
	if len(buf) < 8 {
		return 0, fmt.Errorf("not enough bytes: expected 8, got %d", len(buf))
	}
	return binary.BigEndian.Uint64(buf[:8]), nil
}
