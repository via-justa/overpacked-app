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

type ManufacturersHandler struct {
	store *store.Store
}

const manufacturersErrNotFound = "manufacturer not found"

func NewManufacturersHandler(st *store.Store) *ManufacturersHandler {
	return &ManufacturersHandler{store: st}
}

func (h *ManufacturersHandler) ListManufacturers(w http.ResponseWriter, r *http.Request) {
	manufacturers, err := h.store.Manufacturers.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list manufacturers")
		return
	}

	resp := make([]api.Manufacturer, len(manufacturers))
	for i, m := range manufacturers {
		resp[i] = manufacturerToAPI(&m)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *ManufacturersHandler) CreateManufacturer(w http.ResponseWriter, r *http.Request) {
	var req api.ManufacturerCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	manufacturer := &domain.Manufacturer{
		ID:   uuid.New(),
		Name: req.Name,
	}
	if req.Website != nil {
		manufacturer.Website = req.Website
	}

	if err := h.store.Manufacturers.Create(r.Context(), manufacturer); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create manufacturer")
		return
	}

	writeJSON(w, http.StatusCreated, manufacturerToAPI(manufacturer))
}

func (h *ManufacturersHandler) GetManufacturer(w http.ResponseWriter, r *http.Request, manufacturerId types.UUID) {
	manufacturer, err := h.store.Manufacturers.GetByID(r.Context(), uuid.UUID(manufacturerId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, manufacturersErrNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get manufacturer")
		return
	}

	writeJSON(w, http.StatusOK, manufacturerToAPI(manufacturer))
}

func (h *ManufacturersHandler) UpdateManufacturer(w http.ResponseWriter, r *http.Request, manufacturerId types.UUID) {
	var req api.ManufacturerUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	manufacturer, err := h.store.Manufacturers.GetByID(r.Context(), uuid.UUID(manufacturerId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, manufacturersErrNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get manufacturer")
		return
	}

	if req.Name != nil {
		manufacturer.Name = *req.Name
	}
	if req.Website != nil {
		manufacturer.Website = req.Website
	}

	if err := h.store.Manufacturers.Update(r.Context(), manufacturer); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update manufacturer")
		return
	}

	writeJSON(w, http.StatusOK, manufacturerToAPI(manufacturer))
}

func (h *ManufacturersHandler) DeleteManufacturer(w http.ResponseWriter, r *http.Request, manufacturerId types.UUID) {
	err := h.store.Manufacturers.Delete(r.Context(), uuid.UUID(manufacturerId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, manufacturersErrNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete manufacturer")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func manufacturerToAPI(m *domain.Manufacturer) api.Manufacturer {
	return api.Manufacturer{
		Id:        types.UUID(m.ID),
		Name:      m.Name,
		Website:   m.Website,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
