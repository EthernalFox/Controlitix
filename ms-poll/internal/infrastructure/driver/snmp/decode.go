package snmp

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
)

func DecodePDU(pdu gosnmp.SnmpPDU, dataType string) (any, error) {
	baseValue, decodeError := decodeBaseValue(pdu, dataType)
	if decodeError != nil {
		return nil, decodeError
	}

	return castToDataType(baseValue, dataType)
}

func decodeBaseValue(pdu gosnmp.SnmpPDU, dataType string) (any, error) {
	switch pdu.Type {
	case gosnmp.Integer:
		return toInt64(pdu.Value)
	case gosnmp.Counter32, gosnmp.Gauge32, gosnmp.TimeTicks, gosnmp.Counter64:
		return toUint64(pdu.Value)
	case gosnmp.OctetString:
		octets, ok := pdu.Value.([]byte)
		if !ok {
			return nil, fmt.Errorf("snmp octet string has invalid type %T", pdu.Value)
		}

		stringValue := strings.TrimRight(string(octets), "\x00")
		if strings.EqualFold(strings.TrimSpace(dataType), "string") {
			return stringValue, nil
		}

		if number, parseError := parseStringToNumber(stringValue); parseError == nil {
			return number, nil
		}

		return nil, fmt.Errorf("snmp octet string is not numeric: %q", stringValue)
	case gosnmp.ObjectIdentifier:
		oid, ok := pdu.Value.(string)
		if !ok {
			return nil, fmt.Errorf("snmp object identifier has invalid type %T", pdu.Value)
		}
		return oid, nil
	case gosnmp.OpaqueFloat:
		switch value := pdu.Value.(type) {
		case float32:
			return float64(value), nil
		case float64:
			return value, nil
		default:
			return nil, fmt.Errorf("snmp opaque float has invalid type %T", pdu.Value)
		}
	case gosnmp.OpaqueDouble:
		switch value := pdu.Value.(type) {
		case float64:
			return value, nil
		case float32:
			return float64(value), nil
		default:
			return nil, fmt.Errorf("snmp opaque double has invalid type %T", pdu.Value)
		}
	case gosnmp.NoSuchObject, gosnmp.NoSuchInstance, gosnmp.EndOfMibView:
		return nil, errors.New("no such object")
	case gosnmp.Null:
		return nil, errors.New("null pdu")
	default:
		return nil, fmt.Errorf("unsupported snmp pdu type: %v", pdu.Type)
	}
}

func castToDataType(value any, dataType string) (any, error) {
	normalizedType := strings.ToLower(strings.TrimSpace(dataType))

	switch normalizedType {
	case "bool":
		booleanValue, conversionError := toBool(value)
		if conversionError != nil {
			return nil, conversionError
		}
		return booleanValue, nil
	case "int8":
		intValue, conversionError := toInt64(value)
		if conversionError != nil {
			return nil, conversionError
		}
		if intValue < math.MinInt8 || intValue > math.MaxInt8 {
			return nil, fmt.Errorf("value %d is out of range for int8", intValue)
		}
		return int64(int8(intValue)), nil
	case "int16":
		intValue, conversionError := toInt64(value)
		if conversionError != nil {
			return nil, conversionError
		}
		if intValue < math.MinInt16 || intValue > math.MaxInt16 {
			return nil, fmt.Errorf("value %d is out of range for int16", intValue)
		}
		return int64(int16(intValue)), nil
	case "int32":
		intValue, conversionError := toInt64(value)
		if conversionError != nil {
			return nil, conversionError
		}
		if intValue < math.MinInt32 || intValue > math.MaxInt32 {
			return nil, fmt.Errorf("value %d is out of range for int32", intValue)
		}
		return int64(int32(intValue)), nil
	case "uint8":
		uintValue, conversionError := toUint64(value)
		if conversionError != nil {
			return nil, conversionError
		}
		if uintValue > math.MaxUint8 {
			return nil, fmt.Errorf("value %d is out of range for uint8", uintValue)
		}
		return uint64(uint8(uintValue)), nil
	case "uint16":
		uintValue, conversionError := toUint64(value)
		if conversionError != nil {
			return nil, conversionError
		}
		if uintValue > math.MaxUint16 {
			return nil, fmt.Errorf("value %d is out of range for uint16", uintValue)
		}
		return uint64(uint16(uintValue)), nil
	case "uint32":
		uintValue, conversionError := toUint64(value)
		if conversionError != nil {
			return nil, conversionError
		}
		if uintValue > math.MaxUint32 {
			return nil, fmt.Errorf("value %d is out of range for uint32", uintValue)
		}
		return uint64(uint32(uintValue)), nil
	case "float32", "float64":
		floatValue, conversionError := toFloat64(value)
		if conversionError != nil {
			return nil, conversionError
		}
		return floatValue, nil
	case "string":
		return fmt.Sprintf("%v", value), nil
	default:
		return nil, fmt.Errorf("unsupported data_type: %s", dataType)
	}
}

