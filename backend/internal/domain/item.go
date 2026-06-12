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
	ImageBlob          []byte
	ImageMimeType      *string
	ImageSizeBytes     *int
	ImageWidthPX       *int
	ImageHeightPX      *int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
