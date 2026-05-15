package domain

import (
	"time"

	"github.com/google/uuid"
)

// PackingList represents a global packing list template
type PackingList struct {
	ID          uuid.UUID
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PackingListLabel represents the junction between a packing list and a label
type PackingListLabel struct {
	ID            uuid.UUID
	PackingListID uuid.UUID
	LabelID       uuid.UUID
	CreatedAt     time.Time
}
