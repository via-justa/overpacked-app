package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

type TripsHandler struct {
	store *store.Store
}

const (
	notFoundErrMsg                = "not found"
	failedToGetErrMsg             = "failed to get"
	tripsErrInvalidRequestBody    = "invalid request body"
	tripsErrTripNotFound          = "trip" + notFoundErrMsg
	tripsErrFailedToGetTrip       = failedToGetErrMsg + " trip"
	tripsErrPersonNotFound        = "person not found in trip"
	tripsErrPackNotFound          = "pack" + notFoundErrMsg
	tripsErrItemNotFound          = "item" + notFoundErrMsg
	tripsErrFailedToGetTripPerson = failedToGetErrMsg + " trip person"
)

func NewTripsHandler(st *store.Store) *TripsHandler {
	return &TripsHandler{store: st}
}

// Route preview (Open Graph metadata) constants and helpers.
const (
	routePreviewServiceKomoot   = "komoot"
	routePreviewServiceStrava   = "strava"
	routePreviewServiceWanderer = "wanderer"

	routePreviewFetchTimeout = 6 * time.Second
	routePreviewMaxBodyBytes = 1 << 20 // 1 MiB cap to avoid resource exhaustion
)

// errBlockedAddress is returned by the dial control hook when a target resolves
// to a private, loopback, or otherwise non-public address.
var errBlockedAddress = errors.New("blocked non-public address")

// routePreviewClient resolves and blocks non-public IPs at dial time (after DNS
// resolution, so it is resistant to DNS rebinding) and uses a short timeout to
// limit SSRF impact. The same protection applies to any redirects it follows.
var routePreviewClient = &http.Client{
	Timeout: routePreviewFetchTimeout,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   routePreviewFetchTimeout,
			KeepAlive: routePreviewFetchTimeout,
			Control:   controlBlockPrivateAddr,
		}).DialContext,
	},
}

var (
	ogImageRegexp = regexp.MustCompile(`(?is)<meta[^>]+property=["']og:image["'][^>]+content=["']([^"']+)["']`)
	ogTitleRegexp = regexp.MustCompile(`(?is)<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']+)["']`)
)

// controlBlockPrivateAddr rejects connections to private, loopback, link-local,
// unspecified, or multicast addresses. It runs on the already-resolved address,
// which prevents SSRF via DNS rebinding.
func controlBlockPrivateAddr(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return errBlockedAddress
	}

	ip := net.ParseIP(host)
	if ip == nil || isDisallowedIP(ip) {
		return errBlockedAddress
	}

	return nil
}

// isDisallowedIP reports whether an IP must not be contacted by the preview fetcher.
func isDisallowedIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast()
}

// normalizeHost lowercases a URL host and strips any port suffix.
func normalizeHost(host string) string {
	host = strings.ToLower(host)
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}

// isKomootHost reports whether the host belongs to komoot's official domains.
func isKomootHost(host string) bool {
	return host == "komoot.com" || strings.HasSuffix(host, ".komoot.com") ||
		host == "komoot.de" || strings.HasSuffix(host, ".komoot.de")
}

// isStravaHost reports whether the host belongs to strava's official domains.
func isStravaHost(host string) bool {
	return host == "strava.com" || strings.HasSuffix(host, ".strava.com") ||
		host == "strava.app.link" || strings.HasSuffix(host, ".strava.app.link")
}

// hostAllowedForService enforces per-service host policy. komoot and strava are
// restricted to their official domains; wanderer accepts any public host (its
// SSRF protection is the dial-time private-address block).
func hostAllowedForService(service, host string) bool {
	switch service {
	case routePreviewServiceKomoot:
		return isKomootHost(host)
	case routePreviewServiceStrava:
		return isStravaHost(host)
	case routePreviewServiceWanderer:
		return host != ""
	default:
		return false
	}
}

