package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

func TestLabelCreation(t *testing.T) {
	color := "#FF5733"
	label := &domain.Label{
		ID:        uuid.New(),
		Name:      "Ultralight",
		Color:     &color,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if label.ID == uuid.Nil {
		t.Error("Label ID should not be nil")
	}

	if label.Name != "Ultralight" {
		t.Errorf("Expected name 'Ultralight', got '%s'", label.Name)
	}

	if label.Color == nil || *label.Color != "#FF5733" {
		t.Error("Label color should be '#FF5733'")
	}

	if label.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	if label.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestLabelWithoutColor(t *testing.T) {
	label := &domain.Label{
		Color: nil,
	}

	if label.Color != nil {
		t.Error("Label color should be nil")
	}
}

func TestItemLabelCreation(t *testing.T) {
	itemLabel := &domain.ItemLabel{
		ID:        uuid.New(),
		ItemID:    uuid.New(),
		LabelID:   uuid.New(),
		CreatedAt: time.Now(),
	}

	if itemLabel.ID == uuid.Nil {
		t.Error("ItemLabel ID should not be nil")
	}

	if itemLabel.ItemID == uuid.Nil {
		t.Error("ItemLabel ItemID should not be nil")
	}

	if itemLabel.LabelID == uuid.Nil {
		t.Error("ItemLabel LabelID should not be nil")
	}

	if itemLabel.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}