func parseStringToNumber(value string) (any, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return nil, errors.New("empty string")
	}

	if intValue, parseIntError := strconv.ParseInt(trimmedValue, 10, 64); parseIntError == nil {
		return intValue, nil
	}

	if uintValue, parseUintError := strconv.ParseUint(trimmedValue, 10, 64); parseUintError == nil {
		return uintValue, nil
	}

	floatValue, parseFloatError := strconv.ParseFloat(trimmedValue, 64)
	if parseFloatError != nil {
		return nil, parseFloatError
	}

	return floatValue, nil
}

func toBool(value any) (bool, error) {
	switch convertedValue := value.(type) {
	case bool:
		return convertedValue, nil
	case int:
		return convertedValue != 0, nil
	case int8:
		return convertedValue != 0, nil
	case int16:
		return convertedValue != 0, nil
	case int32:
		return convertedValue != 0, nil
	case int64:
		return convertedValue != 0, nil
	case uint:
		return convertedValue != 0, nil
	case uint8:
		return convertedValue != 0, nil
	case uint16:
		return convertedValue != 0, nil
	case uint32:
		return convertedValue != 0, nil
	case uint64:
		return convertedValue != 0, nil
	case float32:
		return convertedValue != 0, nil
	case float64:
		return convertedValue != 0, nil
	case string:
		trimmed := strings.TrimSpace(convertedValue)
		if strings.EqualFold(trimmed, "true") {
			return true, nil
		}
		if strings.EqualFold(trimmed, "false") {
			return false, nil
		}
		if number, parseError := parseStringToNumber(trimmed); parseError == nil {
			return toBool(number)
		}
		return false, fmt.Errorf("cannot cast string %q to bool", convertedValue)
	default:
		return false, fmt.Errorf("cannot cast %T to bool", value)
	}
}

func toInt64(value any) (int64, error) {
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
	default:
		return 0, fmt.Errorf("cannot cast %T to int64", value)
	}
}

func toUint64(value any) (uint64, error) {
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
	default:
		return 0, fmt.Errorf("cannot cast %T to uint64", value)
	}
}

func toFloat64(value any) (float64, error) {
	switch convertedValue := value.(type) {
	case int:
		return float64(convertedValue), nil
	case int8:
		return float64(convertedValue), nil
	case int16:
		return float64(convertedValue), nil
	case int32:
		return float64(convertedValue), nil
	case int64:
		return float64(convertedValue), nil
	case uint:
		return float64(convertedValue), nil
	case uint8:
		return float64(convertedValue), nil
	case uint16:
		return float64(convertedValue), nil
	case uint32:
		return float64(convertedValue), nil
	case uint64:
		return float64(convertedValue), nil
	case float32:
		return float64(convertedValue), nil
	case float64:
		return convertedValue, nil
	default:
		return 0, fmt.Errorf("cannot cast %T to float64", value)
	}
}
