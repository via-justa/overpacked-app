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

// PackingListsHandler handles packing list-related HTTP requests
type PackingListsHandler struct {
	packingLists *store.PackingListStore
	labels       *store.LabelStore
}

// NewPackingListsHandler creates a new packing lists handler
func NewPackingListsHandler(packingLists *store.PackingListStore, labels *store.LabelStore) *PackingListsHandler {
	return &PackingListsHandler{
		packingLists: packingLists,
		labels:       labels,
	}
}

const (
	packingListErrNotFound          = "packing list not found"
	packingListErrInvalidBody       = "invalid request body"
	packingListErrFailedCreate      = "failed to create packing list"
	packingListErrFailedGet         = "failed to get packing list"
	packingListErrFailedUpdate      = "failed to update packing list"
	packingListErrFailedDelete      = "failed to delete packing list"
	packingListErrFailedList        = "failed to list packing lists"
	packingListErrFailedListLabels  = "failed to list packing list labels"
	packingListErrFailedGetLabel    = "failed to get label"
	packingListErrFailedAddLabel    = "failed to add label to packing list"
	packingListErrFailedRemoveLabel = "failed to remove label from packing list"
	packingListErrLabelNotFound     = "label not found"
)

// ListPackingLists handles GET /api/v1/packing-lists
func (h *PackingListsHandler) ListPackingLists(w http.ResponseWriter, r *http.Request) {
	lists, err := h.packingLists.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, packingListErrFailedList)
		return
	}

	resp := make([]api.PackingList, len(lists))
	for i, pl := range lists {
		resp[i] = packingListToAPI(&pl)
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreatePackingList handles POST /api/v1/packing-lists
func (h *PackingListsHandler) CreatePackingList(w http.ResponseWriter, r *http.Request) {
	var req api.PackingListCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, packingListErrInvalidBody)
		return
	}
	defer r.Body.Close()

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	pl, err := h.packingLists.Create(r.Context(), req.Name, description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, packingListErrFailedCreate)
		return
	}

	writeJSON(w, http.StatusCreated, packingListToAPI(pl))
}

// GetPackingListById handles GET /api/v1/packing-lists/{listId}
func (h *PackingListsHandler) GetPackingListById(w http.ResponseWriter, r *http.Request, listId types.UUID) {
	pl, err := h.packingLists.GetByID(r.Context(), uuid.UUID(listId))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, packingListErrNotFound)
			return
		}
		writeError(w, http.StatusInternalServerError, packingListErrFailedGet)
		return
	}

	writeJSON(w, http.StatusOK, packingListToAPI(pl))
}

// UpdatePackingList handles PATCH /api/v1/packing-lists/{listId}
func (h *PackingListsHandler) UpdatePackingList(w http.ResponseWriter, r *http.Request, listId types.UUID) {
	var req api.PackingListUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, packingListErrInvalidBody)
		return
	}
	defer r.Body.Close()

	pl, err := h.packingLists.Update(r.Context(), uuid.UUID(listId), req.Name, req.Description)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, packingListErrNotFound)
			return
		}
		writeError(w, http.StatusInternalServerError, packingListErrFailedUpdate)
		return
	}

	writeJSON(w, http.StatusOK, packingListToAPI(pl))
}

// DeletePackingList handles DELETE /api/v1/packing-lists/{listId}
func (h *PackingListsHandler) DeletePackingList(w http.ResponseWriter, r *http.Request, listId types.UUID) {
	if err := h.packingLists.Delete(r.Context(), uuid.UUID(listId)); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, packingListErrNotFound)
			return
		}
		writeError(w, http.StatusInternalServerError, packingListErrFailedDelete)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListPackingListLabels handles GET /api/v1/packing-lists/{listId}/labels
func (h *PackingListsHandler) ListPackingListLabels(w http.ResponseWriter, r *http.Request, listId types.UUID) {
	labels, err := h.packingLists.ListLabels(r.Context(), uuid.UUID(listId))
	if err != nil {
		writeError(w, http.StatusInternalServerError, packingListErrFailedListLabels)
		return
	}

	resp := make([]api.Label, len(labels))
	for i, l := range labels {
		resp[i] = labelToAPI(&l)
	}

	writeJSON(w, http.StatusOK, resp)
}

// AddPackingListLabel handles POST /api/v1/packing-lists/{listId}/labels
func (h *PackingListsHandler) AddPackingListLabel(w http.ResponseWriter, r *http.Request, listId types.UUID) {
	var req api.PackingListLabelAdd
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, packingListErrInvalidBody)
		return
	}
	defer r.Body.Close()

	// Verify label exists
	label, err := h.labels.GetByID(r.Context(), uuid.UUID(req.LabelId))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, packingListErrLabelNotFound)
			return
		}
		writeError(w, http.StatusInternalServerError, packingListErrFailedGetLabel)
		return
	}

	// Add label to packing list
	if err := h.packingLists.AddLabel(r.Context(), uuid.UUID(listId), uuid.UUID(req.LabelId)); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, packingListErrNotFound)
			return
		}
		writeError(w, http.StatusInternalServerError, packingListErrFailedAddLabel)
		return
	}

	writeJSON(w, http.StatusCreated, labelToAPI(label))
}

// RemovePackingListLabel handles DELETE /api/v1/packing-lists/{listId}/labels/{labelId}
func (h *PackingListsHandler) RemovePackingListLabel(w http.ResponseWriter, r *http.Request, listId types.UUID, labelId types.UUID) {
	if err := h.packingLists.RemoveLabel(r.Context(), uuid.UUID(listId), uuid.UUID(labelId)); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, packingListErrNotFound)
			return
		}
		writeError(w, http.StatusInternalServerError, packingListErrFailedRemoveLabel)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// packingListToAPI converts domain.PackingList to api.PackingList
func packingListToAPI(pl *domain.PackingList) api.PackingList {
	return api.PackingList{
		Id:          types.UUID(pl.ID),
		Name:        pl.Name,
		Description: pl.Description,
		CreatedAt:   pl.CreatedAt,
		UpdatedAt:   pl.UpdatedAt,
	}
}
