package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
	"github.com/via-justa/overpacked-app/backend/internal/storage"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

type ItemsHandler struct {
	store  *store.Store
	images *storage.ImageStore
}

const (
	itemsErrNotFound  = "item not found"
	imageFormField    = "file"
	imageMaxFormBytes = storage.MaxImageBytes + (1 << 20) // image cap + multipart overhead
)

func NewItemsHandler(st *store.Store, images *storage.ImageStore) *ItemsHandler {
	return &ItemsHandler{store: st, images: images}
}

func (h *ItemsHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.Items.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list items"})
		return
	}

	resp := make([]api.Item, len(items))
	for i, item := range items {
		resp[i] = itemToAPI(&item)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *ItemsHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req api.ItemCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	defer r.Body.Close()

	item := &domain.Item{
		ID:                 uuid.New(),
		ManufacturerID:     uuid.UUID(req.ManufacturerId),
		TypeID:             string(req.Type),
		Name:               req.Name,
		IsActive:           req.IsActive,
		DefaultQuantity:    1,
		DefaultCarryStatus: domain.CarryStatusPacked,
		IsDefault:          false,
	}

	applyCreateBaseFields(item, &req)

	if err := h.store.Items.Create(r.Context(), item); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create item"})
		return
	}

	writeJSON(w, http.StatusCreated, itemToAPI(item))
}

func (h *ItemsHandler) GetItem(w http.ResponseWriter, r *http.Request, itemId types.UUID) {
	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": itemsErrNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item"})
		return
	}

	writeJSON(w, http.StatusOK, itemToAPI(item))
}

func (h *ItemsHandler) UpdateItem(w http.ResponseWriter, r *http.Request, itemId types.UUID) {
	var req api.ItemUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	defer r.Body.Close()

	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": itemsErrNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item"})
		return
	}

	applyUpdateBaseFields(item, &req)

	if err := h.store.Items.Update(r.Context(), item); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": itemsErrNotFound})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update item"})
		return
	}

	writeJSON(w, http.StatusOK, itemToAPI(item))
}

func (h *ItemsHandler) DeleteItem(w http.ResponseWriter, r *http.Request, itemId types.UUID) {
	oldPath, err := h.store.Items.Delete(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": itemsErrNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete item"})
		return
	}

	if oldPath != nil {
		_ = h.images.Delete(*oldPath)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ItemsHandler) UploadItemImage(w http.ResponseWriter, r *http.Request, itemId types.UUID) {
	r.Body = http.MaxBytesReader(w, r.Body, imageMaxFormBytes)
	if err := r.ParseMultipartForm(imageMaxFormBytes); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
		return
	}

	file, _, err := r.FormFile(imageFormField)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing image file"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read image"})
		return
	}

	mimeType := http.DetectContentType(data)

	path, width, height, err := h.images.Save(uuid.UUID(itemId), mimeType, data)
	if errors.Is(err, storage.ErrUnsupportedType) || errors.Is(err, storage.ErrInvalidImage) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported or invalid image"})
		return
	}
	if errors.Is(err, storage.ErrTooLarge) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "image is too large"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to store image"})
		return
	}

	oldPath, err := h.store.Items.SetImage(r.Context(), uuid.UUID(itemId), path, mimeType, len(data), width, height)
	if errors.Is(err, domain.ErrNotFound) {
		_ = h.images.Delete(path)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": itemsErrNotFound})
		return
	}
	if err != nil {
		_ = h.images.Delete(path)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save image"})
		return
	}

	if oldPath != nil && *oldPath != path {
		_ = h.images.Delete(*oldPath)
	}

	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(itemId))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item"})
		return
	}

	writeJSON(w, http.StatusOK, itemToAPI(item))
}

func (h *ItemsHandler) DeleteItemImage(w http.ResponseWriter, r *http.Request, itemId types.UUID) {
	oldPath, err := h.store.Items.ClearImage(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": itemsErrNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete image"})
		return
	}

	if oldPath != nil {
		_ = h.images.Delete(*oldPath)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ItemsHandler) GetItemImage(w http.ResponseWriter, r *http.Request, itemId types.UUID) {
	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": itemsErrNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get item"})
		return
	}
	if item.ImagePath == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "item has no image"})
		return
	}

	file, err := h.images.Open(*item.ImagePath)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "image not found"})
		return
	}
	defer file.Close()

	if item.ImageMimeType != nil {
		w.Header().Set("Content-Type", *item.ImageMimeType)
	}
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	w.Header().Set("ETag", `"`+*item.ImagePath+`"`)

	if match := r.Header.Get("If-None-Match"); match != "" && match == `"`+*item.ImagePath+`"` {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	_, _ = io.Copy(w, file)
}

