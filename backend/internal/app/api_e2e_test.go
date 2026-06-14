package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/via-justa/overpacked-app/backend/internal/auth"
	"github.com/via-justa/overpacked-app/backend/internal/backup"
	"github.com/via-justa/overpacked-app/backend/internal/config"
	"github.com/via-justa/overpacked-app/backend/internal/db"
	"github.com/via-justa/overpacked-app/backend/internal/http/handlers"
	"github.com/via-justa/overpacked-app/backend/internal/migrations"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

const (
	e2eUsername = "admin"
	e2ePassword = "pw123"
)

var (
	e2eMigrationsOnce sync.Once
	e2eMigrationsErr  error
)

// TestAPIEndToEnd drives the fully assembled router (middleware chain + routes)
// over real HTTP via httptest, against a containerized Postgres. It is the Go
// replacement for the former dev/scripts/full_api_curl_test.sh suite, and being
// in-process it also produces coverage across routing, auth, handlers and stores.
func TestAPIEndToEnd(t *testing.T) {
	c, cleanup := newE2EServer(t)
	defer cleanup()

	state := &e2eState{}
	authAndSettings(t, c)
	personsFlow(t, c, state)
	manufacturersFlow(t, c, state)
	itemsFlow(t, c, state)
	searchFlow(t, c, state)
	labelsFlow(t, c, state)
	itemTypesFlow(t, c, state)
	setsFlow(t, c, state)
	tripsFlow(t, c, state)
	packingListsFlow(t, c, state)
	backupFlow(t, c)
	cleanupFlow(t, c, state)
	startFreshFlow(t, c, state)

	c.expect(http.MethodPost, "/api/v1/auth/logout", "", true, http.StatusNoContent)
}

