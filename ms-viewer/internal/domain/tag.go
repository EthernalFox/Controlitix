package domain

import "github.com/google/uuid"

type TagSearchQuery struct {
	Search string
	Limit  int
}

type TagSearchItem struct {
	ID         uuid.UUID
	Name       string
	DeviceID   uuid.UUID
	DeviceName string
	UnitSymbol string
}

type TagSearchResult struct {
	Items []TagSearchItem
	Total int
}