func itemToAPI(item *domain.Item) api.Item {
	resp := api.Item{
		CreatedAt:       item.CreatedAt,
		DefaultQuantity: intPtr(item.DefaultQuantity),
		Description:     item.Description,
		Id:              types.UUID(item.ID),
		ImageMimeType:   item.ImageMimeType,
		ImageSizeBytes:  item.ImageSizeBytes,
		ImageWidthPx:    item.ImageWidthPX,
		ImageHeightPx:   item.ImageHeightPX,
		IsActive:        item.IsActive,
		IsDefault:       boolPtr(item.IsDefault),
		ManufacturerId:  types.UUID(item.ManufacturerID),
		Name:            item.Name,
		SourceUrl:       item.SourceURL,
		Type:            item.TypeID,
		UpdatedAt:       item.UpdatedAt,
		Value:           float32PtrFromFloat64(item.Price),
		VolumeMl:        float32PtrFromFloat64(item.VolumeML),
		WeightGrams:     float32PtrFromFloat64(item.WeightGrams),
		Attributes:      attributesPtr(item.Attributes),
	}
	if item.DefaultCarryStatus != "" {
		status := api.ItemBaseDefaultCarryStatus(item.DefaultCarryStatus)
		resp.DefaultCarryStatus = &status
	}
	if item.ImagePath != nil {
		url := "/api/v1/items/" + item.ID.String() + "/image"
		resp.ImageUrl = &url
	}
	return resp
}

func attributesPtr(value map[string]any) *map[string]any {
	if value == nil {
		return nil
	}

	cloned := make(map[string]any, len(value))
	for key, item := range value {
		cloned[key] = item
	}

	return &cloned
}

func float32PtrFromFloat64(value *float64) *float32 {
	if value == nil {
		return nil
	}
	converted := float32(*value)
	return &converted
}

func boolPtr(value bool) *bool {
	return &value
}

func intPtr(value int) *int {
	return &value
}

func applyCreateBaseFields(item *domain.Item, req *api.ItemCreate) {
	item.Description = req.Description
	item.SourceURL = req.SourceUrl
	item.Price = float64PtrFromFloat32(req.Value)
	item.WeightGrams = float64PtrFromFloat32(req.WeightGrams)
	item.VolumeML = float64PtrFromFloat32(req.VolumeMl)
	item.Attributes = attributesFromReq(req.Attributes)
	if req.DefaultQuantity != nil {
		item.DefaultQuantity = *req.DefaultQuantity
	}
	if req.DefaultCarryStatus != nil {
		item.DefaultCarryStatus = domain.CarryStatus(*req.DefaultCarryStatus)
	}
	if req.IsDefault != nil {
		item.IsDefault = *req.IsDefault
	}
}

func attributesFromReq(value *map[string]interface{}) map[string]any {
	if value == nil {
		return nil
	}

	cloned := make(map[string]any, len(*value))
	for key, v := range *value {
		cloned[key] = v
	}

	return cloned
}

func applyUpdateBaseFields(item *domain.Item, req *api.ItemUpdate) {
	if req.ManufacturerId != nil {
		item.ManufacturerID = uuid.UUID(*req.ManufacturerId)
	}
	if req.Type != nil {
		item.TypeID = string(*req.Type)
	}
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.IsActive != nil {
		item.IsActive = *req.IsActive
	}
	if req.Description != nil {
		item.Description = req.Description
	}
	if req.SourceUrl != nil {
		item.SourceURL = req.SourceUrl
	}
	if req.WeightGrams != nil {
		item.WeightGrams = float64PtrFromFloat32(req.WeightGrams)
	}
	if req.Value != nil {
		item.Price = float64PtrFromFloat32(req.Value)
	}
	if req.VolumeMl != nil {
		item.VolumeML = float64PtrFromFloat32(req.VolumeMl)
	}
	if req.DefaultQuantity != nil {
		item.DefaultQuantity = *req.DefaultQuantity
	}
	if req.DefaultCarryStatus != nil {
		item.DefaultCarryStatus = domain.CarryStatus(*req.DefaultCarryStatus)
	}
	if req.IsDefault != nil {
		item.IsDefault = *req.IsDefault
	}
	if req.Attributes != nil {
		item.Attributes = attributesFromReq(req.Attributes)
	}
}

func float64PtrFromFloat32(value *float32) *float64 {
	if value == nil {
		return nil
	}
	converted := float64(*value)
	return &converted
}
