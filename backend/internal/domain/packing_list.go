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