// GetTripRoutePreview fetches Open Graph preview metadata for a public route share link.
func (h *TripsHandler) GetTripRoutePreview(w http.ResponseWriter, r *http.Request, service api.GetTripRoutePreviewParamsService, params api.GetTripRoutePreviewParams) {
	if !service.Valid() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported route service"})
		return
	}

	parsed, err := url.Parse(strings.TrimSpace(params.Url))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid route url"})
		return
	}

	if !hostAllowedForService(string(service), normalizeHost(parsed.Host)) {
		// Reject hosts that do not match the requested service to prevent abuse.
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url does not match service host"})
		return
	}

	preview := api.TripRoutePreview{Service: api.TripRoutePreviewService(service)}

	imageURL, title := fetchOpenGraphMeta(r.Context(), parsed.String())
	if imageURL != "" {
		preview.ImageUrl = &imageURL
	}
	if title != "" {
		preview.Title = &title
	}

	writeJSON(w, http.StatusOK, preview)
}

// fetchOpenGraphMeta retrieves og:image and og:title from the target page; failures yield empty values.
func fetchOpenGraphMeta(ctx context.Context, target string) (imageURL, title string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", ""
	}
	req.Header.Set("User-Agent", "OverpackedApp/1.0 (+route-preview)")
	req.Header.Set("Accept", "text/html")

	resp, err := routePreviewClient.Do(req)
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, routePreviewMaxBodyBytes))
	if err != nil {
		return "", ""
	}

	if match := ogImageRegexp.FindSubmatch(body); match != nil {
		imageURL = strings.TrimSpace(string(match[1]))
	}
	if match := ogTitleRegexp.FindSubmatch(body); match != nil {
		title = strings.TrimSpace(string(match[1]))
	}

	return imageURL, title
}

// Trip CRUD operations

func (h *TripsHandler) ListTrips(w http.ResponseWriter, r *http.Request) {
	trips, err := h.store.Trips.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": failedToGetErrMsg + " trips"})
		return
	}

	resp := make([]api.Trip, len(trips))
	for i, t := range trips {
		resp[i] = tripToAPI(&t)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TripsHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	var req api.TripCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": tripsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	trip := &domain.Trip{
		ID:       uuid.New(),
		Name:     req.Name,
		TripType: domain.TripType(req.TripType),
	}
	if req.Duration != nil {
		trip.Duration = req.Duration
	}
	if req.Notes != nil {
		trip.Notes = req.Notes
	}
	if req.TripKomootUrl != nil {
		trip.TripKomootURL = req.TripKomootUrl
	}
	if req.TripStravaUrl != nil {
		trip.TripStravaURL = req.TripStravaUrl
	}
	if req.TripWandererUrl != nil {
		trip.TripWandererURL = req.TripWandererUrl
	}
	if req.TotalDistanceKm != nil {
		distKm := float64(*req.TotalDistanceKm)
		trip.TotalDistanceKm = &distKm
	}

	if err := h.store.Trips.Create(r.Context(), trip); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create trip"})
		return
	}

	writeJSON(w, http.StatusCreated, tripToAPI(trip))
}

func (h *TripsHandler) GetTripById(w http.ResponseWriter, r *http.Request, tripId types.UUID) {
	trip, err := h.store.Trips.GetByID(r.Context(), uuid.UUID(tripId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrTripNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTrip})
		return
	}

	// Build nested response with persons, packs, and items
	resp, err := h.buildTripWithDetails(r.Context(), trip)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to build trip details"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TripsHandler) UpdateTrip(w http.ResponseWriter, r *http.Request, tripId types.UUID) {
	var req api.TripUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": tripsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	trip, err := h.store.Trips.GetByID(r.Context(), uuid.UUID(tripId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrTripNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTrip})
		return
	}

	if req.Name != nil {
		trip.Name = *req.Name
	}
	if req.TripType != nil {
		tripType := domain.TripType(*req.TripType)
		trip.TripType = tripType
	}
	if req.Duration != nil {
		trip.Duration = req.Duration
	}
	if req.Notes != nil {
		trip.Notes = req.Notes
	}
	if req.TripKomootUrl != nil {
		trip.TripKomootURL = req.TripKomootUrl
	}
	if req.TripStravaUrl != nil {
		trip.TripStravaURL = req.TripStravaUrl
	}
	if req.TripWandererUrl != nil {
		trip.TripWandererURL = req.TripWandererUrl
	}
	if req.TotalDistanceKm != nil {
		distKm := float64(*req.TotalDistanceKm)
		trip.TotalDistanceKm = &distKm
	}

	if err := h.store.Trips.Update(r.Context(), trip); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update trip"})
		return
	}

	writeJSON(w, http.StatusOK, tripToAPI(trip))
}

