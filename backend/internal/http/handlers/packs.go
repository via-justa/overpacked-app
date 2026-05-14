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

type PacksHandler struct {
	store *store.Store
}

const (
	packsErrInvalidRequestBody = "invalid request body"
	packsErrPackNotFound       = "pack not found"
	packsErrFailedToGetPack    = "failed to get pack"
	packsErrPackItemNotFound   = "pack item not found"
)

func NewPacksHandler(st *store.Store) *PacksHandler {
	return &PacksHandler{store: st}
}

func (h *PacksHandler) ListPacks(w http.ResponseWriter, r *http.Request) {
	packs, err := h.store.Packs.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list packs"})
		return
	}

	resp := make([]api.Pack, len(packs))
	for i, p := range packs {
		resp[i] = packToAPI(&p)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *PacksHandler) CreatePack(w http.ResponseWriter, r *http.Request) {
	var req api.PackCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": packsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	trip := domain.TripType(req.TripType)
	pack := &domain.Pack{
		ID:         uuid.New(),
		Name:       req.Name,
		TripType:   &trip,
		IsTemplate: false,
	}
	if req.PersonId != nil {
		pid := uuid.UUID(*req.PersonId)
		pack.PersonID = &pid
	}

	if err := h.store.Packs.Create(r.Context(), pack); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create pack"})
		return
	}

	writeJSON(w, http.StatusCreated, packToAPI(pack))
}

func (h *PacksHandler) GetPack(w http.ResponseWriter, r *http.Request, packId types.UUID) {
	pack, err := h.store.Packs.GetByID(r.Context(), uuid.UUID(packId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": packsErrFailedToGetPack})
		return
	}

	writeJSON(w, http.StatusOK, packToAPI(pack))
}

func (h *PacksHandler) UpdatePack(w http.ResponseWriter, r *http.Request, packId types.UUID) {
	var req api.PackUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": packsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	pack, err := h.store.Packs.GetByID(r.Context(), uuid.UUID(packId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": packsErrFailedToGetPack})
		return
	}

	if req.Name != nil {
		pack.Name = *req.Name
	}
	if req.TripType != nil {
		trip := domain.TripType(*req.TripType)
		pack.TripType = &trip
	}
	if req.PersonId != nil {
		pid := uuid.UUID(*req.PersonId)
		pack.PersonID = &pid
	}

	if err := h.store.Packs.Update(r.Context(), pack); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackNotFound})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update pack"})
		return
	}

	writeJSON(w, http.StatusOK, packToAPI(pack))
}

func (h *PacksHandler) DeletePack(w http.ResponseWriter, r *http.Request, packId types.UUID) {
	err := h.store.Packs.Delete(r.Context(), uuid.UUID(packId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete pack"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PacksHandler) ListPackItems(w http.ResponseWriter, r *http.Request, packId types.UUID) {
	if _, err := h.store.Packs.GetByID(r.Context(), uuid.UUID(packId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": packsErrFailedToGetPack})
		return
	}

	packItems, err := h.store.Packs.ListItems(r.Context(), uuid.UUID(packId))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list pack items"})
		return
	}

	resp := make([]api.PackItemWithDetails, 0, len(packItems))
	for _, pi := range packItems {
		item, getErr := h.store.Items.GetByID(r.Context(), pi.ItemID)
		if getErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load pack item details"})
			return
		}
		resp = append(resp, packItemToAPI(&pi, item))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *PacksHandler) AddPackItem(w http.ResponseWriter, r *http.Request, packId types.UUID) {
	var req api.PackItemCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": packsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	if _, err := h.store.Packs.GetByID(r.Context(), uuid.UUID(packId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": packsErrFailedToGetPack})
		return
	}

	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(req.ItemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item"})
		return
	}

	packItem := &domain.PackItem{
		PackID:      uuid.UUID(packId),
		ItemID:      uuid.UUID(req.ItemId),
		Quantity:    int(req.Quantity),
		CarryStatus: domain.CarryStatus(req.CarryStatus),
	}

	if err := h.store.Packs.AddItem(r.Context(), packItem); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add item to pack"})
		return
	}

	writeJSON(w, http.StatusCreated, packItemToAPI(packItem, item))
}

