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

type SetsHandler struct {
	store *store.Store
}

const (
	setsErrInvalidRequestBody = "invalid request body"
	setsErrSetNotFound        = "set not found"
	setsErrFailedToGetSet     = "failed to get set"
	setsErrSetItemNotFound    = "set item not found"
)

func NewSetsHandler(st *store.Store) *SetsHandler {
	return &SetsHandler{store: st}
}

func (h *SetsHandler) ListSets(w http.ResponseWriter, r *http.Request) {
	sets, err := h.store.Sets.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sets")
		return
	}

	resp := make([]api.ItemSet, len(sets))
	for i, s := range sets {
		resp[i] = setToAPI(&s)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *SetsHandler) CreateSet(w http.ResponseWriter, r *http.Request) {
	var req api.ItemSetCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, setsErrInvalidRequestBody)
		return
	}
	defer r.Body.Close()

	set := &domain.ItemSet{ID: uuid.New(), Name: req.Name, IsActive: true}
	if req.SetCategory == "" {
		writeError(w, http.StatusBadRequest, "set_category is required")
		return
	}

	if _, err := h.store.ItemTypes.GetByID(r.Context(), req.SetCategory); errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusBadRequest, "invalid set_category")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to validate set category")
		return
	}

	set.SetCategory = req.SetCategory

	if err := h.store.Sets.Create(r.Context(), set); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create set")
		return
	}

	writeJSON(w, http.StatusCreated, setToAPI(set))
}

func (h *SetsHandler) GetSet(w http.ResponseWriter, r *http.Request, setId types.UUID) {
	set, err := h.store.Sets.GetByID(r.Context(), uuid.UUID(setId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, setsErrSetNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, setsErrFailedToGetSet)
		return
	}

	writeJSON(w, http.StatusOK, setToAPI(set))
}

func (h *SetsHandler) UpdateSet(w http.ResponseWriter, r *http.Request, setId types.UUID) {
	var req api.ItemSetUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, setsErrInvalidRequestBody)
		return
	}
	defer r.Body.Close()

	set, err := h.store.Sets.GetByID(r.Context(), uuid.UUID(setId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, setsErrSetNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, setsErrFailedToGetSet)
		return
	}

	if req.Name != nil {
		set.Name = *req.Name
	}

	if req.SetCategory == "" {
		writeError(w, http.StatusBadRequest, "set_category is required")
		return
	}

	if _, err := h.store.ItemTypes.GetByID(r.Context(), req.SetCategory); errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusBadRequest, "invalid set_category")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to validate set category")
		return
	}

	set.SetCategory = req.SetCategory

	if err := h.store.Sets.Update(r.Context(), set); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update set")
		return
	}

	writeJSON(w, http.StatusOK, setToAPI(set))
}

func (h *SetsHandler) DeleteSet(w http.ResponseWriter, r *http.Request, setId types.UUID) {
	err := h.store.Sets.Delete(r.Context(), uuid.UUID(setId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, setsErrSetNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete set")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SetsHandler) ListSetItems(w http.ResponseWriter, r *http.Request, setId types.UUID) {
	if _, err := h.store.Sets.GetByID(r.Context(), uuid.UUID(setId)); errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, setsErrSetNotFound)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, setsErrFailedToGetSet)
		return
	}

	setItems, err := h.store.Sets.ListItems(r.Context(), uuid.UUID(setId))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list set items")
		return
	}

	resp := make([]api.SetItemWithDetails, 0, len(setItems))
	for _, si := range setItems {
		item, getErr := h.store.Items.GetByID(r.Context(), si.ItemID)
		if getErr != nil {
			writeError(w, http.StatusInternalServerError, "failed to load set item details")
			return
		}
		resp = append(resp, setItemToAPI(&si, item))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *SetsHandler) AddSetItem(w http.ResponseWriter, r *http.Request, setId types.UUID) {
	var req api.SetItemCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, setsErrInvalidRequestBody)
		return
	}
	defer r.Body.Close()

	if _, err := h.store.Sets.GetByID(r.Context(), uuid.UUID(setId)); errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, setsErrSetNotFound)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, setsErrFailedToGetSet)
		return
	}

	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(req.ItemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get item")
		return
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	setItem := &domain.SetItem{
		SetID:     uuid.UUID(setId),
		ItemID:    uuid.UUID(req.ItemId),
		Quantity:  int(req.Quantity),
		Notes:     req.Notes,
		SortOrder: sortOrder,
	}

	if err := h.store.Sets.AddItem(r.Context(), setItem); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add item to set")
		return
	}

	writeJSON(w, http.StatusCreated, setItemToAPI(setItem, item))
}

func (h *SetsHandler) UpdateSetItem(w http.ResponseWriter, r *http.Request, setId types.UUID, itemId types.UUID) {
	var req api.SetItemUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, setsErrInvalidRequestBody)
		return
	}
	defer r.Body.Close()

	if _, err := h.store.Sets.GetByID(r.Context(), uuid.UUID(setId)); errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, setsErrSetNotFound)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, setsErrFailedToGetSet)
		return
	}

	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get item")
		return
	}

	setItems, err := h.store.Sets.ListItems(r.Context(), uuid.UUID(setId))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load set items")
		return
	}

	var current *domain.SetItem
	for i := range setItems {
		if setItems[i].ItemID == uuid.UUID(itemId) {
			current = &setItems[i]
			break
		}
	}
	if current == nil {
		writeError(w, http.StatusNotFound, setsErrSetItemNotFound)
		return
	}

	if req.Quantity != nil {
		current.Quantity = int(*req.Quantity)
	}
	if req.Notes != nil {
		current.Notes = req.Notes
	}
	if req.SortOrder != nil {
		current.SortOrder = *req.SortOrder
	}

	if err := h.store.Sets.UpdateItem(r.Context(), current); errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, setsErrSetItemNotFound)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update set item")
		return
	}

	writeJSON(w, http.StatusOK, setItemToAPI(current, item))
}

func (h *SetsHandler) RemoveSetItem(w http.ResponseWriter, r *http.Request, setId types.UUID, itemId types.UUID) {
	if _, err := h.store.Sets.GetByID(r.Context(), uuid.UUID(setId)); errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, setsErrSetNotFound)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, setsErrFailedToGetSet)
		return
	}

	err := h.store.Sets.RemoveItem(r.Context(), uuid.UUID(setId), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, setsErrSetItemNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove set item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func setToAPI(s *domain.ItemSet) api.ItemSet {
	return api.ItemSet{Id: types.UUID(s.ID), Name: s.Name, SetCategory: s.SetCategory, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt}
}

func setItemToAPI(si *domain.SetItem, item *domain.Item) api.SetItemWithDetails {
	qty := float32(si.Quantity)
	sortOrder := si.SortOrder
	return api.SetItemWithDetails{
		ItemId:    types.UUID(si.ItemID),
		Quantity:  qty,
		Notes:     si.Notes,
		SortOrder: &sortOrder,
		Item:      itemToAPI(item),
	}
}
