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

// TripPerson represents a person participating in a trip
type TripPerson struct {
	ID        uuid.UUID
	TripID    uuid.UUID
	PersonID  uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TripPersonPack represents a pack assigned to a person in a trip
type TripPersonPack struct {
	ID           uuid.UUID
	TripPersonID uuid.UUID
	PackID       uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TripPersonItem represents an item assigned to a person in a trip (worn/not in pack)
type TripPersonItem struct {
	ID           uuid.UUID
	TripPersonID uuid.UUID
	ItemID       uuid.UUID
	Quantity     int
	CarryStatus  CarryStatus
	Notes        *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