func (h *PacksHandler) UpdatePackItem(w http.ResponseWriter, r *http.Request, packId types.UUID, itemId types.UUID) {
	var req api.PackItemUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": packsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	if _, err := h.store.Packs.GetByID(r.Context(), uuid.UUID(packId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": packsErrFailedToGetPack})
		return
	}

	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item"})
		return
	}

	packItems, err := h.store.Packs.ListItems(r.Context(), uuid.UUID(packId))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load pack items"})
		return
	}

	var current *domain.PackItem
	for i := range packItems {
		if packItems[i].ItemID == uuid.UUID(itemId) {
			current = &packItems[i]
			break
		}
	}
	if current == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackItemNotFound})
		return
	}

	if req.Quantity != nil {
		current.Quantity = int(*req.Quantity)
	}
	if req.CarryStatus != nil {
		current.CarryStatus = domain.CarryStatus(*req.CarryStatus)
	}

	if err := h.store.Packs.UpdateItem(r.Context(), current); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackItemNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update pack item"})
		return
	}

	writeJSON(w, http.StatusOK, packItemToAPI(current, item))
}

func (h *PacksHandler) RemovePackItem(w http.ResponseWriter, r *http.Request, packId types.UUID, itemId types.UUID) {
	if _, err := h.store.Packs.GetByID(r.Context(), uuid.UUID(packId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": packsErrFailedToGetPack})
		return
	}

	err := h.store.Packs.RemoveItem(r.Context(), uuid.UUID(packId), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackItemNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove pack item"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PacksHandler) ListPackSets(w http.ResponseWriter, r *http.Request, packId types.UUID) {
	if _, err := h.store.Packs.GetByID(r.Context(), uuid.UUID(packId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": packsErrFailedToGetPack})
		return
	}

	packSets, err := h.store.Packs.ListSets(r.Context(), uuid.UUID(packId))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list pack sets"})
		return
	}

	resp := make([]api.ItemSet, 0, len(packSets))
	for _, ps := range packSets {
		set, getErr := h.store.Sets.GetByID(r.Context(), ps.SetID)
		if getErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load pack set details"})
			return
		}
		resp = append(resp, setToAPI(set))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *PacksHandler) AddPackSet(w http.ResponseWriter, r *http.Request, packId types.UUID) {
	var req api.PackSetCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": packsErrInvalidRequestBody})
		return
	}
	defer r.Body.Close()

	if _, err := h.store.Packs.GetByID(r.Context(), uuid.UUID(packId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": packsErrFailedToGetPack})
		return
	}

	set, err := h.store.Sets.GetByID(r.Context(), uuid.UUID(req.SetId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "set not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get set"})
		return
	}

	packSet := &domain.PackSet{PackID: uuid.UUID(packId), SetID: uuid.UUID(req.SetId)}
	if err := h.store.Packs.AssignSet(r.Context(), packSet); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add set to pack"})
		return
	}

	writeJSON(w, http.StatusCreated, setToAPI(set))
}

func (h *PacksHandler) RemovePackSet(w http.ResponseWriter, r *http.Request, packId types.UUID, setId types.UUID) {
	if _, err := h.store.Packs.GetByID(r.Context(), uuid.UUID(packId)); errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": packsErrPackNotFound})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": packsErrFailedToGetPack})
		return
	}

	err := h.store.Packs.RemoveSet(r.Context(), uuid.UUID(packId), uuid.UUID(setId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "pack set not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove set from pack"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func packToAPI(p *domain.Pack) api.Pack {
	tripType := api.PackTripTypeDayHike
	if p.TripType != nil {
		tripType = api.PackTripType(*p.TripType)
	}

	resp := api.Pack{Id: types.UUID(p.ID), Name: p.Name, TripType: tripType, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
	if p.PersonID != nil {
		pid := types.UUID(*p.PersonID)
		resp.PersonId = &pid
	}
	return resp
}

func packItemToAPI(pi *domain.PackItem, item *domain.Item) api.PackItemWithDetails {
	qty := float32(pi.Quantity)
	return api.PackItemWithDetails{ItemId: types.UUID(pi.ItemID), Quantity: qty, CarryStatus: api.PackItemWithDetailsCarryStatus(pi.CarryStatus), Item: itemToAPI(item)}
}
