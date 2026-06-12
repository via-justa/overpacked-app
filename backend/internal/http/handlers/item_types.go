package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/oapi-codegen/runtime/types"
	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

type ItemTypesHandler struct {
	store *store.Store
}

const itemTypesErrNotFound = "item type not found"
const itemTypesErrDeleteInUse = "category is used by items and cannot be deleted"
const itemTypesErrInvalidRequestBody = "invalid request body"
const itemTypesErrGetFailed = "failed to get item type"

const postgresUniqueViolationCode = "23505"
const postgresForeignKeyViolationCode = "23503"

func NewItemTypesHandler(st *store.Store) *ItemTypesHandler {
	return &ItemTypesHandler{store: st}
}

func (h *ItemTypesHandler) ListItemTypes(w http.ResponseWriter, r *http.Request) {
	itemTypes, err := h.store.ItemTypes.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list item types")
		return
	}

	resp := make([]api.ItemType, len(itemTypes))
	for i, itemType := range itemTypes {
		resp[i] = itemTypeToAPI(&itemType)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *ItemTypesHandler) CreateItemType(w http.ResponseWriter, r *http.Request) {
	var req api.ItemTypeCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, itemTypesErrInvalidRequestBody)
		return
	}
	defer r.Body.Close()

	itemType := &domain.ItemType{
		ID:       req.Id,
		Name:     req.Name,
		IsSystem: false,
	}
	if req.Description != nil {
		itemType.Description = req.Description
	}
	if req.BaseProfile != nil {
		baseProfile := string(*req.BaseProfile)
		itemType.BaseProfile = &baseProfile
	}

	if err := h.store.ItemTypes.Create(r.Context(), itemType); err != nil {
		if isPostgresUniqueViolation(err) {
			writeError(w, http.StatusConflict, "category already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create item type")
		return
	}

	writeJSON(w, http.StatusCreated, itemTypeToAPI(itemType))
}

func (h *ItemTypesHandler) GetItemType(w http.ResponseWriter, r *http.Request, typeID string) {
	itemType, err := h.store.ItemTypes.GetByID(r.Context(), typeID)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, itemTypesErrNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, itemTypesErrGetFailed)
		return
	}

	writeJSON(w, http.StatusOK, itemTypeToAPI(itemType))
}

func (h *ItemTypesHandler) UpdateItemType(w http.ResponseWriter, r *http.Request, typeID string) {
	var req api.ItemTypeUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, itemTypesErrInvalidRequestBody)
		return
	}
	defer r.Body.Close()

	itemType, err := h.store.ItemTypes.GetByID(r.Context(), typeID)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, itemTypesErrNotFound)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, itemTypesErrGetFailed)
		return
	}

	if req.Name != nil {
		itemType.Name = *req.Name
	}
	if req.Description != nil {
		itemType.Description = req.Description
	}
	if req.BaseProfile != nil {
		baseProfile := string(*req.BaseProfile)
		itemType.BaseProfile = &baseProfile
	}

	if err := h.store.ItemTypes.Update(r.Context(), itemType); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, itemTypesErrNotFound)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update item type")
		return
	}

	writeJSON(w, http.StatusOK, itemTypeToAPI(itemType))
}

