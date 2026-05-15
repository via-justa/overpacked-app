package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

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
	tripsErrInvalidRequestBody = "invalid request body"
	tripsErrTripNotFound       = "trip not found"
	tripsErrFailedToGetTrip    = "failed to get trip"
	tripsErrPackNotFound       = "pack not found in trip"
	tripsErrItemNotFound       = "item not found in trip"
	tripsErrSetNotFound        = "set not found in trip"
	tripsErrPersonNotFound     = "person not found in trip"
)

func NewTripsHandler(st *store.Store) *TripsHandler {
	return &TripsHandler{store: st}
}

// Trip CRUD operations

func (h *TripsHandler) ListTrips(w http.ResponseWriter, r *http.Request) {
	trips, err := h.store.Trips.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list trips"})
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

	writeJSON(w, http.StatusOK, tripToAPI(trip))
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

// TripPacks operations

func (h *TripsHandler) ListTripPacks(w http.ResponseWriter, r *http.Request, tripId types.UUID) {
	// Verify trip exists
	if _, err := h.store.Trips.GetByID(r.Context(), uuid.UUID(tripId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrTripNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTrip})
		return
	}

	packIDs, err := h.store.Trips.ListPacks(r.Context(), uuid.UUID(tripId))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list trip packs"})
		return
	}

	resp := make([]api.TripPackWithDetails, 0, len(packIDs))
	for _, packID := range packIDs {
		pack, err := h.store.Packs.GetByID(r.Context(), packID)
		if err != nil {
			continue // Skip packs that no longer exist
		}
		apiPack := packToAPI(pack)
		resp = append(resp, api.TripPackWithDetails{
			PackId: types.UUID(packID),
			Pack:   apiPack,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TripsHandler) AddTripPack(w http.ResponseWriter, r *http.Request, tripId types.UUID) {
	var req api.TripPackCreate
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

	// Verify pack exists
	pack, err := h.store.Packs.GetByID(r.Context(), uuid.UUID(req.PackId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "pack not found"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get pack"})
		return
	}

	tripPack := &domain.TripPack{
		TripID: uuid.UUID(tripId),
		PackID: uuid.UUID(req.PackId),
	}

	if err := h.store.Trips.AddPack(r.Context(), tripPack); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add pack to trip"})
		return
	}

	apiPack := packToAPI(pack)
	resp := api.TripPackWithDetails{
		PackId: req.PackId,
		Pack:   apiPack,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *TripsHandler) RemoveTripPack(w http.ResponseWriter, r *http.Request, tripId types.UUID, packId types.UUID) {
	if err := h.store.Trips.RemovePack(r.Context(), uuid.UUID(tripId), uuid.UUID(packId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrPackNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove pack from trip"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TripItems operations

func (h *TripsHandler) ListTripItems(w http.ResponseWriter, r *http.Request, tripId types.UUID) {
	// Verify trip exists
	if _, err := h.store.Trips.GetByID(r.Context(), uuid.UUID(tripId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrTripNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTrip})
		return
	}

	tripItems, err := h.store.Trips.ListItems(r.Context(), uuid.UUID(tripId))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list trip items"})
		return
	}

	resp := make([]api.TripItemWithDetails, 0, len(tripItems))
	for _, ti := range tripItems {
		item, err := h.store.Items.GetByID(r.Context(), ti.ItemID)
		if err != nil {
			continue // Skip items that no longer exist
		}
		apiItem := itemToAPI(item)
		resp = append(resp, api.TripItemWithDetails{
			ItemId:      types.UUID(ti.ItemID),
			Quantity:    float32(ti.Quantity),
			CarryStatus: api.TripItemWithDetailsCarryStatus(ti.CarryStatus),
			Notes:       ti.Notes,
			Item:        apiItem,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TripsHandler) AddTripItem(w http.ResponseWriter, r *http.Request, tripId types.UUID) {
	var req api.TripItemCreate
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

	// Verify item exists
	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(req.ItemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item not found"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item"})
		return
	}

	tripItem := &domain.TripItem{
		TripID:      uuid.UUID(tripId),
		ItemID:      uuid.UUID(req.ItemId),
		Quantity:    int(req.Quantity),
		CarryStatus: domain.CarryStatus(req.CarryStatus),
	}
	if req.Notes != nil {
		tripItem.Notes = req.Notes
	}

	if err := h.store.Trips.AddItem(r.Context(), tripItem); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add item to trip"})
		return
	}

	apiItem := itemToAPI(item)
	resp := api.TripItemWithDetails{
		ItemId:      req.ItemId,
		Quantity:    req.Quantity,
		CarryStatus: api.TripItemWithDetailsCarryStatus(req.CarryStatus),
		Notes:       tripItem.Notes,
		Item:        apiItem,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *TripsHandler) UpdateTripItem(w http.ResponseWriter, r *http.Request, tripId types.UUID, itemId types.UUID) {
	var req api.TripItemUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": tripsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	tripItem, err := h.store.Trips.GetItemByID(r.Context(), uuid.UUID(tripId), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrItemNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get trip item"})
		return
	}

	if req.Quantity != nil {
		tripItem.Quantity = int(*req.Quantity)
	}
	if req.CarryStatus != nil {
		tripItem.CarryStatus = domain.CarryStatus(*req.CarryStatus)
	}
	if req.Notes != nil {
		tripItem.Notes = req.Notes
	}

	if err := h.store.Trips.UpdateItem(r.Context(), tripItem); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update trip item"})
		return
	}

	item, err := h.store.Items.GetByID(r.Context(), tripItem.ItemID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item details"})
		return
	}

	apiItem := itemToAPI(item)
	resp := api.TripItemWithDetails{
		ItemId:      types.UUID(tripItem.ItemID),
		Quantity:    float32(tripItem.Quantity),
		CarryStatus: api.TripItemWithDetailsCarryStatus(tripItem.CarryStatus),
		Notes:       tripItem.Notes,
		Item:        apiItem,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TripsHandler) RemoveTripItem(w http.ResponseWriter, r *http.Request, tripId types.UUID, itemId types.UUID) {
	if err := h.store.Trips.RemoveItem(r.Context(), uuid.UUID(tripId), uuid.UUID(itemId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrItemNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove item from trip"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TripSets operations

func (h *TripsHandler) ListTripSets(w http.ResponseWriter, r *http.Request, tripId types.UUID) {
	// Verify trip exists
	if _, err := h.store.Trips.GetByID(r.Context(), uuid.UUID(tripId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrTripNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTrip})
		return
	}

	setIDs, err := h.store.Trips.ListSets(r.Context(), uuid.UUID(tripId))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list trip sets"})
		return
	}

	resp := make([]api.TripSetWithDetails, 0, len(setIDs))
	for _, setID := range setIDs {
		set, err := h.store.Sets.GetByID(r.Context(), setID)
		if err != nil {
			continue // Skip sets that no longer exist
		}
		apiSet := setToAPI(set)
		resp = append(resp, api.TripSetWithDetails{
			SetId: types.UUID(setID),
			Set:   apiSet,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TripsHandler) AddTripSet(w http.ResponseWriter, r *http.Request, tripId types.UUID) {
	var req api.TripSetCreate
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

	// Verify set exists
	set, err := h.store.Sets.GetByID(r.Context(), uuid.UUID(req.SetId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "set not found"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get set"})
		return
	}

	tripSet := &domain.TripSet{
		TripID: uuid.UUID(tripId),
		SetID:  uuid.UUID(req.SetId),
	}

	if err := h.store.Trips.AddSet(r.Context(), tripSet); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add set to trip"})
		return
	}

	apiSet := setToAPI(set)
	resp := api.TripSetWithDetails{
		SetId: req.SetId,
		Set:   apiSet,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *TripsHandler) RemoveTripSet(w http.ResponseWriter, r *http.Request, tripId types.UUID, setId types.UUID) {
	if err := h.store.Trips.RemoveSet(r.Context(), uuid.UUID(tripId), uuid.UUID(setId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrSetNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove set from trip"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TripPersons operations

func (h *TripsHandler) ListTripPersons(w http.ResponseWriter, r *http.Request, tripId types.UUID) {
	// Verify trip exists
	if _, err := h.store.Trips.GetByID(r.Context(), uuid.UUID(tripId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": tripsErrTripNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": tripsErrFailedToGetTrip})
		return
	}

	personIDs, err := h.store.Trips.ListPersons(r.Context(), uuid.UUID(tripId))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list trip persons"})
		return
	}

	resp := make([]api.TripPersonWithDetails, 0, len(personIDs))
	for _, personID := range personIDs {
		person, err := h.store.Persons.GetByID(r.Context(), personID)
		if err != nil {
			continue // Skip persons that no longer exist
		}
		apiPerson := personToAPI(person)
		resp = append(resp, api.TripPersonWithDetails{
			PersonId: types.UUID(personID),
			Person:   apiPerson,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

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