// TestAppLifecycle exercises the production assembly: New wires the server and
// scheduler against a real database, StartScheduler boots the background loop,
// and Shutdown tears it all down cleanly.
func TestAppLifecycle(t *testing.T) {
	if os.Getenv("RUN_CONTAINERIZED_TESTS") != "true" {
		t.Skip("containerized integration tests are disabled")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is required for integration tests")
	}

	// New does not migrate (main.go does that separately), so ensure the schema
	// exists before asserting the server/scheduler assemble.
	migConn, err := db.Connect(databaseURL)
	if err != nil {
		t.Fatalf("connect for migrations: %v", err)
	}
	e2eMigrationsOnce.Do(func() {
		e2eMigrationsErr = migrations.Run(context.Background(), migConn, "up", nil)
	})
	_ = migConn.Close()
	if e2eMigrationsErr != nil {
		t.Fatalf("run migrations: %v", e2eMigrationsErr)
	}

	cfg := &config.Config{
		DatabaseURL:   databaseURL,
		ServerAddr:    "127.0.0.1:0",
		AppUsername:   e2eUsername,
		AppPassword:   e2ePassword,
		JWTSecret:     "test-secret",
		BackupBaseDir: t.TempDir(),
	}

	ctx := context.Background()
	application, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	if err := application.StartScheduler(ctx); err != nil {
		t.Fatalf("StartScheduler: %v", err)
	}
	if err := application.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

// e2eState threads IDs captured from earlier responses into later requests.
type e2eState struct {
	personID       string
	manufacturerID string
	itemID         string
	labelID        string
	customTypeID   string
	customItemID   string
	setID          string
	tripID         string
	personItemID   string
	packID         string
	packingListID  string
}

// ── server + client harness ─────────────────────────────────────────────────

func newE2EServer(t *testing.T) (*apiClient, func()) {
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

	e2eMigrationsOnce.Do(func() {
		e2eMigrationsErr = migrations.Run(context.Background(), dbConn, "up", nil)
	})
	if e2eMigrationsErr != nil {
		_ = dbConn.Close()
		t.Fatalf("run migrations: %v", e2eMigrationsErr)
	}

	// Start from an empty user dataset (no seeds) so the empty-list assertions
	// hold and the run is order-independent under the shared test database.
	if _, err := dbConn.ExecContext(context.Background(), `
		TRUNCATE TABLE trip_person_items, trip_person_packs, trip_persons, trips,
			pack_items, packs, set_items, item_sets, item_labels, packing_list_labels,
			items, packing_lists, labels, manufacturers, persons RESTART IDENTITY CASCADE`); err != nil {
		_ = dbConn.Close()
		t.Fatalf("truncate tables: %v", err)
	}
	if _, err := dbConn.ExecContext(context.Background(), `DELETE FROM item_types WHERE is_system = FALSE`); err != nil {
		_ = dbConn.Close()
		t.Fatalf("delete custom item types: %v", err)
	}

	authService, err := auth.NewService(e2eUsername, e2ePassword, "test-secret")
	if err != nil {
		_ = dbConn.Close()
		t.Fatalf("auth service: %v", err)
	}
	st := store.New(dbConn)
	backupSvc := backup.NewService(dbConn, t.TempDir())
	scheduler := backup.NewScheduler(backupSvc, st.BackupConfig)
	backupHandler := handlers.NewBackupHandler(backupSvc, st, scheduler, e2ePassword)

	srv := httptest.NewServer(NewHTTPHandler(authService, st, e2ePassword, backupHandler))

	jar, err := cookiejar.New(nil)
	if err != nil {
		srv.Close()
		_ = dbConn.Close()
		t.Fatalf("cookie jar: %v", err)
	}

	c := &apiClient{t: t, baseURL: srv.URL, http: &http.Client{Jar: jar}}
	return c, func() {
		srv.Close()
		_ = dbConn.Close()
	}
}

type apiClient struct {
	t       *testing.T
	baseURL string
	http    *http.Client
	token   string
}

// do issues a request and returns the status code and raw body.
func (c *apiClient) do(method, path, body string, useAuth bool) (int, []byte) {
	c.t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		c.t.Fatalf("build request %s %s: %v", method, path, err)
	}
	req.Header.Set("Accept", "application/json")
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if useAuth && c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		c.t.Fatalf("request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}

// expect issues an authenticated request and fails unless the status matches.
func (c *apiClient) expect(method, path, body string, useAuth bool, want int) []byte {
	c.t.Helper()
	status, data := c.do(method, path, body, useAuth)
	if status != want {
		c.t.Fatalf("%s %s: expected status %d, got %d: %s", method, path, want, status, data)
	}
	return data
}

// ── decode/assert helpers ───────────────────────────────────────────────────

func asMap(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("decode object: %v: %s", err, data)
	}
	return m
}

func asList(t *testing.T, data []byte) []any {
	t.Helper()
	var l []any
	if err := json.Unmarshal(data, &l); err != nil {
		t.Fatalf("decode array: %v: %s", err, data)
	}
	return l
}

func strOf(t *testing.T, m map[string]any, key string) string {
	t.Helper()
	v, ok := m[key].(string)
	if !ok {
		t.Fatalf("expected string field %q in %v", key, m)
	}
	return v
}

func wantStr(t *testing.T, m map[string]any, key, want string) {
	t.Helper()
	if got := strOf(t, m, key); got != want {
		t.Fatalf("field %q: expected %q, got %q", key, want, got)
	}
}

func wantNum(t *testing.T, m map[string]any, key string, want float64) {
	t.Helper()
	got, ok := m[key].(float64)
	if !ok || got != want {
		t.Fatalf("field %q: expected %v, got %v", key, want, m[key])
	}
}

func wantLen(t *testing.T, data []byte, want int) []any {
	t.Helper()
	l := asList(t, data)
	if len(l) != want {
		t.Fatalf("expected list length %d, got %d: %s", want, len(l), data)
	}
	return l
}

// ── flows ────────────────────────────────────────────────────────────────────

func authAndSettings(t *testing.T, c *apiClient) {
	c.expect(http.MethodGet, "/health", "", false, http.StatusOK)

	login := asMap(t, c.expect(http.MethodPost, "/api/v1/auth/login",
		`{"username":"`+e2eUsername+`","password":"`+e2ePassword+`"}`, false, http.StatusOK))
	c.token = strOf(t, login, "access_token")
	wantStr(t, login, "token_type", "Bearer")

	// Refresh carries the HttpOnly op_refresh cookie via the client's jar — no body.
	refreshed := asMap(t, c.expect(http.MethodPost, "/api/v1/auth/refresh", "", false, http.StatusOK))
	c.token = strOf(t, refreshed, "access_token")

	c.expect(http.MethodGet, "/api/v1/settings", "", true, http.StatusOK)
	updated := asMap(t, c.expect(http.MethodPatch, "/api/v1/settings",
		`{"weight_unit":"g","distance_unit":"km","temperature_unit":"c","volume_unit":"ml"}`, true, http.StatusOK))
	wantStr(t, updated, "weight_unit", "g")
	wantStr(t, updated, "distance_unit", "km")
}

func personsFlow(t *testing.T, c *apiClient, s *e2eState) {
	wantLen(t, c.expect(http.MethodGet, "/api/v1/persons", "", true, http.StatusOK), 0)

	created := asMap(t, c.expect(http.MethodPost, "/api/v1/persons",
		`{"name":"API Test Person","gender":"other","birthdate":"1990-01-01","body_weight_grams":70000,"conditioning_level":"average"}`,
		true, http.StatusCreated))
	wantStr(t, created, "conditioning_level", "average")
	wantNum(t, created, "body_weight_grams", 70000)
	s.personID = strOf(t, created, "id")

	got := asMap(t, c.expect(http.MethodGet, "/api/v1/persons/"+s.personID, "", true, http.StatusOK))
	wantStr(t, got, "conditioning_level", "average")

	updated := asMap(t, c.expect(http.MethodPatch, "/api/v1/persons/"+s.personID,
		`{"name":"API Test Person Updated","conditioning_level":"athletic","body_weight_grams":72000}`, true, http.StatusOK))
	wantStr(t, updated, "name", "API Test Person Updated")
	wantStr(t, updated, "conditioning_level", "athletic")
}

func manufacturersFlow(t *testing.T, c *apiClient, s *e2eState) {
	wantLen(t, c.expect(http.MethodGet, "/api/v1/manufacturers", "", true, http.StatusOK), 0)

	created := asMap(t, c.expect(http.MethodPost, "/api/v1/manufacturers",
		`{"name":"API Test Manufacturer","website":"https://example.com"}`, true, http.StatusCreated))
	s.manufacturerID = strOf(t, created, "id")

	c.expect(http.MethodGet, "/api/v1/manufacturers/"+s.manufacturerID, "", true, http.StatusOK)
	updated := asMap(t, c.expect(http.MethodPatch, "/api/v1/manufacturers/"+s.manufacturerID,
		`{"name":"API Test Manufacturer Updated"}`, true, http.StatusOK))
	wantStr(t, updated, "name", "API Test Manufacturer Updated")
}

func itemsFlow(t *testing.T, c *apiClient, s *e2eState) {
	wantLen(t, c.expect(http.MethodGet, "/api/v1/items", "", true, http.StatusOK), 0)

	created := asMap(t, c.expect(http.MethodPost, "/api/v1/items",
		`{"name":"API Test Item","type":"consumable","is_active":true,"manufacturer_id":"`+s.manufacturerID+`","weight_grams":100,"volume_ml":50}`,
		true, http.StatusCreated))
	wantStr(t, created, "type", "consumable")
	wantNum(t, created, "weight_grams", 100)
	s.itemID = strOf(t, created, "id")

	got := asMap(t, c.expect(http.MethodGet, "/api/v1/items/"+s.itemID, "", true, http.StatusOK))
	wantStr(t, got, "name", "API Test Item")

	updated := asMap(t, c.expect(http.MethodPatch, "/api/v1/items/"+s.itemID,
		`{"name":"API Test Item Updated","weight_grams":110}`, true, http.StatusOK))
	wantStr(t, updated, "name", "API Test Item Updated")
	wantNum(t, updated, "weight_grams", 110)
}

func searchFlow(t *testing.T, c *apiClient, s *e2eState) {
	c.expect(http.MethodGet, "/api/v1/search?q=a", "", true, http.StatusBadRequest)

	results := asList(t, c.expect(http.MethodGet, "/api/v1/search?q=API%20Test", "", true, http.StatusOK))
	if !anyMatch(results, func(m map[string]any) bool { return m["id"] == s.itemID && m["entity_type"] == "item" }) {
		t.Fatalf("global search did not return the created item: %v", results)
	}

	persons := asList(t, c.expect(http.MethodGet, "/api/v1/search?q=API%20Test&types=person", "", true, http.StatusOK))
	for _, r := range persons {
		if m, ok := r.(map[string]any); ok && m["entity_type"] != "person" {
			t.Fatalf("type-filtered search returned non-person: %v", m)
		}
	}
}

func labelsFlow(t *testing.T, c *apiClient, s *e2eState) {
	wantLen(t, c.expect(http.MethodGet, "/api/v1/labels", "", true, http.StatusOK), 0)

	created := asMap(t, c.expect(http.MethodPost, "/api/v1/labels",
		`{"name":"Ultralight","color":"#FF5733"}`, true, http.StatusCreated))
	wantStr(t, created, "name", "Ultralight")
	s.labelID = strOf(t, created, "id")

	updated := asMap(t, c.expect(http.MethodPatch, "/api/v1/labels/"+s.labelID,
		`{"name":"Ultra-Lightweight","color":"#00FF00"}`, true, http.StatusOK))
	wantStr(t, updated, "name", "Ultra-Lightweight")

	// item ↔ label link lifecycle
	wantLen(t, c.expect(http.MethodGet, "/api/v1/items/"+s.itemID+"/labels", "", true, http.StatusOK), 0)
	c.expect(http.MethodPost, "/api/v1/items/"+s.itemID+"/labels", `{"label_id":"`+s.labelID+`"}`, true, http.StatusCreated)
	linked := wantLen(t, c.expect(http.MethodGet, "/api/v1/items/"+s.itemID+"/labels", "", true, http.StatusOK), 1)
	wantStr(t, linked[0].(map[string]any), "name", "Ultra-Lightweight")
	c.expect(http.MethodDelete, "/api/v1/items/"+s.itemID+"/labels/"+s.labelID, "", true, http.StatusNoContent)
	wantLen(t, c.expect(http.MethodGet, "/api/v1/items/"+s.itemID+"/labels", "", true, http.StatusOK), 0)
}

func itemTypesFlow(t *testing.T, c *apiClient, s *e2eState) {
	s.customTypeID = "api_test_type_e2e"

	types := asList(t, c.expect(http.MethodGet, "/api/v1/item-types", "", true, http.StatusOK))
	if !anyMatch(types, func(m map[string]any) bool { return m["id"] == "consumable" }) {
		t.Fatalf("expected system item type 'consumable' to exist: %v", types)
	}

	created := asMap(t, c.expect(http.MethodPost, "/api/v1/item-types",
		`{"id":"`+s.customTypeID+`","name":"API Test Custom Type","description":"created by e2e","base_profile":"electronics"}`,
		true, http.StatusCreated))
	wantStr(t, created, "id", s.customTypeID)
	if created["is_system"] != false {
		t.Fatalf("expected is_system=false, got %v", created["is_system"])
	}

	c.expect(http.MethodGet, "/api/v1/item-types/"+s.customTypeID, "", true, http.StatusOK)
	c.expect(http.MethodPatch, "/api/v1/item-types/"+s.customTypeID,
		`{"name":"API Test Custom Type Updated","description":"updated"}`, true, http.StatusOK)

	// PUT replaces the full field set; expect exactly the three provided.
	fields := wantLen(t, c.expect(http.MethodPut, "/api/v1/item-types/"+s.customTypeID+"/fields",
		`{"fields":[`+
			`{"field_key":"output_watts","field_label":"Output watts","data_type":"number","is_required":false,"sort_order":1,"unit":"W"},`+
			`{"field_key":"has_usb_pd","field_label":"Has USB PD","data_type":"boolean","is_required":false,"sort_order":2},`+
			`{"field_key":"battery_chemistry","field_label":"Battery chemistry","data_type":"enum","is_required":false,"sort_order":3,"enum_options":["li-ion","lifepo4"]}`+
			`]}`, true, http.StatusOK), 3)
	if !anyMatch(fields, func(m map[string]any) bool {
		return m["field_key"] == "output_watts" && m["data_type"] == "number"
	}) {
		t.Fatalf("replaced fields missing output_watts: %v", fields)
	}
	wantLen(t, c.expect(http.MethodGet, "/api/v1/item-types/"+s.customTypeID+"/fields", "", true, http.StatusOK), 3)

	custom := asMap(t, c.expect(http.MethodPost, "/api/v1/items",
		`{"name":"API Test Custom Item","type":"`+s.customTypeID+`","is_active":true,"manufacturer_id":"`+s.manufacturerID+`","weight_grams":220,"volume_ml":330,"attributes":{"output_watts":30,"has_usb_pd":true,"battery_chemistry":"li-ion"}}`,
		true, http.StatusCreated))
	wantStr(t, custom, "type", s.customTypeID)
	s.customItemID = strOf(t, custom, "id")

	updated := asMap(t, c.expect(http.MethodPatch, "/api/v1/items/"+s.customItemID,
		`{"name":"API Test Custom Item Updated","attributes":{"output_watts":45,"has_usb_pd":true,"battery_chemistry":"lifepo4"}}`,
		true, http.StatusOK))
	attrs, ok := updated["attributes"].(map[string]any)
	if !ok || attrs["output_watts"] != float64(45) || attrs["battery_chemistry"] != "lifepo4" {
		t.Fatalf("custom item attributes not updated: %v", updated["attributes"])
	}
}

func setsFlow(t *testing.T, c *apiClient, s *e2eState) {
	wantLen(t, c.expect(http.MethodGet, "/api/v1/sets", "", true, http.StatusOK), 0)

	created := asMap(t, c.expect(http.MethodPost, "/api/v1/sets",
		`{"name":"API Test Set","set_category":"consumable"}`, true, http.StatusCreated))
	s.setID = strOf(t, created, "id")

	c.expect(http.MethodGet, "/api/v1/sets/"+s.setID, "", true, http.StatusOK)
	c.expect(http.MethodPatch, "/api/v1/sets/"+s.setID,
		`{"name":"API Test Set Updated","set_category":"consumable"}`, true, http.StatusOK)

	wantLen(t, c.expect(http.MethodGet, "/api/v1/sets/"+s.setID+"/items", "", true, http.StatusOK), 0)
	c.expect(http.MethodPost, "/api/v1/sets/"+s.setID+"/items",
		`{"item_id":"`+s.itemID+`","quantity":2,"notes":"test note","sort_order":1}`, true, http.StatusCreated)
	c.expect(http.MethodPatch, "/api/v1/sets/"+s.setID+"/items/"+s.itemID,
		`{"quantity":3,"notes":"updated note"}`, true, http.StatusOK)
	c.expect(http.MethodDelete, "/api/v1/sets/"+s.setID+"/items/"+s.itemID, "", true, http.StatusNoContent)
	wantLen(t, c.expect(http.MethodGet, "/api/v1/sets/"+s.setID+"/items", "", true, http.StatusOK), 0)

	c.expect(http.MethodDelete, "/api/v1/sets/"+s.setID, "", true, http.StatusNoContent)
	wantLen(t, c.expect(http.MethodGet, "/api/v1/sets", "", true, http.StatusOK), 0)
}

func tripsFlow(t *testing.T, c *apiClient, s *e2eState) {
	wantLen(t, c.expect(http.MethodGet, "/api/v1/trips", "", true, http.StatusOK), 0)

	created := asMap(t, c.expect(http.MethodPost, "/api/v1/trips",
		`{"name":"API Test Trip","trip_type":"overnight","notes":"Test trip notes","total_distance_km":15.5}`,
		true, http.StatusCreated))
	s.tripID = strOf(t, created, "id")

	trip := asMap(t, c.expect(http.MethodGet, "/api/v1/trips/"+s.tripID, "", true, http.StatusOK))
	wantStr(t, trip, "name", "API Test Trip")
	wantLenVal(t, trip["persons"], 0)

	updated := asMap(t, c.expect(http.MethodPatch, "/api/v1/trips/"+s.tripID,
		`{"name":"Updated Trip Name","total_distance_km":20.0}`, true, http.StatusOK))
	wantStr(t, updated, "name", "Updated Trip Name")

	c.expect(http.MethodPost, "/api/v1/trips/"+s.tripID+"/persons", `{"person_id":"`+s.personID+`"}`, true, http.StatusCreated)
	withPerson := asMap(t, c.expect(http.MethodGet, "/api/v1/trips/"+s.tripID, "", true, http.StatusOK))
	persons := wantLenVal(t, withPerson["persons"], 1)
	wantStr(t, persons[0].(map[string]any), "person_id", s.personID)

	tripPersonItemsFlow(t, c, s)
	tripPacksFlow(t, c, s)

	// Tear the trip membership back down.
	c.expect(http.MethodDelete, "/api/v1/trips/"+s.tripID+"/persons/"+s.personID, "", true, http.StatusNoContent)
	after := asMap(t, c.expect(http.MethodGet, "/api/v1/trips/"+s.tripID, "", true, http.StatusOK))
	wantLenVal(t, after["persons"], 0)
}

func tripPersonItemsFlow(t *testing.T, c *apiClient, s *e2eState) {
	base := "/api/v1/trips/" + s.tripID + "/persons/" + s.personID
	item := asMap(t, c.expect(http.MethodPost, base+"/items",
		`{"item_id":"`+s.itemID+`","quantity":1,"carry_status":"worn","notes":"Worn directly"}`, true, http.StatusCreated))
	s.personItemID = strOf(t, item, "id")

	updated := asMap(t, c.expect(http.MethodPatch, base+"/items/"+s.personItemID,
		`{"quantity":2,"notes":"Updated notes"}`, true, http.StatusOK))
	wantNum(t, updated, "quantity", 2)

	trip := asMap(t, c.expect(http.MethodGet, "/api/v1/trips/"+s.tripID, "", true, http.StatusOK))
	person := listVal(t, trip["persons"])[0].(map[string]any)
	wantLenVal(t, person["items"], 1)

	c.expect(http.MethodDelete, base+"/items/"+s.personItemID, "", true, http.StatusNoContent)
}

func tripPacksFlow(t *testing.T, c *apiClient, s *e2eState) {
	base := "/api/v1/trips/" + s.tripID + "/persons/" + s.personID
	pack := asMap(t, c.expect(http.MethodPost, base+"/packs",
		`{"name":"Test Trip Pack","trip_type":"overnight","notes":"Pack for overnight trip"}`, true, http.StatusCreated))
	s.packID = strOf(t, pack, "pack_id")

	trip := asMap(t, c.expect(http.MethodGet, "/api/v1/trips/"+s.tripID, "", true, http.StatusOK))
	person := listVal(t, trip["persons"])[0].(map[string]any)
	packs := wantLenVal(t, person["packs"], 1)
	packObj := mustMap(t, packs[0].(map[string]any)["pack"])
	wantStr(t, packObj, "name", "Test Trip Pack")

	c.expect(http.MethodPost, base+"/packs/"+s.packID+"/items",
		`{"item_id":"`+s.itemID+`","quantity":3,"carry_status":"packed","notes":"Packed in pack"}`, true, http.StatusCreated)
	updated := asMap(t, c.expect(http.MethodPatch, base+"/packs/"+s.packID+"/items/"+s.itemID,
		`{"quantity":5,"notes":"Updated pack item notes"}`, true, http.StatusOK))
	wantNum(t, updated, "quantity", 5)

	c.expect(http.MethodDelete, base+"/packs/"+s.packID+"/items/"+s.itemID, "", true, http.StatusNoContent)
	c.expect(http.MethodDelete, base+"/packs/"+s.packID, "", true, http.StatusNoContent)
}

func packingListsFlow(t *testing.T, c *apiClient, s *e2eState) {
	wantLen(t, c.expect(http.MethodGet, "/api/v1/packing-lists", "", true, http.StatusOK), 0)

	created := asMap(t, c.expect(http.MethodPost, "/api/v1/packing-lists",
		`{"name":"Summer Packing List","description":"Essential items for summer trips"}`, true, http.StatusCreated))
	s.packingListID = strOf(t, created, "id")

	c.expect(http.MethodGet, "/api/v1/packing-lists/"+s.packingListID, "", true, http.StatusOK)
	c.expect(http.MethodPatch, "/api/v1/packing-lists/"+s.packingListID,
		`{"name":"Updated Packing List","description":"Updated description"}`, true, http.StatusOK)

	wantLen(t, c.expect(http.MethodGet, "/api/v1/packing-lists/"+s.packingListID+"/labels", "", true, http.StatusOK), 0)
	c.expect(http.MethodPost, "/api/v1/packing-lists/"+s.packingListID+"/labels", `{"label_id":"`+s.labelID+`"}`, true, http.StatusCreated)
	labels := wantLen(t, c.expect(http.MethodGet, "/api/v1/packing-lists/"+s.packingListID+"/labels", "", true, http.StatusOK), 1)
	wantStr(t, labels[0].(map[string]any), "name", "Ultra-Lightweight")
	c.expect(http.MethodDelete, "/api/v1/packing-lists/"+s.packingListID+"/labels/"+s.labelID, "", true, http.StatusNoContent)
	c.expect(http.MethodDelete, "/api/v1/packing-lists/"+s.packingListID, "", true, http.StatusNoContent)
}

func backupFlow(t *testing.T, c *apiClient) {
	c.expect(http.MethodGet, "/api/v1/backup/config", "", true, http.StatusOK)
	c.expect(http.MethodPut, "/api/v1/backup/config",
		`{"enabled":false,"schedule":"0 3 * * *","retention_count":7}`, true, http.StatusOK)
	c.expect(http.MethodPost, "/api/v1/backup/run", "", true, http.StatusOK)

	// Export the full archive over HTTP, then import it back (merge) to drive the
	// upload path. Also exercise the items CSV/zip export endpoint.
	status, archive := c.do(http.MethodGet, "/api/v1/backup/export", "", true)
	if status != http.StatusOK {
		t.Fatalf("GET /api/v1/backup/export: expected 200, got %d", status)
	}
	c.expect(http.MethodGet, "/api/v1/export/items", "", true, http.StatusOK)
	c.importArchive(t, archive, "merge", http.StatusOK)
}

// importArchive POSTs a backup ZIP as multipart/form-data (file + password + mode).
func (c *apiClient) importArchive(t *testing.T, archive []byte, mode string, want int) {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("password", e2ePassword)
	_ = mw.WriteField("mode", mode)
	fw, err := mw.CreateFormFile("file", "backup.zip")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write(archive); err != nil {
		t.Fatalf("write archive: %v", err)
	}
	_ = mw.Close()

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/backup/import", &body)
	if err != nil {
		t.Fatalf("build import request: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatalf("import request: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != want {
		t.Fatalf("import: expected %d, got %d: %s", want, resp.StatusCode, data)
	}
}

func cleanupFlow(t *testing.T, c *apiClient, s *e2eState) {
	c.expect(http.MethodDelete, "/api/v1/labels/"+s.labelID, "", true, http.StatusNoContent)
	c.expect(http.MethodDelete, "/api/v1/trips/"+s.tripID, "", true, http.StatusNoContent)
	c.expect(http.MethodDelete, "/api/v1/items/"+s.customItemID, "", true, http.StatusNoContent)
	c.expect(http.MethodDelete, "/api/v1/items/"+s.itemID, "", true, http.StatusNoContent)
	c.expect(http.MethodDelete, "/api/v1/item-types/"+s.customTypeID, "", true, http.StatusNoContent)
	c.expect(http.MethodDelete, "/api/v1/manufacturers/"+s.manufacturerID, "", true, http.StatusNoContent)
	c.expect(http.MethodDelete, "/api/v1/persons/"+s.personID, "", true, http.StatusNoContent)
}

func startFreshFlow(t *testing.T, c *apiClient, s *e2eState) {
	c.expect(http.MethodPost, "/api/v1/settings/start-fresh", `{"password":"wrong-password"}`, true, http.StatusUnauthorized)
	c.expect(http.MethodPost, "/api/v1/settings/start-fresh", `{"password":"`+e2ePassword+`"}`, true, http.StatusNoContent)

	wantLen(t, c.expect(http.MethodGet, "/api/v1/persons", "", true, http.StatusOK), 0)
	wantLen(t, c.expect(http.MethodGet, "/api/v1/manufacturers", "", true, http.StatusOK), 0)
	wantLen(t, c.expect(http.MethodGet, "/api/v1/items", "", true, http.StatusOK), 0)
	wantLen(t, c.expect(http.MethodGet, "/api/v1/sets", "", true, http.StatusOK), 0)
	wantLen(t, c.expect(http.MethodGet, "/api/v1/trips", "", true, http.StatusOK), 0)
	c.expect(http.MethodGet, "/api/v1/item-types/"+s.customTypeID, "", true, http.StatusNotFound)

	settings := asMap(t, c.expect(http.MethodGet, "/api/v1/settings", "", true, http.StatusOK))
	wantStr(t, settings, "weight_unit", "g")
	wantStr(t, settings, "distance_unit", "km")
	wantStr(t, settings, "temperature_unit", "c")
	wantStr(t, settings, "volume_unit", "ml")
	wantStr(t, settings, "currency", "eur")
}

// ── small generic helpers ────────────────────────────────────────────────────

func anyMatch(list []any, pred func(map[string]any) bool) bool {
	for _, v := range list {
		if m, ok := v.(map[string]any); ok && pred(m) {
			return true
		}
	}
	return false
}

// listVal asserts an already-decoded JSON value is an array.
func listVal(t *testing.T, v any) []any {
	t.Helper()
	l, ok := v.([]any)
	if !ok {
		t.Fatalf("expected nested array, got %T", v)
	}
	return l
}

// wantLenVal asserts an already-decoded JSON array value has the expected length.
func wantLenVal(t *testing.T, v any, want int) []any {
	t.Helper()
	l := listVal(t, v)
	if len(l) != want {
		t.Fatalf("expected nested array length %d, got %d", want, len(l))
	}
	return l
}

func mustMap(t *testing.T, v any) map[string]any {
	t.Helper()
	m, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("expected nested object, got %T", v)
	}
	return m
}
