package domain

import (
	"time"

	"github.com/google/uuid"
)

// Label represents a tag that can be applied to items for categorization and filtering.
type Label struct {
	ID        uuid.UUID
	Name      string
	Color     *string // Optional hex color code for UI display
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ItemLabel represents the junction between an item and a label.
type ItemLabel struct {
	ID        uuid.UUID
	ItemID    uuid.UUID
	LabelID   uuid.UUID
	CreatedAt time.Time
}
