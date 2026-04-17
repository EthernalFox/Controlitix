package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"regexp"
	"strings"
	"unicode/utf8"
)

const maxNameLength = 255

var oidPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)+$`)

func ValidateRequiredName(field string, value string) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return fmt.Errorf("%s is required: %w", field, ErrInvalidInput)
	}

	if utf8.RuneCountInString(trimmedValue) > maxNameLength {
		return NewValidationError(FieldError{
			Field:   field,
			Message: "must not exceed 255 characters",
		})
	}

	return nil
}

func ValidateOptionalName(field string, value *string) error {
	if value == nil {
		return nil
	}

	trimmedValue := strings.TrimSpace(*value)
	if trimmedValue == "" {
		return fmt.Errorf("%s is required: %w", field, ErrInvalidInput)
	}

	if utf8.RuneCountInString(trimmedValue) > maxNameLength {
		return NewValidationError(FieldError{
			Field:   field,
			Message: "must not exceed 255 characters",
		})
	}

	return nil
}

func ValidateOptionalMaxLength(field string, value *string, maxLength int) error {
	if value == nil {
		return nil
	}

	if utf8.RuneCountInString(*value) > maxLength {
		return NewValidationError(FieldError{
			Field:   field,
			Message: fmt.Sprintf("must not exceed %d characters", maxLength),
		})
	}

	return nil
}

func ValidateTagAddress(deviceTypeName string, address json.RawMessage) *ValidationError {
	addressObject, objectValidationError := decodeJSONObject(address)
	if objectValidationError != nil {
		return objectValidationError
	}

	validationFields := make([]FieldError, 0)

	switch deviceTypeName {
	case "modbus_rtu", "modbus_tcp":
		registerType := requiredStringField(addressObject, "register_type", "address.", &validationFields)
		if registerType != "" && !isOneOf(registerType, "holding", "input", "coil", "discrete") {
			validationFields = append(validationFields, FieldError{
				Field:   "address.register_type",
				Message: `must be one of: "holding", "input", "coil", "discrete"`,
			})
		}

		registerAddress, ok := requiredIntField(addressObject, "address", "address.", &validationFields)
		if ok && (registerAddress < 0 || registerAddress > 65535) {
			validationFields = append(validationFields, FieldError{
				Field:   "address.address",
				Message: "must be between 0 and 65535",
			})
		}
	case "snmp_v1", "snmp_v2c", "snmp_v3":
		oid := requiredStringField(addressObject, "oid", "address.", &validationFields)
		if oid != "" && !oidPattern.MatchString(oid) {
			validationFields = append(validationFields, FieldError{
				Field:   "address.oid",
				Message: `must match pattern ^[0-9]+(\.[0-9]+)+$`,
			})
		}
	default:
		validationFields = append(validationFields, FieldError{
			Field:   "address",
			Message: "unsupported device type for address validation",
		})
	}

	if len(validationFields) == 0 {
		return nil
	}

	return NewValidationError(validationFields...)
}

func ValidateSetpoints(setpoints TagSetpoints) *ValidationError {
	validationFields := make([]FieldError, 0)

	validateFiniteFloat("lolo", setpoints.LoLo, &validationFields)
	validateFiniteFloat("lo", setpoints.Lo, &validationFields)
	validateFiniteFloat("hi", setpoints.Hi, &validationFields)
	validateFiniteFloat("hihi", setpoints.HiHi, &validationFields)

	validateOrderedPair("lolo", setpoints.LoLo, "lo", setpoints.Lo, &validationFields)
	validateOrderedPair("lolo", setpoints.LoLo, "hi", setpoints.Hi, &validationFields)
	validateOrderedPair("lolo", setpoints.LoLo, "hihi", setpoints.HiHi, &validationFields)
	validateOrderedPair("lo", setpoints.Lo, "hi", setpoints.Hi, &validationFields)
	validateOrderedPair("lo", setpoints.Lo, "hihi", setpoints.HiHi, &validationFields)
	validateOrderedPair("hi", setpoints.Hi, "hihi", setpoints.HiHi, &validationFields)

	if len(validationFields) == 0 {
		return nil
	}

	return NewValidationError(validationFields...)
}

func ValidateSetpointsValue(setpoints *TagSetpoints) *ValidationError {
	if setpoints == nil {
		return nil
	}

	return ValidateSetpoints(*setpoints)
}

func ValidateScaling(scaling TagScaling) *ValidationError {
	validationFields := make([]FieldError, 0)

	validateFiniteFloat("raw_min", scaling.RawMin, &validationFields)
	validateFiniteFloat("raw_max", scaling.RawMax, &validationFields)
	validateFiniteFloat("eng_min", scaling.EngMin, &validationFields)
	validateFiniteFloat("eng_max", scaling.EngMax, &validationFields)
	validateFiniteFloat("factor", scaling.Factor, &validationFields)
	validateFiniteFloat("offset", scaling.Offset, &validationFields)

	interpolationValues := []*float64{
		scaling.RawMin,
		scaling.RawMax,
		scaling.EngMin,
		scaling.EngMax,
	}
	interpolationFields := []string{"raw_min", "raw_max", "eng_min", "eng_max"}
	interpolationCount := 0
	for _, value := range interpolationValues {
		if value != nil {
			interpolationCount++
		}
	}

	if interpolationCount > 0 && interpolationCount < len(interpolationValues) {
		for index, value := range interpolationValues {
			if value == nil {
				validationFields = append(validationFields, FieldError{
					Field:   interpolationFields[index],
					Message: "must be provided when interpolation is used",
				})
			}
		}
	}

	if scaling.RawMin != nil && scaling.RawMax != nil && *scaling.RawMin == *scaling.RawMax {
		validationFields = append(validationFields, FieldError{
			Field:   "raw_max",
			Message: "must not be equal to raw_min",
		})
	}

	if scaling.Factor != nil && *scaling.Factor == 0 {
		validationFields = append(validationFields, FieldError{
			Field:   "factor",
			Message: "must not be zero",
		})
	}

	if len(validationFields) == 0 {
		return nil
	}

	return NewValidationError(validationFields...)
}

func ValidateScalingValue(scaling *TagScaling) *ValidationError {
	if scaling == nil {
		return nil
	}

	return ValidateScaling(*scaling)
}

func decodeJSONObject(payload json.RawMessage) (map[string]json.RawMessage, *ValidationError) {
	trimmedPayload := strings.TrimSpace(string(payload))
	if trimmedPayload == "" || trimmedPayload == "null" {
		trimmedPayload = "{}"
	}

	var object map[string]json.RawMessage
	if unmarshalError := json.Unmarshal([]byte(trimmedPayload), &object); unmarshalError != nil {
		return nil, NewValidationError(FieldError{
			Field:   "",
			Message: "must be a valid JSON object",
		})
	}

	if object == nil {
		object = make(map[string]json.RawMessage)
	}

	return object, nil
}

func requiredStringField(
	object map[string]json.RawMessage,
	field string,
	prefix string,
	validationFields *[]FieldError,
) string {
	rawValue, exists := object[field]
	fullFieldName := prefix + field
	if !exists {
		*validationFields = append(*validationFields, FieldError{
			Field:   fullFieldName,
			Message: "is required",
		})
		return ""
	}

	var stringValue string
	if unmarshalError := json.Unmarshal(rawValue, &stringValue); unmarshalError != nil {
		*validationFields = append(*validationFields, FieldError{
			Field:   fullFieldName,
			Message: "must be a string",
		})
		return ""
	}

	if strings.TrimSpace(stringValue) == "" {
		*validationFields = append(*validationFields, FieldError{
			Field:   fullFieldName,
			Message: "must not be empty",
		})
		return ""
	}

	return stringValue
}

func requiredIntField(
	object map[string]json.RawMessage,
	field string,
	prefix string,
	validationFields *[]FieldError,
) (int, bool) {
	rawValue, exists := object[field]
	fullFieldName := prefix + field
	if !exists {
		*validationFields = append(*validationFields, FieldError{
			Field:   fullFieldName,
			Message: "is required",
		})
		return 0, false
	}

	var intValue int
	if unmarshalError := json.Unmarshal(rawValue, &intValue); unmarshalError != nil {
		*validationFields = append(*validationFields, FieldError{
			Field:   fullFieldName,
			Message: "must be an integer",
		})
		return 0, false
	}

	return intValue, true
}

func optionalStringField(
	object map[string]json.RawMessage,
	field string,
	prefix string,
	validationFields *[]FieldError,
) (string, bool) {
	rawValue, exists := object[field]
	if !exists {
		return "", false
	}

	var stringValue string
	if unmarshalError := json.Unmarshal(rawValue, &stringValue); unmarshalError != nil {
		*validationFields = append(*validationFields, FieldError{
			Field:   prefix + field,
			Message: "must be a string",
		})
		return "", false
	}

	if strings.TrimSpace(stringValue) == "" {
		*validationFields = append(*validationFields, FieldError{
			Field:   prefix + field,
			Message: "must not be empty",
		})
		return "", false
	}

	return stringValue, true
}

func optionalIntField(
	object map[string]json.RawMessage,
	field string,
	prefix string,
	validationFields *[]FieldError,
) (int, bool) {
	rawValue, exists := object[field]
	if !exists {
		return 0, false
	}

	var intValue int
	if unmarshalError := json.Unmarshal(rawValue, &intValue); unmarshalError != nil {
		*validationFields = append(*validationFields, FieldError{
			Field:   prefix + field,
			Message: "must be an integer",
		})
		return 0, false
	}

	return intValue, true
}

func isOneOf(value string, allowedValues ...string) bool {
	for _, allowedValue := range allowedValues {
		if value == allowedValue {
			return true
		}
	}

	return false
}

func isValidHost(value string) bool {
	if net.ParseIP(value) != nil {
		return true
	}

	if len(value) > 253 {
		return false
	}

	labels := strings.Split(value, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 {
			return false
		}

		for index, character := range label {
			isLetter := character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z'
			isDigit := character >= '0' && character <= '9'
			isHyphen := character == '-'
			if !isLetter && !isDigit && !isHyphen {
				return false
			}

			if isHyphen && (index == 0 || index == len(label)-1) {
				return false
			}
		}
	}

	return true
}

func validateFiniteFloat(field string, value *float64, validationFields *[]FieldError) {
	if value == nil {
		return
	}

	if math.IsNaN(*value) || math.IsInf(*value, 0) {
		*validationFields = append(*validationFields, FieldError{
			Field:   field,
			Message: "must be a finite number",
		})
	}
}

func validateOrderedPair(
	leftField string,
	leftValue *float64,
	rightField string,
	rightValue *float64,
	validationFields *[]FieldError,
) {
	if leftValue == nil || rightValue == nil {
		return
	}

	if *leftValue >= *rightValue {
		*validationFields = append(*validationFields, FieldError{
			Field:   leftField,
			Message: fmt.Sprintf("must be less than %s", rightField),
		})
	}
}
