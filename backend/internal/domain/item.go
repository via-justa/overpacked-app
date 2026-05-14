package domain

import (
	"time"

	"github.com/google/uuid"
)

type CarryStatus string

const (
	CarryStatusPacked CarryStatus = "packed"
	CarryStatusWorn   CarryStatus = "worn"
)

type Item struct {
	ID                 uuid.UUID
	ManufacturerID     uuid.UUID
	TypeID             string
	Name               string
	IsActive           bool
	Attributes         map[string]any
	Description        *string
	SourceURL          *string
	Price              *float64
	WeightGrams        *float64
	VolumeML           *float64
	DefaultQuantity    int
	DefaultCarryStatus CarryStatus
	IsDefault          bool
	DoseCount          *int
	Calories           *float64
	CaloriesPerServing *float64
	RequiresWater      *bool
	Season             *string
	Layer              *string
	Waterproof         *bool
	Size               *string
	Color              *string
	CapacityPeople     *float64
	SeasonRating       *string
	Freestanding       *bool
	HasFootprint       *bool
	ComfortTempC       *float64
	FillType           *string
	RValue             *float64
	CapacityMAH        *int
	ChargePort         *string
	Rechargeable       *bool
	ImageBlob          []byte
	ImageMimeType      *string
	ImageSizeBytes     *int
	ImageWidthPX       *int
	ImageHeightPX      *int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
