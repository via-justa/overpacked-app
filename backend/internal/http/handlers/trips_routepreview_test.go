package handlers

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/via-justa/overpacked-app/backend/internal/api"
)

func TestRoutePreviewHostHelpers(t *testing.T) {
	t.Parallel()

	if !isKomootHost("komoot.com") || !isKomootHost("www.komoot.de") || isKomootHost("evil.com") {
		t.Error("isKomootHost policy wrong")
	}
	if !isStravaHost("strava.com") || !isStravaHost("foo.strava.app.link") || isStravaHost("evil.com") {
		t.Error("isStravaHost policy wrong")
	}
	if got := normalizeHost("Komoot.COM:443"); got != "komoot.com" {
		t.Errorf("normalizeHost = %q, want komoot.com", got)
	}
	if !hostAllowedForService("wanderer", "anything.example") {
		t.Error("wanderer should allow any non-empty host")
	}
	if hostAllowedForService("wanderer", "") || hostAllowedForService("komoot", "evil.com") || hostAllowedForService("bogus", "x") {
		t.Error("hostAllowedForService policy wrong")
	}

	for _, s := range []string{"127.0.0.1", "10.0.0.1", "::1", "169.254.0.1", "224.0.0.1", "0.0.0.0"} {
		if !isDisallowedIP(net.ParseIP(s)) {
			t.Errorf("expected %s to be disallowed", s)
		}
	}
	if isDisallowedIP(net.ParseIP("8.8.8.8")) {
		t.Error("public IP should be allowed")
	}

	// controlBlockPrivateAddr runs on the resolved dial address.
	if controlBlockPrivateAddr("tcp", "8.8.8.8:443", nil) != nil {
		t.Error("public address should be allowed to dial")
	}
	if controlBlockPrivateAddr("tcp", "127.0.0.1:443", nil) == nil {
		t.Error("loopback address should be blocked")
	}
	if controlBlockPrivateAddr("tcp", "garbage", nil) == nil {
		t.Error("malformed address should be blocked")
	}
}

func TestGetTripRoutePreviewValidation(t *testing.T) {
	t.Parallel()

	h := NewTripsHandler(nil) // validation/fetch only — no store access
	call := func(service, rawURL string) int {
		w := httptest.NewRecorder()
		h.GetTripRoutePreview(w, httptest.NewRequest(http.MethodGet, "/x", http.NoBody),
			api.GetTripRoutePreviewParamsService(service), api.GetTripRoutePreviewParams{Url: rawURL})
		return w.Code
	}

	if call("bogus", "https://komoot.com/tour/1") != http.StatusBadRequest {
		t.Error("unsupported service should be 400")
	}
	if call("komoot", "http://komoot.com/tour/1") != http.StatusBadRequest {
		t.Error("non-https url should be 400")
	}
	if call("komoot", "https://evil.com/tour/1") != http.StatusBadRequest {
		t.Error("host not matching service should be 400")
	}
	// Valid wanderer host that resolves to a blocked (loopback) address: passes
	// validation, the fetch is refused at dial time, and the handler returns 200
	// with an empty preview. No real network egress occurs.
	if call("wanderer", "https://127.0.0.1:9/") != http.StatusOK {
		t.Error("blocked-dial fetch should still yield 200 with an empty preview")
	}
}
