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

type LabelsHandler struct {
	store *store.Store
}

const (
	labelsErrNotFound       = "label not found"
	itemLabelsErrNotFound   = "item label not found"
	labelsErrInvalidBody    = "invalid request body"
	labelsErrFailedGetLabel = "failed to get label"
)

func NewLabelsHandler(st *store.Store) *LabelsHandler {
	return &LabelsHandler{store: st}
}

func (h *LabelsHandler) ListLabels(w http.ResponseWriter, r *http.Request) {
	labels, err := h.store.Labels.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list labels"})
		return
	}

	resp := make([]api.Label, len(labels))
	for i, l := range labels {
		resp[i] = labelToAPI(&l)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *LabelsHandler) CreateLabel(w http.ResponseWriter, r *http.Request) {
	var req api.LabelCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": labelsErrInvalidBody})
		return
	}
	defer r.Body.Close()

	label := &domain.Label{
		ID:    uuid.New(),
		Name:  req.Name,
		Color: req.Color,
	}

	if err := h.store.Labels.Create(r.Context(), label); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create label"})
		return
	}

	writeJSON(w, http.StatusCreated, labelToAPI(label))
}

func (h *LabelsHandler) GetLabel(w http.ResponseWriter, r *http.Request, labelId types.UUID) {
	label, err := h.store.Labels.GetByID(r.Context(), uuid.UUID(labelId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": labelsErrNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": labelsErrFailedGetLabel})
		return
	}

	writeJSON(w, http.StatusOK, labelToAPI(label))
}

func (h *LabelsHandler) UpdateLabel(w http.ResponseWriter, r *http.Request, labelId types.UUID) {
	var req api.LabelUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": labelsErrInvalidBody})
		return
	}
	defer r.Body.Close()

	label, err := h.store.Labels.GetByID(r.Context(), uuid.UUID(labelId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": labelsErrNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": labelsErrFailedGetLabel})
		return
	}

	if req.Name != nil {
		label.Name = *req.Name
	}
	if req.Color != nil {
		label.Color = req.Color
	}

	if err := h.store.Labels.Update(r.Context(), label); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update label"})
		return
	}

	writeJSON(w, http.StatusOK, labelToAPI(label))
}

func (h *LabelsHandler) DeleteLabel(w http.ResponseWriter, r *http.Request, labelId types.UUID) {
	err := h.store.Labels.Delete(r.Context(), uuid.UUID(labelId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": labelsErrNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete label"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LabelsHandler) ListItemLabels(w http.ResponseWriter, r *http.Request, itemId types.UUID) {
	// First verify item exists
	_, err := h.store.Items.GetByID(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item"})
		return
	}

	labels, err := h.store.Labels.ListItemLabels(r.Context(), uuid.UUID(itemId))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list item labels"})
		return
	}

	resp := make([]api.Label, len(labels))
	for i, l := range labels {
		resp[i] = labelToAPI(&l)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *LabelsHandler) AddItemLabel(w http.ResponseWriter, r *http.Request, itemId types.UUID) {
	var req api.ItemLabelAdd
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": labelsErrInvalidBody})
		return
	}
	defer r.Body.Close()

	// Verify item exists
	_, err := h.store.Items.GetByID(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item"})
		return
	}

	// Verify label exists
	label, err := h.store.Labels.GetByID(r.Context(), uuid.UUID(req.LabelId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": labelsErrNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": labelsErrFailedGetLabel})
		return
	}

	if err := h.store.Labels.AddItemLabel(r.Context(), uuid.UUID(itemId), uuid.UUID(req.LabelId)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add label to item"})
		return
	}

	writeJSON(w, http.StatusCreated, labelToAPI(label))
}

func (h *LabelsHandler) RemoveItemLabel(w http.ResponseWriter, r *http.Request, itemId types.UUID, labelId types.UUID) {
	err := h.store.Labels.RemoveItemLabel(r.Context(), uuid.UUID(itemId), uuid.UUID(labelId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": itemLabelsErrNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove label from item"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func labelToAPI(l *domain.Label) api.Label {
	return api.Label{
		Id:        types.UUID(l.ID),
		Name:      l.Name,
		Color:     l.Color,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}
