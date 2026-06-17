package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/db"
	"github.com/via-justa/overpacked-app/backend/internal/migrations"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

func TestSettingsHandlerUpdateSettingsInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewSettingsHandler(nil, "test-password")
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", bytes.NewReader([]byte(`{"weight_unit":`)))
	w := httptest.NewRecorder()

	h.UpdateSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestSettingsHandlerStartFreshInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewSettingsHandler(nil, "test-password")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings/start-fresh", bytes.NewReader([]byte(`{"password":`)))
	w := httptest.NewRecorder()

	h.StartFresh(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

var (
	settingsMigrationsOnce sync.Once
	settingsMigrationsErr  error
)

func newContainerizedSettingsHandler(t *testing.T) (*SettingsHandler, *sql.DB) {
	t.Helper()

	if os.Getenv("RUN_CONTAINERIZED_TESTS") != "true" {
		t.Skip("containerized integration tests are disabled")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is required for integration tests")
	}

	dbConn, err := db.Connect(databaseURL)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}

	settingsMigrationsOnce.Do(func() {
		settingsMigrationsErr = migrations.Run(context.Background(), dbConn, "up", nil)
	})
	if settingsMigrationsErr != nil {
		_ = dbConn.Close()
		t.Fatalf("run migrations: %v", settingsMigrationsErr)
	}

	if _, err := dbConn.ExecContext(context.Background(), "UPDATE settings SET weight_unit='g', distance_unit='km', temperature_unit='c', volume_unit='ml' WHERE id=1"); err != nil {
		_ = dbConn.Close()
		t.Fatalf("reset settings row: %v", err)
	}

	return NewSettingsHandler(store.New(dbConn), "pw123"), dbConn
}

func TestSettingsHandlerIntegrationGetAndUpdate(t *testing.T) {
	h, dbConn := newContainerizedSettingsHandler(t)
	defer func() { _ = dbConn.Close() }()

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/settings", http.NoBody)
	getW := httptest.NewRecorder()
	h.GetSettings(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 from get settings, got %d", getW.Code)
	}

	var before api.Settings
	if err := json.NewDecoder(getW.Body).Decode(&before); err != nil {
		t.Fatalf("decode get settings response: %v", err)
	}
	if before.WeightUnit != api.SettingsWeightUnitG {
		t.Fatalf("expected default weight unit g, got %q", before.WeightUnit)
	}

	weightUnit := api.SettingsUpdateWeightUnitOz
	distanceUnit := api.SettingsUpdateDistanceUnitMi
	temperatureUnit := api.SettingsUpdateTemperatureUnitF
	volumeUnit := api.SettingsUpdateVolumeUnitFlOz
	updateBody, err := json.Marshal(api.SettingsUpdate{
		WeightUnit:      &weightUnit,
		DistanceUnit:    &distanceUnit,
		TemperatureUnit: &temperatureUnit,
		VolumeUnit:      &volumeUnit,
	})
	if err != nil {
		t.Fatalf("marshal update settings body: %v", err)
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", bytes.NewReader(updateBody))
	updateW := httptest.NewRecorder()
	h.UpdateSettings(updateW, updateReq)
	if updateW.Code != http.StatusOK {
		t.Fatalf("expected 200 from update settings, got %d", updateW.Code)
	}

	var updated api.Settings
	if err := json.NewDecoder(updateW.Body).Decode(&updated); err != nil {
		t.Fatalf("decode update settings response: %v", err)
	}
	if updated.WeightUnit != api.SettingsWeightUnitOz {
		t.Fatalf("expected updated weight unit oz, got %q", updated.WeightUnit)
	}
	if updated.DistanceUnit != api.SettingsDistanceUnitMi {
		t.Fatalf("expected updated distance unit mi, got %q", updated.DistanceUnit)
	}
	if updated.TemperatureUnit != api.SettingsTemperatureUnitF {
		t.Fatalf("expected updated temperature unit f, got %q", updated.TemperatureUnit)
	}
	if updated.VolumeUnit != api.SettingsVolumeUnitFlOz {
		t.Fatalf("expected updated volume unit fl_oz, got %q", updated.VolumeUnit)
	}
}

func TestSettingsHandlerIntegrationStartFresh(t *testing.T) {
	h, dbConn := newContainerizedSettingsHandler(t)
	defer func() { _ = dbConn.Close() }()

	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO manufacturers (name) VALUES ($1)", "Reset Test Manufacturer"); err != nil {
		t.Fatalf("seed manufacturer: %v", err)
	}
	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO persons (name, body_weight_grams) VALUES ($1, $2)", "Reset Test Person", 70000); err != nil {
		t.Fatalf("seed person: %v", err)
	}

	wrongPasswordReq := httptest.NewRequest(http.MethodPost, "/api/v1/settings/start-fresh", bytes.NewReader([]byte(`{"password":"wrong"}`)))
	wrongPasswordW := httptest.NewRecorder()
	h.StartFresh(wrongPasswordW, wrongPasswordReq)
	if wrongPasswordW.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong password, got %d", wrongPasswordW.Code)
	}

	startFreshReq := httptest.NewRequest(http.MethodPost, "/api/v1/settings/start-fresh", bytes.NewReader([]byte(`{"password":"pw123"}`)))
	startFreshW := httptest.NewRecorder()
	h.StartFresh(startFreshW, startFreshReq)
	if startFreshW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 from start fresh, got %d", startFreshW.Code)
	}

	var personsCount int
	if err := dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM persons").Scan(&personsCount); err != nil {
		t.Fatalf("count persons: %v", err)
	}
	if personsCount != 0 {
		t.Fatalf("expected persons to be cleared, got count %d", personsCount)
	}

	var manufacturersCount int
	if err := dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM manufacturers").Scan(&manufacturersCount); err != nil {
		t.Fatalf("count manufacturers: %v", err)
	}
	if manufacturersCount != 0 {
		t.Fatalf("expected manufacturers to be cleared, got count %d", manufacturersCount)
	}
}

func TestSettingsHandlerIntegrationStartFreshReseed(t *testing.T) {
	h, dbConn := newContainerizedSettingsHandler(t)
	defer func() { _ = dbConn.Close() }()

	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO persons (name, body_weight_grams) VALUES ($1, $2)", "Reset Test Person", 70000); err != nil {
		t.Fatalf("seed person: %v", err)
	}

	startFreshReq := httptest.NewRequest(http.MethodPost, "/api/v1/settings/start-fresh", bytes.NewReader([]byte(`{"password":"pw123","reseed":true}`)))
	startFreshW := httptest.NewRecorder()
	h.StartFresh(startFreshW, startFreshReq)
	if startFreshW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 from start fresh, got %d", startFreshW.Code)
	}

	var personsCount int
	if err := dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM persons").Scan(&personsCount); err != nil {
		t.Fatalf("count persons: %v", err)
	}
	if personsCount != 0 {
		t.Fatalf("expected user data to be cleared, got persons count %d", personsCount)
	}

	// Catalog seed data should be restored when reseed is requested.
	var manufacturersCount int
	if err := dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM manufacturers").Scan(&manufacturersCount); err != nil {
		t.Fatalf("count manufacturers: %v", err)
	}
	if manufacturersCount == 0 {
		t.Fatalf("expected manufacturers to be reseeded, got count %d", manufacturersCount)
	}

	var labelsCount int
	if err := dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM labels").Scan(&labelsCount); err != nil {
		t.Fatalf("count labels: %v", err)
	}
	if labelsCount == 0 {
		t.Fatalf("expected labels to be reseeded, got count %d", labelsCount)
	}
}
