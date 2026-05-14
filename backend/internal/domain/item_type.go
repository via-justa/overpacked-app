package domain

import (
	"time"

	"github.com/google/uuid"
)

type ItemType struct {
	ID          string
	Name        string
	Description *string
	BaseProfile *string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ItemTypeField struct {
	ID          uuid.UUID
	ItemTypeID  string
	FieldKey    string
	FieldLabel  string
	DataType    string
	IsRequired  bool
	EnumOptions []string
	MinValue    *float64
	MaxValue    *float64
	Unit        *string
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
