package domain

import (
	"time"

	"github.com/google/uuid"
)

type TripType string

const (
	TripTypeDayHike   TripType = "day_hike"
	TripTypeOvernight TripType = "overnight"
	TripTypeMultiDay  TripType = "multi_day"
	TripTypeThruHike  TripType = "thru_hike"
)

type Pack struct {
	ID         uuid.UUID
	PersonID   *uuid.UUID
	Name       string
	TripType   *TripType
	Notes      *string
	IsTemplate bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PackItem struct {
	ID          uuid.UUID
	PackID      uuid.UUID
	ItemID      uuid.UUID
	Quantity    int
	CarryStatus CarryStatus
	Notes       *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PackSet struct {
	ID        uuid.UUID
	PackID    uuid.UUID
	SetID     uuid.UUID
	AppliedAt time.Time
	CreatedAt time.Time
}