func (h *ItemTypesHandler) ListItemTypeFields(w http.ResponseWriter, r *http.Request, typeID string) {
	if _, err := h.store.ItemTypes.GetByID(r.Context(), typeID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, itemTypesErrNotFound)
			return
		}
		writeError(w, http.StatusInternalServerError, itemTypesErrGetFailed)
		return
	}

	fields, err := h.store.ItemTypes.ListFields(r.Context(), typeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list item type fields")
		return
	}

	resp := make([]api.ItemTypeField, len(fields))
	for i, field := range fields {
		resp[i] = itemTypeFieldToAPI(&field)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *ItemTypesHandler) ReplaceItemTypeFields(w http.ResponseWriter, r *http.Request, typeID string) {
	var req api.ItemTypeFieldsReplace
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, itemTypesErrInvalidRequestBody)
		return
	}
	defer r.Body.Close()

	if _, err := h.store.ItemTypes.GetByID(r.Context(), typeID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, itemTypesErrNotFound)
			return
		}
		writeError(w, http.StatusInternalServerError, itemTypesErrGetFailed)
		return
	}

	fields := make([]domain.ItemTypeField, 0, len(req.Fields))
	for i, field := range req.Fields {
		sortOrder := i + 1
		if field.SortOrder != nil {
			sortOrder = *field.SortOrder
		}
		isRequired := false
		if field.IsRequired != nil {
			isRequired = *field.IsRequired
		}
		var enumOptions []string
		if field.EnumOptions != nil {
			enumOptions = append([]string(nil), (*field.EnumOptions)...)
		}

		fields = append(fields, domain.ItemTypeField{
			ID:          uuid.New(),
			ItemTypeID:  typeID,
			FieldKey:    field.FieldKey,
			FieldLabel:  field.FieldLabel,
			DataType:    string(field.DataType),
			IsRequired:  isRequired,
			EnumOptions: enumOptions,
			MinValue:    float64PtrFromFloat32(field.MinValue),
			MaxValue:    float64PtrFromFloat32(field.MaxValue),
			Unit:        field.Unit,
			SortOrder:   sortOrder,
		})
	}

	updatedFields, err := h.store.ItemTypes.ReplaceFields(r.Context(), typeID, fields)
	if err != nil {
		if isPostgresUniqueViolation(err) {
			writeError(w, http.StatusBadRequest, "field keys must be unique within a category")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to replace item type fields")
		return
	}

	resp := make([]api.ItemTypeField, len(updatedFields))
	for i, field := range updatedFields {
		resp[i] = itemTypeFieldToAPI(&field)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *ItemTypesHandler) DeleteItemType(w http.ResponseWriter, r *http.Request, typeID string) {
	err := h.store.ItemTypes.Delete(r.Context(), typeID)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, itemTypesErrNotFound)
		return
	}

	var validationErr domain.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Message)
		return
	}

	if isPostgresForeignKeyViolation(err) {
		writeError(w, http.StatusConflict, itemTypesErrDeleteInUse)
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete item type")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func itemTypeToAPI(itemType *domain.ItemType) api.ItemType {
	resp := api.ItemType{
		Id:        itemType.ID,
		Name:      itemType.Name,
		IsSystem:  itemType.IsSystem,
		CreatedAt: itemType.CreatedAt,
		UpdatedAt: itemType.UpdatedAt,
	}
	if itemType.Description != nil {
		resp.Description = itemType.Description
	}
	if itemType.BaseProfile != nil {
		baseProfile := api.ItemTypeBaseProfile(*itemType.BaseProfile)
		resp.BaseProfile = &baseProfile
	}
	return resp
}

func itemTypeFieldToAPI(field *domain.ItemTypeField) api.ItemTypeField {
	resp := api.ItemTypeField{
		Id:         types.UUID(field.ID),
		ItemTypeId: field.ItemTypeID,
		FieldKey:   field.FieldKey,
		FieldLabel: field.FieldLabel,
		DataType:   api.ItemTypeFieldDataType(field.DataType),
		IsRequired: field.IsRequired,
		SortOrder:  field.SortOrder,
		CreatedAt:  field.CreatedAt,
		UpdatedAt:  field.UpdatedAt,
	}
	if len(field.EnumOptions) > 0 {
		enumOptions := append([]string(nil), field.EnumOptions...)
		resp.EnumOptions = &enumOptions
	}
	resp.MinValue = float32PtrFromFloat64(field.MinValue)
	resp.MaxValue = float32PtrFromFloat64(field.MaxValue)
	resp.Unit = field.Unit
	return resp
}

func isPostgresUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}

	return string(pqErr.Code) == postgresUniqueViolationCode
}

func isPostgresForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}

	return string(pqErr.Code) == postgresForeignKeyViolationCode
}
