package domain

import (
	"time"

	"github.com/google/uuid"
)

type ItemSet struct {
	ID          uuid.UUID
	Name        string
	SetCategory string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SetItem struct {
	ID        uuid.UUID
	SetID     uuid.UUID
	ItemID    uuid.UUID
	Quantity  int
	Notes     *string
	SortOrder int
}
