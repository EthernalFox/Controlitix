package domain

import "time"

type Diagram struct {
	ID          string
	ObjectID    string
	Name        *string
	Description *string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
