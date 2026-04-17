package domain

import (
	"errors"
	"strings"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrConflict       = errors.New("conflict")
	ErrNotImplemented = errors.New("not implemented")
)

type FieldError struct {
	Field   string
	Message string
}

type ValidationError struct {
	Fields []FieldError
}

func NewValidationError(fields ...FieldError) *ValidationError {
	validationFields := make([]FieldError, 0, len(fields))
	for _, field := range fields {
		if strings.TrimSpace(field.Field) == "" && strings.TrimSpace(field.Message) == "" {
			continue
		}

		validationFields = append(validationFields, field)
	}

	return &ValidationError{Fields: validationFields}
}

func (validationError *ValidationError) Error() string {
	if validationError == nil || len(validationError.Fields) == 0 {
		return "validation failed"
	}

	fieldMessages := make([]string, 0, len(validationError.Fields))
	for _, field := range validationError.Fields {
		if strings.TrimSpace(field.Field) == "" {
			fieldMessages = append(fieldMessages, field.Message)
			continue
		}

		fieldMessages = append(fieldMessages, field.Field+": "+field.Message)
	}

	return "validation failed: " + strings.Join(fieldMessages, ", ")
}
