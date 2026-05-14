package domain

import (
	"time"

	"github.com/google/uuid"
)

type Manufacturer struct {
	ID        uuid.UUID
	Name      string
	Website   *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
