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

func NewItemsHandler(st *store.Store) *ItemsHandler {
	return &ItemsHandler{store: st}
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
	applyCreateOptionalFields(item, &req)
	if err := applyCreateImage(item, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

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
	applyUpdateOptionalFields(item, &req)
	if err := applyUpdateImage(item, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

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
	err := h.store.Items.Delete(r.Context(), uuid.UUID(itemId))
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": itemsErrNotFound})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete item"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func itemToAPI(item *domain.Item) api.Item {
	var resp api.Item

	switch item.TypeID {
	case "wearable":
		_ = resp.FromWearableItem(buildWearableItem(item))
	case "shelter":
		_ = resp.FromShelterItem(buildShelterItem(item))
	case "sleep":
		_ = resp.FromSleepItem(buildSleepItem(item))
	case "electronics":
		_ = resp.FromElectronicsItem(buildElectronicsItem(item))
	case "consumable":
		_ = resp.FromConsumableItem(buildConsumableItem(item))
	default:
		resp = customItemToUnion(buildCustomItem(item))
	}

	return resp
}

func buildConsumableItem(item *domain.Item) api.ConsumableItem {
	resp := api.ConsumableItem{
		CreatedAt:       item.CreatedAt,
		DefaultQuantity: intPtr(item.DefaultQuantity),
		Description:     item.Description,
		DoseCount:       item.DoseCount,
		Id:              types.UUID(item.ID),
		ImageMimeType:   item.ImageMimeType,
		ImageSizeBytes:  item.ImageSizeBytes,
		ImageWidthPx:    item.ImageWidthPX,
		ImageHeightPx:   item.ImageHeightPX,
		IsActive:        item.IsActive,
		IsDefault:       boolPtr(item.IsDefault),
		ManufacturerId:  types.UUID(item.ManufacturerID),
		Name:            item.Name,
		RequiresWater:   item.RequiresWater,
		SourceUrl:       item.SourceURL,
		Type:            api.ConsumableItemTypeConsumable,
		UpdatedAt:       item.UpdatedAt,
		VolumeMl:        float32PtrFromFloat64(item.VolumeML),
		WeightGrams:     float32PtrFromFloat64(item.WeightGrams),
		Attributes:      attributesPtr(item.Attributes),
	}
	resp.DefaultCarryStatus = itemDefaultCarryStatusPtr(api.ConsumableItemDefaultCarryStatus(item.DefaultCarryStatus))
	resp.Value = float32PtrFromFloat64(item.Price)
	if item.ImageBlob != nil {
		resp.ImageBlob = bytesPtr(item.ImageBlob)
	}
	if item.Calories != nil {
		resp.Calories = float32PtrFromFloat64(item.Calories)
	}
	if item.CaloriesPerServing != nil {
		resp.CaloriesPerServing = float32PtrFromFloat64(item.CaloriesPerServing)
	}
	return resp
}

func buildWearableItem(item *domain.Item) api.WearableItem {
	resp := api.WearableItem{
		Color:           item.Color,
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
		Size:            item.Size,
		SourceUrl:       item.SourceURL,
		Type:            api.Wearable,
		UpdatedAt:       item.UpdatedAt,
		Value:           float32PtrFromFloat64(item.Price),
		VolumeMl:        float32PtrFromFloat64(item.VolumeML),
		WeightGrams:     float32PtrFromFloat64(item.WeightGrams),
		Waterproof:      item.Waterproof,
		Attributes:      attributesPtr(item.Attributes),
	}
	resp.DefaultCarryStatus = itemDefaultCarryStatusPtr(api.WearableItemDefaultCarryStatus(item.DefaultCarryStatus))
	if item.Season != nil {
		season := api.WearableItemSeason(*item.Season)
		resp.Season = &season
	}
	if item.Layer != nil {
		layer := api.WearableItemLayer(*item.Layer)
		resp.Layer = &layer
	}
	if item.ImageBlob != nil {
		resp.ImageBlob = bytesPtr(item.ImageBlob)
	}
	return resp
}

func buildShelterItem(item *domain.Item) api.ShelterItem {
	resp := api.ShelterItem{
		CapacityPeople:  float32PtrFromFloat64(item.CapacityPeople),
		CreatedAt:       item.CreatedAt,
		DefaultQuantity: intPtr(item.DefaultQuantity),
		Description:     item.Description,
		Freestanding:    item.Freestanding,
		HasFootprint:    item.HasFootprint,
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
		Type:            api.ShelterItemTypeShelter,
		UpdatedAt:       item.UpdatedAt,
		Value:           float32PtrFromFloat64(item.Price),
		VolumeMl:        float32PtrFromFloat64(item.VolumeML),
		WeightGrams:     float32PtrFromFloat64(item.WeightGrams),
		Attributes:      attributesPtr(item.Attributes),
	}
	resp.DefaultCarryStatus = itemDefaultCarryStatusPtr(api.ShelterItemDefaultCarryStatus(item.DefaultCarryStatus))
	if item.SeasonRating != nil {
		rating := api.ShelterItemSeasonRating(*item.SeasonRating)
		resp.SeasonRating = &rating
	}
	if item.ImageBlob != nil {
		resp.ImageBlob = bytesPtr(item.ImageBlob)
	}
	return resp
}

func buildSleepItem(item *domain.Item) api.SleepItem {
	resp := api.SleepItem{
		ComfortTempC:    float32PtrFromFloat64(item.ComfortTempC),
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
		RValue:          float32PtrFromFloat64(item.RValue),
		SourceUrl:       item.SourceURL,
		Type:            api.Sleep,
		UpdatedAt:       item.UpdatedAt,
		Value:           float32PtrFromFloat64(item.Price),
		VolumeMl:        float32PtrFromFloat64(item.VolumeML),
		WeightGrams:     float32PtrFromFloat64(item.WeightGrams),
		Attributes:      attributesPtr(item.Attributes),
	}
	resp.DefaultCarryStatus = itemDefaultCarryStatusPtr(api.SleepItemDefaultCarryStatus(item.DefaultCarryStatus))
	if item.FillType != nil {
		fillType := api.SleepItemFillType(*item.FillType)
		resp.FillType = &fillType
	}
	if item.ImageBlob != nil {
		resp.ImageBlob = bytesPtr(item.ImageBlob)
	}
	return resp
}

func buildElectronicsItem(item *domain.Item) api.ElectronicsItem {
	resp := api.ElectronicsItem{
		CapacityMah:     item.CapacityMAH,
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
		Rechargeable:    item.Rechargeable,
		SourceUrl:       item.SourceURL,
		Type:            api.ElectronicsItemTypeElectronics,
		UpdatedAt:       item.UpdatedAt,
		Value:           float32PtrFromFloat64(item.Price),
		VolumeMl:        float32PtrFromFloat64(item.VolumeML),
		WeightGrams:     float32PtrFromFloat64(item.WeightGrams),
		Attributes:      attributesPtr(item.Attributes),
	}
	resp.DefaultCarryStatus = itemDefaultCarryStatusPtr(api.ElectronicsItemDefaultCarryStatus(item.DefaultCarryStatus))
	if item.ChargePort != nil {
		chargePort := api.ElectronicsItemChargePort(*item.ChargePort)
		resp.ChargePort = &chargePort
	}
	if item.ImageBlob != nil {
		resp.ImageBlob = bytesPtr(item.ImageBlob)
	}
	return resp
}

func itemDefaultCarryStatusPtr[T ~string](status T) *T {
	value := status
	return &value
}

func buildCustomItem(item *domain.Item) api.CustomItem {
	resp := api.CustomItem{
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
		VolumeMl:        float32PtrFromFloat64(item.VolumeML),
		WeightGrams:     float32PtrFromFloat64(item.WeightGrams),
		Attributes:      attributesPtr(item.Attributes),
	}
	resp.DefaultCarryStatus = itemDefaultCarryStatusPtr(api.CustomItemDefaultCarryStatus(item.DefaultCarryStatus))
	resp.Value = float32PtrFromFloat64(item.Price)
	if item.ImageBlob != nil {
		resp.ImageBlob = bytesPtr(item.ImageBlob)
	}
	return resp
}

func customItemToUnion(custom api.CustomItem) api.Item {
	var resp api.Item
	b, err := json.Marshal(custom)
	if err != nil {
		return resp
	}
	if err := resp.UnmarshalJSON(b); err != nil {
		return api.Item{}
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

func applyCreateOptionalFields(item *domain.Item, req *api.ItemCreate) {
	item.DoseCount = req.DoseCount
	item.Calories = float64PtrFromFloat32(req.Calories)
	item.CaloriesPerServing = float64PtrFromFloat32(req.CaloriesPerServing)
	item.RequiresWater = req.RequiresWater
	item.Season = enumStringPtr(req.Season)
	item.Layer = enumStringPtr(req.Layer)
	item.Waterproof = req.Waterproof
	item.Size = req.Size
	item.Color = req.Color
	item.CapacityPeople = float64PtrFromFloat32(req.CapacityPeople)
	item.SeasonRating = enumStringPtr(req.SeasonRating)
	item.Freestanding = req.Freestanding
	item.HasFootprint = req.HasFootprint
	item.ComfortTempC = float64PtrFromFloat32(req.ComfortTempC)
	item.FillType = enumStringPtr(req.FillType)
	item.RValue = float64PtrFromFloat32(req.RValue)
	item.CapacityMAH = req.CapacityMah
	item.ChargePort = enumStringPtr(req.ChargePort)
	item.Rechargeable = req.Rechargeable
}

func applyCreateImage(item *domain.Item, req *api.ItemCreate) error {
	if req.ImageBlob == nil {
		return nil
	}
	if req.ImageMimeType == nil || req.ImageSizeBytes == nil {
		return errors.New("image metadata is required when image is provided")
	}

	width, height, err := decodeImageDimensions(*req.ImageBlob)
	if err != nil {
		return errors.New("invalid image content")
	}

	imageBlob := append([]byte(nil), (*req.ImageBlob)...)
	imageMimeType := *req.ImageMimeType
	imageSizeBytes := *req.ImageSizeBytes

	item.ImageBlob = imageBlob
	item.ImageMimeType = &imageMimeType
	item.ImageSizeBytes = &imageSizeBytes
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

func applyUpdateOptionalFields(item *domain.Item, req *api.ItemUpdate) {
	applyUpdateOptionalConsumableWearableFields(item, req)
	applyUpdateOptionalShelterSleepElectronicsFields(item, req)
}

func applyUpdateOptionalConsumableWearableFields(item *domain.Item, req *api.ItemUpdate) {
	if req.DoseCount != nil {
		item.DoseCount = req.DoseCount
	}
	if req.Calories != nil {
		item.Calories = float64PtrFromFloat32(req.Calories)
	}
	if req.CaloriesPerServing != nil {
		item.CaloriesPerServing = float64PtrFromFloat32(req.CaloriesPerServing)
	}
	if req.RequiresWater != nil {
		item.RequiresWater = req.RequiresWater
	}
	if req.Season != nil {
		item.Season = enumStringPtr(req.Season)
	}
	if req.Layer != nil {
		item.Layer = enumStringPtr(req.Layer)
	}
	if req.Waterproof != nil {
		item.Waterproof = req.Waterproof
	}
	if req.Size != nil {
		item.Size = req.Size
	}
	if req.Color != nil {
		item.Color = req.Color
	}
}

func applyUpdateOptionalShelterSleepElectronicsFields(item *domain.Item, req *api.ItemUpdate) {
	if req.CapacityPeople != nil {
		item.CapacityPeople = float64PtrFromFloat32(req.CapacityPeople)
	}
	if req.SeasonRating != nil {
		item.SeasonRating = enumStringPtr(req.SeasonRating)
	}
	if req.Freestanding != nil {
		item.Freestanding = req.Freestanding
	}
	if req.HasFootprint != nil {
		item.HasFootprint = req.HasFootprint
	}
	if req.ComfortTempC != nil {
		item.ComfortTempC = float64PtrFromFloat32(req.ComfortTempC)
	}
	if req.FillType != nil {
		item.FillType = enumStringPtr(req.FillType)
	}
	if req.RValue != nil {
		item.RValue = float64PtrFromFloat32(req.RValue)
	}
	if req.CapacityMah != nil {
		item.CapacityMAH = req.CapacityMah
	}
	if req.ChargePort != nil {
		item.ChargePort = enumStringPtr(req.ChargePort)
	}
	if req.Rechargeable != nil {
		item.Rechargeable = req.Rechargeable
	}
}

func applyUpdateImage(item *domain.Item, req *api.ItemUpdate) error {
	if req.ImageBlob == nil {
		return nil
	}
	if req.ImageMimeType == nil || req.ImageSizeBytes == nil {
		return errors.New("image metadata is required when image is provided")
	}

	width, height, err := decodeImageDimensions(*req.ImageBlob)
	if err != nil {
		return errors.New("invalid image content")
	}

	imageBlob := append([]byte(nil), (*req.ImageBlob)...)
	imageMimeType := *req.ImageMimeType
	imageSizeBytes := *req.ImageSizeBytes

	item.ImageBlob = imageBlob
	item.ImageMimeType = &imageMimeType
	item.ImageSizeBytes = &imageSizeBytes
	item.ImageWidthPX = &width
	item.ImageHeightPX = &height
	return nil
}

func float64PtrFromFloat32(value *float32) *float64 {
	if value == nil {
		return nil
	}
	converted := float64(*value)
	return &converted
}

func enumStringPtr[T ~string](value *T) *string {
	if value == nil {
		return nil
	}
	converted := string(*value)
	return &converted
}

func decodeImageDimensions(imageBlob []byte) (int, int, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(imageBlob))
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}
