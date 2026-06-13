package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

type ItemsHandler struct {
	store *store.Store
}

const itemsErrNotFound = "item not found"

// maxImageBytes bounds a stored item image to guard against storage/DoS abuse.
const maxImageBytes = 5 << 20 // 5 MiB

func NewItemsHandler(st *store.Store) *ItemsHandler {
	return &ItemsHandler{store: st}
}

func (h *ItemsHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.Items.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list items")
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
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	if err := validateOptionalHTTPURL(req.SourceUrl); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

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
	if err := applyImage(item, req.ImageBlob, req.ImageMimeType, req.ImageSizeBytes); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.store.Items.Create(r.Context(), item); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create item")
		return
	}

	writeJSON(w, http.StatusCreated, itemToAPI(item))
}

func (h *ItemsHandler) GetItem(w http.ResponseWriter, r *http.Request, itemId types.UUID) {
	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, itemsErrNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get item")
		return
	}

	writeJSON(w, http.StatusOK, itemToAPI(item))
}

func (h *ItemsHandler) UpdateItem(w http.ResponseWriter, r *http.Request, itemId types.UUID) {
	var req api.ItemUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	if err := validateOptionalHTTPURL(req.SourceUrl); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, itemsErrNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get item")
		return
	}

	applyUpdateBaseFields(item, &req)
	if err := applyImage(item, req.ImageBlob, req.ImageMimeType, req.ImageSizeBytes); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.store.Items.Update(r.Context(), item); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, itemsErrNotFound)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update item")
		return
	}

	writeJSON(w, http.StatusOK, itemToAPI(item))
}

func (h *ItemsHandler) DeleteItem(w http.ResponseWriter, r *http.Request, itemId types.UUID) {
	err := h.store.Items.Delete(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, itemsErrNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	if item.ImageBlob != nil {
		resp.ImageBlob = bytesPtr(item.ImageBlob)
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

func bytesPtr(value []byte) *[]byte {
	converted := append([]byte(nil), value...)
	return &converted
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

// applyImage validates and copies an optional image (from a create or update
// request) onto the item, decoding its dimensions.
func applyImage(item *domain.Item, imageBlob *[]byte, mimeType *string, sizeBytes *int) error {
	if imageBlob == nil {
		return nil
	}
	if mimeType == nil || sizeBytes == nil {
		return errors.New("image metadata is required when image is provided")
	}
	if len(*imageBlob) > maxImageBytes {
		return errors.New("image exceeds maximum allowed size")
	}
	if *sizeBytes != len(*imageBlob) {
		return errors.New("image size does not match image content")
	}

	width, height, err := decodeImageDimensions(*imageBlob)
	if err != nil {
		return errors.New("invalid image content")
	}

	blob := append([]byte(nil), (*imageBlob)...)
	mime := *mimeType
	size := *sizeBytes

	item.ImageBlob = blob
	item.ImageMimeType = &mime
	item.ImageSizeBytes = &size
	item.ImageWidthPX = &width
	item.ImageHeightPX = &height
	return nil
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

func decodeImageDimensions(imageBlob []byte) (int, int, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(imageBlob))
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}
