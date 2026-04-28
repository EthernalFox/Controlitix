package domain

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrNotPublished = errors.New("not published")
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("forbidden")
	ErrUnavailable  = errors.New("unavailable")
)
