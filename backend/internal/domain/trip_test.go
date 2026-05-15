package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertTripBasicFields(t *testing.T, trip Trip, expectedID uuid.UUID, expectedName string, expectedType TripType) {
	t.Helper()
	if trip.ID != expectedID {
		t.Errorf("expected trip ID %v, got %v", expectedID, trip.ID)
	}
	if trip.Name != expectedName {
		t.Errorf("expected name %s, got %s", expectedName, trip.Name)
	}
	if trip.TripType != expectedType {
		t.Errorf("expected trip type %s, got %s", expectedType, trip.TripType)
	}
}

func assertTripOptionalFields(t *testing.T, trip Trip, expectedDuration, expectedNotes string, expectedDistance float64) {
	t.Helper()
	if trip.Duration == nil || *trip.Duration != expectedDuration {
		t.Errorf("expected duration %s, got %v", expectedDuration, trip.Duration)
	}
	if trip.TotalDistanceKm == nil || *trip.TotalDistanceKm != expectedDistance {
		t.Errorf("expected distance %f, got %v", expectedDistance, trip.TotalDistanceKm)
	}
	if trip.Notes == nil || *trip.Notes != expectedNotes {
		t.Errorf("expected notes %s, got %v", expectedNotes, trip.Notes)
	}
}

func assertTripURLs(t *testing.T, trip Trip, komootURL, stravaURL, wandererURL string) {
	t.Helper()
	if trip.TripKomootURL == nil || *trip.TripKomootURL != komootURL {
		t.Errorf("expected komoot URL %s, got %v", komootURL, trip.TripKomootURL)
	}
	if trip.TripStravaURL == nil || *trip.TripStravaURL != stravaURL {
		t.Errorf("expected strava URL %s, got %v", stravaURL, trip.TripStravaURL)
	}
	if trip.TripWandererURL == nil || *trip.TripWandererURL != wandererURL {
		t.Errorf("expected wanderer URL %s, got %v", wandererURL, trip.TripWandererURL)
	}
}

func assertTripTimestamps(t *testing.T, trip Trip) {
	t.Helper()
	if trip.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if trip.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestTripCreation(t *testing.T) {
	t.Parallel()

	tripID := uuid.New()
	name := "Summer Backpacking Trip"
	tripType := TripTypeOvernight
	duration := "2 days"
	notes := "Testing trip notes"
	komootURL := "https://komoot.com/trip/123"
	stravaURL := "https://strava.com/activities/456"
	wandererURL := "https://wanderer.travel/trip/789"
	distanceKm := 25.5

	trip := Trip{
		ID:              tripID,
		Name:            name,
		TripType:        tripType,
		Duration:        &duration,
		Notes:           &notes,
		TripKomootURL:   &komootURL,
		TripStravaURL:   &stravaURL,
		TripWandererURL: &wandererURL,
		TotalDistanceKm: &distanceKm,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	assertTripBasicFields(t, trip, tripID, name, tripType)
	assertTripOptionalFields(t, trip, duration, notes, distanceKm)
	assertTripURLs(t, trip, komootURL, stravaURL, wandererURL)
	assertTripTimestamps(t, trip)
}

func TestTripTypeConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tripType TripType
		expected string
	}{
		{"day hike", TripTypeDayHike, "day_hike"},
		{"overnight", TripTypeOvernight, "overnight"},
		{"multi-day", TripTypeMultiDay, "multi_day"},
		{"thru-hike", TripTypeThruHike, "thru_hike"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.tripType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.tripType))
			}
		})
	}
}

func TestTripPackCreation(t *testing.T) {
	t.Parallel()

	tripPack := TripPack{
		ID:     uuid.New(),
		TripID: uuid.New(),
		PackID: uuid.New(),
	}

	if tripPack.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if tripPack.TripID == uuid.Nil {
		t.Error("expected non-nil TripID")
	}
	if tripPack.PackID == uuid.Nil {
		t.Error("expected non-nil PackID")
	}
}

func TestTripItemCreation(t *testing.T) {
	t.Parallel()

	notes := "Bring extra batteries"
	tripItem := TripItem{
		ID:          uuid.New(),
		TripID:      uuid.New(),
		ItemID:      uuid.New(),
		Quantity:    2,
		CarryStatus: CarryStatusPacked,
		Notes:       &notes,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if tripItem.Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", tripItem.Quantity)
	}
	if tripItem.CarryStatus != CarryStatusPacked {
		t.Errorf("expected carry status %s, got %s", CarryStatusPacked, tripItem.CarryStatus)
	}
	if tripItem.Notes == nil || *tripItem.Notes != notes {
		t.Errorf("expected notes %s, got %v", notes, tripItem.Notes)
	}
	if tripItem.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if tripItem.TripID == uuid.Nil {
		t.Error("expected non-nil TripID")
	}
	if tripItem.ItemID == uuid.Nil {
		t.Error("expected non-nil ItemID")
	}
	if tripItem.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if tripItem.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestTripSetCreation(t *testing.T) {
	t.Parallel()

	tripSet := TripSet{
		ID:        uuid.New(),
		TripID:    uuid.New(),
		SetID:     uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if tripSet.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if tripSet.TripID == uuid.Nil {
		t.Error("expected non-nil TripID")
	}
	if tripSet.SetID == uuid.Nil {
		t.Error("expected non-nil SetID")
	}
	if tripSet.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if tripSet.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestTripPersonCreation(t *testing.T) {
	t.Parallel()

	tripPerson := TripPerson{
		ID:       uuid.New(),
		TripID:   uuid.New(),
		PersonID: uuid.New(),
	}

	if tripPerson.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if tripPerson.TripID == uuid.Nil {
		t.Error("expected non-nil TripID")
	}
	if tripPerson.PersonID == uuid.Nil {
		t.Error("expected non-nil PersonID")
	}
}
