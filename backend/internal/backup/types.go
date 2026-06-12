// Package backup builds and restores full-data archives (deliverable A) and
// produces generic items-only CSV exports (deliverable B) for the overpacked-app.
//
// A backup archive is a ZIP containing:
//   - manifest.json: format version, creation time, and per-table counts
//   - data.json:     a Snapshot of all user data (see Snapshot)
//   - images/<id>.<ext>: raw item image bytes, referenced from data.json by filename
//
// Manufacturers and labels are carried by stable id together with their natural
// keys (name), so a restore resolves them against whatever catalog the target
// instance has (created by the seed) without dumping the full catalog.
package backup

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// FormatVersion is the on-disk schema version of the archive. Bump it when the
// Snapshot shape changes incompatibly; Import refuses newer versions.
const FormatVersion = 1

const (
	manifestFilename = "manifest.json"
	dataFilename     = "data.json"
	imagesDir        = "images"
)

// Mode controls how Import reconciles archive data with existing data.
type Mode string

const (
	// ModeReplace wipes user content (preserving the manufacturers/labels catalog
	// and system item types) before inserting the archive.
	ModeReplace Mode = "replace"
	// ModeMerge upserts the archive on top of existing data.
	ModeMerge Mode = "merge"
)

func (m Mode) valid() bool { return m == ModeReplace || m == ModeMerge }

// Manifest is the archive header.
type Manifest struct {
	FormatVersion int            `json:"format_version"`
	CreatedAt     time.Time      `json:"created_at"`
	Counts        map[string]int `json:"counts"`
}

// Snapshot is the complete set of user data carried by an archive.
type Snapshot struct {
	Settings          *settingsDTO          `json:"settings,omitempty"`
	Manufacturers     []manufacturerDTO     `json:"manufacturers"`
	Labels            []labelDTO            `json:"labels"`
	ItemTypes         []itemTypeDTO         `json:"item_types"`
	ItemTypeFields    []itemTypeFieldDTO    `json:"item_type_fields"`
	Persons           []personDTO           `json:"persons"`
	Items             []itemDTO             `json:"items"`
	ItemLabels        []itemLabelDTO        `json:"item_labels"`
	Sets              []setDTO              `json:"item_sets"`
	SetItems          []setItemDTO          `json:"set_items"`
	Packs             []packDTO             `json:"packs"`
	PackItems         []packItemDTO         `json:"pack_items"`
	PackingLists      []packingListDTO      `json:"packing_lists"`
	PackingListLabels []packingListLabelDTO `json:"packing_list_labels"`
	Trips             []tripDTO             `json:"trips"`
	TripPersons       []tripPersonDTO       `json:"trip_persons"`
	TripPersonPacks   []tripPersonPackDTO   `json:"trip_person_packs"`
	TripPersonItems   []tripPersonItemDTO   `json:"trip_person_items"`
}

type settingsDTO struct {
	WeightUnit      string `json:"weight_unit"`
	DistanceUnit    string `json:"distance_unit"`
	TemperatureUnit string `json:"temperature_unit"`
	VolumeUnit      string `json:"volume_unit"`
	Currency        string `json:"currency"`
}

type manufacturerDTO struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Website *string   `json:"website,omitempty"`
}

type labelDTO struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color *string   `json:"color,omitempty"`
}

type itemTypeDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	BaseProfile *string `json:"base_profile,omitempty"`
}

type itemTypeFieldDTO struct {
	ID          uuid.UUID       `json:"id"`
	ItemTypeID  string          `json:"item_type_id"`
	FieldKey    string          `json:"field_key"`
	FieldLabel  string          `json:"field_label"`
	DataType    string          `json:"data_type"`
	IsRequired  bool            `json:"is_required"`
	EnumOptions json.RawMessage `json:"enum_options_json,omitempty"`
	MinValue    *float64        `json:"min_value,omitempty"`
	MaxValue    *float64        `json:"max_value,omitempty"`
	Unit        *string         `json:"unit,omitempty"`
	SortOrder   int             `json:"sort_order"`
}

