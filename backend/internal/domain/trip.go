package domain

import (
	"time"

	"github.com/google/uuid"
)

// Trip represents a hiking trip with associated packs, items, sets, and persons
type Trip struct {
	ID              uuid.UUID
	Name            string
	TripType        TripType
	Duration        *string // stored as PostgreSQL INTERVAL, represented as string
	Notes           *string
	TripKomootURL   *string
	TripStravaURL   *string
	TripWandererURL *string
	TotalDistanceKm *float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TripPack represents a pack assigned to a trip
type TripPack struct {
	ID     uuid.UUID
	TripID uuid.UUID
	PackID uuid.UUID
}

// TripItem represents an item assigned to a trip with quantity and carry status
type TripItem struct {
	ID          uuid.UUID
	TripID      uuid.UUID
	ItemID      uuid.UUID
	Quantity    int
	CarryStatus CarryStatus
	Notes       *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TripSet represents a set assigned to a trip
type TripSet struct {
	ID        uuid.UUID
	TripID    uuid.UUID
	SetID     uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TripPerson represents a person participating in a trip
type TripPerson struct {
	ID       uuid.UUID
	TripID   uuid.UUID
	PersonID uuid.UUID
}