func (h *TripsHandler) DeleteTrip(w http.ResponseWriter, r *http.Request, tripId types.UUID) {
	if err := h.store.Trips.Delete(r.Context(), uuid.UUID(tripId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrTripNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete trip"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TripPersons operations

func (h *TripsHandler) AddTripPerson(w http.ResponseWriter, r *http.Request, tripId types.UUID) {
	var req api.TripPersonCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": tripsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	// Verify trip exists
	if _, err := h.store.Trips.GetByID(r.Context(), uuid.UUID(tripId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrTripNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTrip})
		return
	}

	// Verify person exists
	person, err := h.store.Persons.GetByID(r.Context(), uuid.UUID(req.PersonId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "person not found"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get person"})
		return
	}

	tripPerson := &domain.TripPerson{
		TripID:   uuid.UUID(tripId),
		PersonID: uuid.UUID(req.PersonId),
	}

	if err := h.store.Trips.AddPerson(r.Context(), tripPerson); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add person to trip"})
		return
	}

	apiPerson := personToAPI(person)
	resp := api.TripPersonWithDetails{
		PersonId: req.PersonId,
		Person:   apiPerson,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *TripsHandler) RemoveTripPerson(w http.ResponseWriter, r *http.Request, tripId types.UUID, personId types.UUID) {
	if err := h.store.Trips.RemovePerson(r.Context(), uuid.UUID(tripId), uuid.UUID(personId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrPersonNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove person from trip"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TripPersonPack operations

func (h *TripsHandler) AddTripPersonPack(w http.ResponseWriter, r *http.Request, tripId types.UUID, personId types.UUID) {
	var req api.TripPersonPackCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": tripsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	// Get trip_person_id
	tripPersonID, err := h.store.Trips.GetTripPersonID(r.Context(), uuid.UUID(tripId), uuid.UUID(personId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrPersonNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTripPerson})
		return
	}

	// Create new pack
	tripType := domain.TripType(req.TripType)
	pack := &domain.Pack{
		Name:     req.Name,
		TripType: &tripType,
		Notes:    req.Notes,
		PersonID: (*uuid.UUID)(&personId),
	}

	if err := h.store.Packs.Create(r.Context(), pack); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create pack"})
		return
	}

	// Link pack to trip person
	tripPersonPack := &domain.TripPersonPack{
		TripPersonID: tripPersonID,
		PackID:       pack.ID,
	}

	if err := h.store.Trips.AddPersonPack(r.Context(), tripPersonPack); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add pack to person"})
		return
	}

	apiPack := packToAPI(pack)
	resp := api.TripPersonPackWithDetails{
		Id:           types.UUID(tripPersonPack.ID),
		TripPersonId: types.UUID(tripPersonID),
		PackId:       types.UUID(pack.ID),
		Pack:         apiPack,
		CreatedAt:    tripPersonPack.CreatedAt,
		UpdatedAt:    tripPersonPack.UpdatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *TripsHandler) RemoveTripPersonPack(w http.ResponseWriter, r *http.Request, tripId types.UUID, personId types.UUID, packId types.UUID) {
	// Get trip_person_id
	tripPersonID, err := h.store.Trips.GetTripPersonID(r.Context(), uuid.UUID(tripId), uuid.UUID(personId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrPersonNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTripPerson})
		return
	}

	if err := h.store.Trips.RemovePersonPack(r.Context(), tripPersonID, uuid.UUID(packId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrPackNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove pack from person"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TripPersonItem operations

func (h *TripsHandler) AddTripPersonItem(w http.ResponseWriter, r *http.Request, tripId types.UUID, personId types.UUID) {
	var req api.TripPersonItemCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": tripsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	// Get trip_person_id
	tripPersonID, err := h.store.Trips.GetTripPersonID(r.Context(), uuid.UUID(tripId), uuid.UUID(personId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrPersonNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTripPerson})
		return
	}

	// Verify item exists
	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(req.ItemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrItemNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item"})
		return
	}

	tripPersonItem := &domain.TripPersonItem{
		TripPersonID: tripPersonID,
		ItemID:       uuid.UUID(req.ItemId),
		Quantity:     req.Quantity,
		CarryStatus:  domain.CarryStatus(req.CarryStatus),
	}
	if req.Notes != nil {
		tripPersonItem.Notes = req.Notes
	}

	if err := h.store.Trips.AddPersonItem(r.Context(), tripPersonItem); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add item to person"})
		return
	}

	apiItem := itemToAPI(item)
	resp := api.TripPersonItemWithDetails{
		Id:           types.UUID(tripPersonItem.ID),
		TripPersonId: types.UUID(tripPersonID),
		ItemId:       req.ItemId,
		Quantity:     req.Quantity,
		CarryStatus:  api.TripPersonItemWithDetailsCarryStatus(req.CarryStatus),
		Notes:        tripPersonItem.Notes,
		Item:         apiItem,
		CreatedAt:    tripPersonItem.CreatedAt,
		UpdatedAt:    tripPersonItem.UpdatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *TripsHandler) UpdateTripPersonItem(w http.ResponseWriter, r *http.Request, tripId types.UUID, personId types.UUID, itemId types.UUID) {
	var req api.TripPersonItemUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": tripsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	// Verify trip person exists
	_, err := h.store.Trips.GetTripPersonID(r.Context(), uuid.UUID(tripId), uuid.UUID(personId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrPersonNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTripPerson})
		return
	}

	tripPersonItem, err := h.store.Trips.GetPersonItemByID(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrItemNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get person item"})
		return
	}

	if req.Quantity != nil {
		tripPersonItem.Quantity = *req.Quantity
	}
	if req.CarryStatus != nil {
		tripPersonItem.CarryStatus = domain.CarryStatus(*req.CarryStatus)
	}
	if req.Notes != nil {
		tripPersonItem.Notes = req.Notes
	}

	if err := h.store.Trips.UpdatePersonItem(r.Context(), tripPersonItem); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update person item"})
		return
	}

	item, err := h.store.Items.GetByID(r.Context(), tripPersonItem.ItemID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item details"})
		return
	}

	apiItem := itemToAPI(item)
	resp := api.TripPersonItemWithDetails{
		Id:           types.UUID(tripPersonItem.ID),
		TripPersonId: types.UUID(tripPersonItem.TripPersonID),
		ItemId:       types.UUID(tripPersonItem.ItemID),
		Quantity:     tripPersonItem.Quantity,
		CarryStatus:  api.TripPersonItemWithDetailsCarryStatus(tripPersonItem.CarryStatus),
		Notes:        tripPersonItem.Notes,
		Item:         apiItem,
		CreatedAt:    tripPersonItem.CreatedAt,
		UpdatedAt:    tripPersonItem.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TripsHandler) RemoveTripPersonItem(w http.ResponseWriter, r *http.Request, tripId types.UUID, personId types.UUID, itemId types.UUID) {
	// Verify trip person exists
	_, err := h.store.Trips.GetTripPersonID(r.Context(), uuid.UUID(tripId), uuid.UUID(personId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrPersonNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTripPerson})
		return
	}

	if err := h.store.Trips.RemovePersonItem(r.Context(), uuid.UUID(itemId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrItemNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove item from person"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TripPersonPackItem operations (pack items are managed through the Pack resource)
// These endpoints manage pack_items for packs within a trip context

func (h *TripsHandler) AddTripPersonPackItem(w http.ResponseWriter, r *http.Request, tripId types.UUID, personId types.UUID, packId types.UUID) {
	var req api.PackItemCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": tripsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	// Verify trip person exists
	_, err := h.store.Trips.GetTripPersonID(r.Context(), uuid.UUID(tripId), uuid.UUID(personId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrPersonNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTripPerson})
		return
	}

	// Verify pack exists
	_, err = h.store.Packs.GetByID(r.Context(), uuid.UUID(packId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "pack not found"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get pack"})
		return
	}

	// Verify item exists
	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(req.ItemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item not found"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item"})
		return
	}

	packItem := &domain.PackItem{
		PackID:      uuid.UUID(packId),
		ItemID:      uuid.UUID(req.ItemId),
		Quantity:    int(req.Quantity),
		CarryStatus: domain.CarryStatus(req.CarryStatus),
		Notes:       req.Notes,
	}

	if err := h.store.Packs.AddItem(r.Context(), packItem); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add item to pack"})
		return
	}

	apiItem := itemToAPI(item)
	resp := api.PackItemWithDetails{
		Id:          types.UUID(packItem.ID),
		ItemId:      req.ItemId,
		Quantity:    req.Quantity,
		CarryStatus: api.PackItemWithDetailsCarryStatus(req.CarryStatus),
		Notes:       req.Notes,
		Item:        apiItem,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *TripsHandler) UpdateTripPersonPackItem(w http.ResponseWriter, r *http.Request, tripId types.UUID, personId types.UUID, packId types.UUID, itemId types.UUID) {
	var req api.PackItemUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": tripsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	// Verify trip person exists
	_, err := h.store.Trips.GetTripPersonID(r.Context(), uuid.UUID(tripId), uuid.UUID(personId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrPersonNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get trip person"})
		return
	}

	packItem, err := h.store.Packs.GetItemByID(r.Context(), uuid.UUID(packId), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrItemNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get pack item"})
		return
	}

	if req.Quantity != nil {
		packItem.Quantity = int(*req.Quantity)
	}
	if req.CarryStatus != nil {
		packItem.CarryStatus = domain.CarryStatus(*req.CarryStatus)
	}
	if req.Notes != nil {
		packItem.Notes = req.Notes
	}

	if err := h.store.Packs.UpdateItem(r.Context(), packItem); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update pack item"})
		return
	}

	item, err := h.store.Items.GetByID(r.Context(), packItem.ItemID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item details"})
		return
	}

	apiItem := itemToAPI(item)
	resp := api.PackItemWithDetails{
		Id:          types.UUID(packItem.ID),
		ItemId:      types.UUID(packItem.ItemID),
		Quantity:    float32(packItem.Quantity),
		CarryStatus: api.PackItemWithDetailsCarryStatus(packItem.CarryStatus),
		Notes:       packItem.Notes,
		Item:        apiItem,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TripsHandler) RemoveTripPersonPackItem(w http.ResponseWriter, r *http.Request, tripId types.UUID, personId types.UUID, packId types.UUID, itemId types.UUID) {
	// Verify trip person exists
	_, err := h.store.Trips.GetTripPersonID(r.Context(), uuid.UUID(tripId), uuid.UUID(personId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrPersonNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get trip person"})
		return
	}

	if err := h.store.Packs.RemoveItem(r.Context(), uuid.UUID(packId), uuid.UUID(itemId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrItemNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove item from pack"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func tripToAPI(t *domain.Trip) api.Trip {
	var duration *string
	if t.Duration != nil {
		duration = t.Duration
	}

	var distanceKm *float32
	if t.TotalDistanceKm != nil {
		flt := float32(*t.TotalDistanceKm)
		distanceKm = &flt
	}

	return api.Trip{
		Id:              types.UUID(t.ID),
		Name:            t.Name,
		TripType:        api.TripTripType(t.TripType),
		Duration:        duration,
		Notes:           t.Notes,
		TripKomootUrl:   t.TripKomootURL,
		TripStravaUrl:   t.TripStravaURL,
		TripWandererUrl: t.TripWandererURL,
		TotalDistanceKm: distanceKm,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}

// buildTripWithDetails constructs the complete nested trip response
func (h *TripsHandler) buildTripWithDetails(ctx context.Context, trip *domain.Trip) (*api.TripWithDetails, error) {
	personIDs, err := h.store.Trips.ListPersons(ctx, trip.ID)
	if err != nil {
		return nil, err
	}

	persons := make([]api.TripPersonDetailsNested, 0, len(personIDs))
	for _, personID := range personIDs {
		personDetails, err := h.buildPersonDetails(ctx, trip.ID, personID)
		if err != nil {
			continue
		}
		persons = append(persons, *personDetails)
	}

	var distanceKm *float32
	if trip.TotalDistanceKm != nil {
		flt := float32(*trip.TotalDistanceKm)
		distanceKm = &flt
	}

	return &api.TripWithDetails{
		Id:              types.UUID(trip.ID),
		Name:            trip.Name,
		TripType:        api.TripWithDetailsTripType(trip.TripType),
		Duration:        trip.Duration,
		Notes:           trip.Notes,
		TripKomootUrl:   trip.TripKomootURL,
		TripStravaUrl:   trip.TripStravaURL,
		TripWandererUrl: trip.TripWandererURL,
		TotalDistanceKm: distanceKm,
		Persons:         persons,
		CreatedAt:       trip.CreatedAt,
		UpdatedAt:       trip.UpdatedAt,
	}, nil
}

func (h *TripsHandler) buildPersonDetails(ctx context.Context, tripID, personID uuid.UUID) (*api.TripPersonDetailsNested, error) {
	person, err := h.store.Persons.GetByID(ctx, personID)
	if err != nil {
		return nil, err
	}

	tripPersonID, err := h.store.Trips.GetTripPersonID(ctx, tripID, personID)
	if err != nil {
		return nil, err
	}

	packs, err := h.buildPersonPacks(ctx, tripPersonID)
	if err != nil {
		return nil, err
	}

	items, err := h.buildPersonItems(ctx, tripPersonID)
	if err != nil {
		return nil, err
	}

	return &api.TripPersonDetailsNested{
		TripPersonId: types.UUID(tripPersonID),
		PersonId:     types.UUID(personID),
		Person:       personToAPI(person),
		Packs:        packs,
		Items:        items,
	}, nil
}

func (h *TripsHandler) buildPersonPacks(ctx context.Context, tripPersonID uuid.UUID) ([]api.TripPersonPackDetailsNested, error) {
	packIDs, err := h.store.Trips.ListPersonPacks(ctx, tripPersonID)
	if err != nil {
		return nil, err
	}

	packs := make([]api.TripPersonPackDetailsNested, 0, len(packIDs))
	for _, packID := range packIDs {
		packDetails, err := h.buildPackDetails(ctx, tripPersonID, packID)
		if err != nil {
			continue
		}
		packs = append(packs, *packDetails)
	}

	return packs, nil
}

func (h *TripsHandler) buildPackDetails(ctx context.Context, tripPersonID, packID uuid.UUID) (*api.TripPersonPackDetailsNested, error) {
	pack, err := h.store.Packs.GetByID(ctx, packID)
	if err != nil {
		return nil, err
	}

	items, err := h.buildPackItems(ctx, packID)
	if err != nil {
		return nil, err
	}

	return &api.TripPersonPackDetailsNested{
		Id:           types.UUID(uuid.New()),
		TripPersonId: types.UUID(tripPersonID),
		PackId:       types.UUID(packID),
		Pack:         packToAPI(pack),
		Items:        items,
		CreatedAt:    pack.CreatedAt,
		UpdatedAt:    pack.UpdatedAt,
	}, nil
}

func (h *TripsHandler) buildPackItems(ctx context.Context, packID uuid.UUID) ([]api.PackItemWithDetails, error) {
	packItems, err := h.store.Packs.ListItems(ctx, packID)
	if err != nil {
		return nil, err
	}

	items := make([]api.PackItemWithDetails, 0, len(packItems))
	for _, pi := range packItems {
		item, err := h.store.Items.GetByID(ctx, pi.ItemID)
		if err != nil {
			continue
		}
		items = append(items, api.PackItemWithDetails{
			Id:          types.UUID(pi.ID),
			ItemId:      types.UUID(pi.ItemID),
			Quantity:    float32(pi.Quantity),
			CarryStatus: api.PackItemWithDetailsCarryStatus(pi.CarryStatus),
			Notes:       pi.Notes,
			Item:        itemToAPI(item),
		})
	}

	return items, nil
}

func (h *TripsHandler) buildPersonItems(ctx context.Context, tripPersonID uuid.UUID) ([]api.TripPersonItemWithDetails, error) {
	personItems, err := h.store.Trips.ListPersonItems(ctx, tripPersonID)
	if err != nil {
		return nil, err
	}

	items := make([]api.TripPersonItemWithDetails, 0, len(personItems))
	for _, tpi := range personItems {
		item, err := h.store.Items.GetByID(ctx, tpi.ItemID)
		if err != nil {
			continue
		}
		items = append(items, api.TripPersonItemWithDetails{
			Id:           types.UUID(tpi.ID),
			TripPersonId: types.UUID(tpi.TripPersonID),
			ItemId:       types.UUID(tpi.ItemID),
			Quantity:     tpi.Quantity,
			CarryStatus:  api.TripPersonItemWithDetailsCarryStatus(tpi.CarryStatus),
			Notes:        tpi.Notes,
			Item:         itemToAPI(item),
			CreatedAt:    tpi.CreatedAt,
			UpdatedAt:    tpi.UpdatedAt,
		})
	}

	return items, nil
}