type personDTO struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Gender            *string    `json:"gender,omitempty"`
	Birthdate         *time.Time `json:"birthdate,omitempty"`
	BodyWeightGrams   *float64   `json:"body_weight_grams,omitempty"`
	ConditioningLevel *string    `json:"conditioning_level,omitempty"`
}

type itemDTO struct {
	ID                 uuid.UUID       `json:"id"`
	ManufacturerID     uuid.UUID       `json:"manufacturer_id"`
	TypeID             string          `json:"type_id"`
	Name               string          `json:"name"`
	IsActive           bool            `json:"is_active"`
	Description        *string         `json:"description,omitempty"`
	SourceURL          *string         `json:"source_url,omitempty"`
	Price              *float64        `json:"price,omitempty"`
	WeightGrams        *float64        `json:"weight_grams,omitempty"`
	VolumeML           *float64        `json:"volume_ml,omitempty"`
	DefaultQuantity    int             `json:"default_quantity"`
	DefaultCarryStatus string          `json:"default_carry_status"`
	IsDefault          bool            `json:"is_default"`
	ImageFile          *string         `json:"image_file,omitempty"`
	ImageMimeType      *string         `json:"image_mime_type,omitempty"`
	ImageSizeBytes     *int            `json:"image_size_bytes,omitempty"`
	ImageWidthPX       *int            `json:"image_width_px,omitempty"`
	ImageHeightPX      *int            `json:"image_height_px,omitempty"`
	Attributes         json.RawMessage `json:"attributes,omitempty"`
}

type itemLabelDTO struct {
	ItemID  uuid.UUID `json:"item_id"`
	LabelID uuid.UUID `json:"label_id"`
}

type setDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	SetCategory string    `json:"set_category"`
}

type setItemDTO struct {
	SetID     uuid.UUID `json:"set_id"`
	ItemID    uuid.UUID `json:"item_id"`
	Quantity  int       `json:"quantity"`
	Notes     *string   `json:"notes,omitempty"`
	SortOrder int       `json:"sort_order"`
}

type packDTO struct {
	ID         uuid.UUID  `json:"id"`
	PersonID   *uuid.UUID `json:"person_id,omitempty"`
	Name       string     `json:"name"`
	TripType   *string    `json:"trip_type,omitempty"`
	Notes      *string    `json:"notes,omitempty"`
	IsTemplate bool       `json:"is_template"`
}

type packItemDTO struct {
	PackID      uuid.UUID `json:"pack_id"`
	ItemID      uuid.UUID `json:"item_id"`
	Quantity    int       `json:"quantity"`
	CarryStatus string    `json:"carry_status"`
	Notes       *string   `json:"notes,omitempty"`
}

type packingListDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
}

type packingListLabelDTO struct {
	PackingListID uuid.UUID `json:"packing_list_id"`
	LabelID       uuid.UUID `json:"label_id"`
}

type tripDTO struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	TripType        string    `json:"trip_type"`
	Duration        *string   `json:"duration,omitempty"`
	Notes           *string   `json:"notes,omitempty"`
	TripKomootURL   *string   `json:"trip_komoot_url,omitempty"`
	TripStravaURL   *string   `json:"trip_strava_url,omitempty"`
	TripWandererURL *string   `json:"trip_wanderer_url,omitempty"`
	TotalDistanceKm *float64  `json:"total_distance_km,omitempty"`
}

type tripPersonDTO struct {
	ID       uuid.UUID `json:"id"`
	TripID   uuid.UUID `json:"trip_id"`
	PersonID uuid.UUID `json:"person_id"`
}

type tripPersonPackDTO struct {
	TripPersonID uuid.UUID `json:"trip_person_id"`
	PackID       uuid.UUID `json:"pack_id"`
}

type tripPersonItemDTO struct {
	TripPersonID uuid.UUID `json:"trip_person_id"`
	ItemID       uuid.UUID `json:"item_id"`
	Quantity     int       `json:"quantity"`
	CarryStatus  *string   `json:"carry_status,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
}

// ImportResult reports how many rows were written per table.
type ImportResult struct {
	Mode   Mode           `json:"mode"`
	Counts map[string]int `json:"counts"`
}
