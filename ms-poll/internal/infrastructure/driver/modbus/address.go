package modbus

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RegisterType string

const (
	RegHolding  RegisterType = "holding"
	RegInput    RegisterType = "input"
	RegCoil     RegisterType = "coil"
	RegDiscrete RegisterType = "discrete"
)

type ParsedAddress struct {
	RegisterType RegisterType
	Offset       uint16
	Length       uint16
}

type rawTagAddress struct {
	RegisterType string `json:"register_type"`
	Address      int    `json:"address"`
	Length       *int   `json:"length"`
}

func ParseTagAddress(
	raw json.RawMessage,
	dataType string,
) (ParsedAddress, error) {
	if len(raw) == 0 {
		return ParsedAddress{}, fmt.Errorf("tag address is empty")
	}

	var parsedRawAddress rawTagAddress
	if unmarshalError := json.Unmarshal(raw, &parsedRawAddress); unmarshalError != nil {
		return ParsedAddress{}, fmt.Errorf("unmarshal tag address: %w", unmarshalError)
	}

	registerType := RegisterType(strings.ToLower(strings.TrimSpace(
		parsedRawAddress.RegisterType,
	)))
	if !isKnownRegisterType(registerType) {
		return ParsedAddress{}, fmt.Errorf(
			"unsupported register_type: %s",
			parsedRawAddress.RegisterType,
		)
	}

	length, resolveLengthError := resolveAddressLength(
		registerType,
		strings.ToLower(strings.TrimSpace(dataType)),
		parsedRawAddress.Length,
	)
	if resolveLengthError != nil {
		return ParsedAddress{}, resolveLengthError
	}

	rangeStart, rangeEnd := modbusAddressRange(registerType)
	if parsedRawAddress.Address < rangeStart || parsedRawAddress.Address > rangeEnd {
		return ParsedAddress{}, fmt.Errorf(
			"address %d is out of range for register_type %s",
			parsedRawAddress.Address,
			registerType,
		)
	}

	offset := parsedRawAddress.Address - rangeStart
	rangeCapacity := (rangeEnd - rangeStart) + 1
	if offset+length > rangeCapacity {
		return ParsedAddress{}, fmt.Errorf(
			"address %d length %d is out of range for register_type %s",
			parsedRawAddress.Address,
			length,
			registerType,
		)
	}

	return ParsedAddress{
		RegisterType: registerType,
		Offset:       uint16(offset),
		Length:       uint16(length),
	}, nil
}

func isKnownRegisterType(registerType RegisterType) bool {
	switch registerType {
	case RegHolding, RegInput, RegCoil, RegDiscrete:
		return true
	default:
		return false
	}
}

func modbusAddressRange(registerType RegisterType) (int, int) {
	switch registerType {
	case RegHolding:
		return 40001, 49999
	case RegInput:
		return 30001, 39999
	case RegCoil:
		return 1, 9999
	case RegDiscrete:
		return 10001, 19999
	default:
		return 0, 0
	}
}

func resolveAddressLength(
	registerType RegisterType,
	dataType string,
	rawLength *int,
) (int, error) {
	switch registerType {
	case RegCoil, RegDiscrete:
		if dataType != "bool" {
			return 0, fmt.Errorf(
				"data_type %s is not supported for register_type %s",
				dataType,
				registerType,
			)
		}
		if rawLength != nil && *rawLength != 1 {
			return 0, fmt.Errorf(
				"length must be 1 for register_type %s",
				registerType,
			)
		}
		return 1, nil
	case RegHolding, RegInput:
		switch dataType {
		case "bool", "int8", "uint8", "int16", "uint16":
			return resolveFixedLength(rawLength, 1, dataType)
		case "int32", "uint32", "float32":
			return resolveFixedLength(rawLength, 2, dataType)
		case "float64":
			return resolveFixedLength(rawLength, 4, dataType)
		case "string":
			if rawLength == nil {
				return 0, fmt.Errorf("length is required for data_type string")
			}
			if *rawLength <= 0 {
				return 0, fmt.Errorf(
					"length must be positive for data_type string",
				)
			}
			return *rawLength, nil
		default:
			return 0, fmt.Errorf("unsupported data_type: %s", dataType)
		}
	default:
		return 0, fmt.Errorf("unsupported register_type: %s", registerType)
	}
}

func resolveFixedLength(
	rawLength *int,
	expectedLength int,
	dataType string,
) (int, error) {
	if rawLength == nil {
		return expectedLength, nil
	}

	if *rawLength != expectedLength {
		return 0, fmt.Errorf(
			"length %d does not match data_type %s (expected %d)",
			*rawLength,
			dataType,
			expectedLength,
		)
	}

	return expectedLength, nil
}
