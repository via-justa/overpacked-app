package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

type ItemStore struct {
	db *sql.DB
}

func NewItemStore(db *sql.DB) *ItemStore {
	return &ItemStore{db: db}
}

func (s *ItemStore) Create(ctx context.Context, item *domain.Item) error {
	query := `
		INSERT INTO items (
			manufacturer_id, type_id, name, is_active, description, source_url,
			price, weight_grams, volume_ml, default_quantity, default_carry_status, is_default,
			dose_count, calories, calories_per_serving, requires_water, season, layer, waterproof,
			size, color, capacity_people, season_rating, freestanding, has_footprint,
			comfort_temp_c, fill_type, r_value, capacity_mah, charge_port, rechargeable,
			image_blob, image_mime_type, image_size_bytes, image_width_px, image_height_px,
			attributes_json
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24, $25,
			$26, $27, $28, $29, $30, $31,
			$32, $33, $34, $35, $36, $37
		)
		RETURNING id, created_at, updated_at`

	attributes := item.Attributes
	if attributes == nil {
		attributes = map[string]any{}
	}
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("marshal item attributes: %w", err)
	}

	err = s.db.QueryRowContext(
		ctx,
		query,
		item.ManufacturerID,
		item.TypeID,
		item.Name,
		item.IsActive,
		toNullString(item.Description),
		toNullString(item.SourceURL),
		toNullFloat64(item.Price),
		toNullFloat64(item.WeightGrams),
		toNullFloat64(item.VolumeML),
		item.DefaultQuantity,
		string(item.DefaultCarryStatus),
		item.IsDefault,
		toNullInt(item.DoseCount),
		toNullFloat64(item.Calories),
		toNullFloat64(item.CaloriesPerServing),
		toNullBool(item.RequiresWater),
		toNullString(item.Season),
		toNullString(item.Layer),
		toNullBool(item.Waterproof),
		toNullString(item.Size),
		toNullString(item.Color),
		toNullFloat64(item.CapacityPeople),
		toNullString(item.SeasonRating),
		toNullBool(item.Freestanding),
		toNullBool(item.HasFootprint),
		toNullFloat64(item.ComfortTempC),
		toNullString(item.FillType),
		toNullFloat64(item.RValue),
		toNullInt(item.CapacityMAH),
		toNullString(item.ChargePort),
		toNullBool(item.Rechargeable),
		toNullBytes(item.ImageBlob),
		toNullString(item.ImageMimeType),
		toNullInt(item.ImageSizeBytes),
		toNullInt(item.ImageWidthPX),
		toNullInt(item.ImageHeightPX),
		attributesJSON,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create item: %w", err)
	}

	return nil
}

func (s *ItemStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	query := `
		SELECT
			id, manufacturer_id, type_id, name, is_active, description, source_url,
			price, weight_grams, volume_ml, default_quantity, default_carry_status,
			is_default, dose_count, calories, calories_per_serving, requires_water, season, layer, waterproof,
			size, color, capacity_people, season_rating, freestanding, has_footprint,
			comfort_temp_c, fill_type, r_value, capacity_mah, charge_port, rechargeable,
			image_blob, image_mime_type, image_size_bytes, image_width_px, image_height_px, attributes_json,
			created_at, updated_at
		FROM items
		WHERE id = $1`

	var item domain.Item
	var description sql.NullString
	var sourceURL sql.NullString
	var price sql.NullFloat64
	var weightGrams sql.NullFloat64
	var volumeML sql.NullFloat64
	var doseCount sql.NullInt64
	var calories sql.NullFloat64
	var caloriesPerServing sql.NullFloat64
	var requiresWater sql.NullBool
	var season sql.NullString
	var layer sql.NullString
	var waterproof sql.NullBool
	var size sql.NullString
	var color sql.NullString
	var capacityPeople sql.NullFloat64
	var seasonRating sql.NullString
	var freestanding sql.NullBool
	var hasFootprint sql.NullBool
	var comfortTempC sql.NullFloat64
	var fillType sql.NullString
	var rValue sql.NullFloat64
	var capacityMah sql.NullInt64
	var chargePort sql.NullString
	var rechargeable sql.NullBool
	var imageBlob []byte
	var imageMimeType sql.NullString
	var imageSizeBytes sql.NullInt64
	var imageWidthPX sql.NullInt64
	var imageHeightPX sql.NullInt64
	var attributesJSON []byte
	var defaultCarryStatus string

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.ManufacturerID,
		&item.TypeID,
		&item.Name,
		&item.IsActive,
		&description,
		&sourceURL,
		&price,
		&weightGrams,
		&volumeML,
		&item.DefaultQuantity,
		&defaultCarryStatus,
		&item.IsDefault,
		&doseCount,
		&calories,
		&caloriesPerServing,
		&requiresWater,
		&season,
		&layer,
		&waterproof,
		&size,
		&color,
		&capacityPeople,
		&seasonRating,
		&freestanding,
		&hasFootprint,
		&comfortTempC,
		&fillType,
		&rValue,
		&capacityMah,
		&chargePort,
		&rechargeable,
		&imageBlob,
		&imageMimeType,
		&imageSizeBytes,
		&imageWidthPX,
		&imageHeightPX,
		&attributesJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get item by id: %w", err)
	}

	item.Description = strPtr(description)
	item.SourceURL = strPtr(sourceURL)
	item.Price = floatPtr(price)
	item.WeightGrams = floatPtr(weightGrams)
	item.VolumeML = floatPtr(volumeML)
	item.DoseCount = intPtrFromNull(doseCount)
	item.Calories = floatPtr(calories)
	item.CaloriesPerServing = floatPtr(caloriesPerServing)
	item.RequiresWater = boolPtrFromNull(requiresWater)
	item.Season = strPtr(season)
	item.Layer = strPtr(layer)
	item.Waterproof = boolPtrFromNull(waterproof)
	item.Size = strPtr(size)
	item.Color = strPtr(color)
	item.CapacityPeople = floatPtr(capacityPeople)
	item.SeasonRating = strPtr(seasonRating)
	item.Freestanding = boolPtrFromNull(freestanding)
	item.HasFootprint = boolPtrFromNull(hasFootprint)
	item.ComfortTempC = floatPtr(comfortTempC)
	item.FillType = strPtr(fillType)
	item.RValue = floatPtr(rValue)
	item.CapacityMAH = intPtrFromNull(capacityMah)
	item.ChargePort = strPtr(chargePort)
	item.Rechargeable = boolPtrFromNull(rechargeable)
	item.ImageBlob = toNullBytes(imageBlob)
	item.ImageMimeType = strPtr(imageMimeType)
	item.ImageSizeBytes = intPtrFromNull(imageSizeBytes)
	item.ImageWidthPX = intPtrFromNull(imageWidthPX)
	item.ImageHeightPX = intPtrFromNull(imageHeightPX)
	if len(attributesJSON) > 0 {
		if err := json.Unmarshal(attributesJSON, &item.Attributes); err != nil {
			return nil, fmt.Errorf("decode item attributes: %w", err)
		}
	}
	item.DefaultCarryStatus = domain.CarryStatus(defaultCarryStatus)

	return &item, nil
}

func (s *ItemStore) List(ctx context.Context) ([]domain.Item, error) {
	query := `
		SELECT
			id, manufacturer_id, type_id, name, is_active, description, source_url,
			price, weight_grams, volume_ml, default_quantity, default_carry_status,
			is_default, dose_count, calories, calories_per_serving, requires_water, season, layer, waterproof,
			size, color, capacity_people, season_rating, freestanding, has_footprint,
			comfort_temp_c, fill_type, r_value, capacity_mah, charge_port, rechargeable,
			image_blob, image_mime_type, image_size_bytes, image_width_px, image_height_px, attributes_json,
			created_at, updated_at
		FROM items
		ORDER BY name ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Item, 0)
	for rows.Next() {
		var item domain.Item
		var description sql.NullString
		var sourceURL sql.NullString
		var price sql.NullFloat64
		var weightGrams sql.NullFloat64
		var volumeML sql.NullFloat64
		var doseCount sql.NullInt64
		var calories sql.NullFloat64
		var caloriesPerServing sql.NullFloat64
		var requiresWater sql.NullBool
		var season sql.NullString
		var layer sql.NullString
		var waterproof sql.NullBool
		var size sql.NullString
		var color sql.NullString
		var capacityPeople sql.NullFloat64
		var seasonRating sql.NullString
		var freestanding sql.NullBool
		var hasFootprint sql.NullBool
		var comfortTempC sql.NullFloat64
		var fillType sql.NullString
		var rValue sql.NullFloat64
		var capacityMah sql.NullInt64
		var chargePort sql.NullString
		var rechargeable sql.NullBool
		var imageBlob []byte
		var imageMimeType sql.NullString
		var imageSizeBytes sql.NullInt64
		var imageWidthPX sql.NullInt64
		var imageHeightPX sql.NullInt64
		var attributesJSON []byte
		var defaultCarryStatus string

		if err := rows.Scan(
			&item.ID,
			&item.ManufacturerID,
			&item.TypeID,
			&item.Name,
			&item.IsActive,
			&description,
			&sourceURL,
			&price,
			&weightGrams,
			&volumeML,
			&item.DefaultQuantity,
			&defaultCarryStatus,
			&item.IsDefault,
			&doseCount,
			&calories,
			&caloriesPerServing,
			&requiresWater,
			&season,
			&layer,
			&waterproof,
			&size,
			&color,
			&capacityPeople,
			&seasonRating,
			&freestanding,
			&hasFootprint,
			&comfortTempC,
			&fillType,
			&rValue,
			&capacityMah,
			&chargePort,
			&rechargeable,
			&imageBlob,
			&imageMimeType,
			&imageSizeBytes,
			&imageWidthPX,
			&imageHeightPX,
			&attributesJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}

		item.Description = strPtr(description)
		item.SourceURL = strPtr(sourceURL)
		item.Price = floatPtr(price)
		item.WeightGrams = floatPtr(weightGrams)
		item.VolumeML = floatPtr(volumeML)
		item.DoseCount = intPtrFromNull(doseCount)
		item.Calories = floatPtr(calories)
		item.CaloriesPerServing = floatPtr(caloriesPerServing)
		item.RequiresWater = boolPtrFromNull(requiresWater)
		item.Season = strPtr(season)
		item.Layer = strPtr(layer)
		item.Waterproof = boolPtrFromNull(waterproof)
		item.Size = strPtr(size)
		item.Color = strPtr(color)
		item.CapacityPeople = floatPtr(capacityPeople)
		item.SeasonRating = strPtr(seasonRating)
		item.Freestanding = boolPtrFromNull(freestanding)
		item.HasFootprint = boolPtrFromNull(hasFootprint)
		item.ComfortTempC = floatPtr(comfortTempC)
		item.FillType = strPtr(fillType)
		item.RValue = floatPtr(rValue)
		item.CapacityMAH = intPtrFromNull(capacityMah)
		item.ChargePort = strPtr(chargePort)
		item.Rechargeable = boolPtrFromNull(rechargeable)
		item.ImageBlob = toNullBytes(imageBlob)
		item.ImageMimeType = strPtr(imageMimeType)
		item.ImageSizeBytes = intPtrFromNull(imageSizeBytes)
		item.ImageWidthPX = intPtrFromNull(imageWidthPX)
		item.ImageHeightPX = intPtrFromNull(imageHeightPX)
		if len(attributesJSON) > 0 {
			if err := json.Unmarshal(attributesJSON, &item.Attributes); err != nil {
				return nil, fmt.Errorf("decode item attributes: %w", err)
			}
		}
		item.DefaultCarryStatus = domain.CarryStatus(defaultCarryStatus)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	return items, nil
}

func (s *ItemStore) Update(ctx context.Context, item *domain.Item) error {
	query := `
		UPDATE items
		SET manufacturer_id = $2,
			type_id = $3,
			name = $4,
			is_active = $5,
			description = $6,
			source_url = $7,
			price = $8,
			weight_grams = $9,
			volume_ml = $10,
			default_quantity = $11,
			default_carry_status = $12,
			is_default = $13,
			dose_count = $14,
			calories = $15,
			calories_per_serving = $16,
			requires_water = $17,
			season = $18,
			layer = $19,
			waterproof = $20,
			size = $21,
			color = $22,
			capacity_people = $23,
			season_rating = $24,
			freestanding = $25,
			has_footprint = $26,
			comfort_temp_c = $27,
			fill_type = $28,
			r_value = $29,
			capacity_mah = $30,
			charge_port = $31,
			rechargeable = $32,
			image_blob = $33,
			image_mime_type = $34,
			image_size_bytes = $35,
			image_width_px = $36,
			image_height_px = $37,
			attributes_json = $38,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	attributes := item.Attributes
	if attributes == nil {
		attributes = map[string]any{}
	}
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("marshal item attributes: %w", err)
	}

	err = s.db.QueryRowContext(
		ctx,
		query,
		item.ID,
		item.ManufacturerID,
		item.TypeID,
		item.Name,
		item.IsActive,
		toNullString(item.Description),
		toNullString(item.SourceURL),
		toNullFloat64(item.Price),
		toNullFloat64(item.WeightGrams),
		toNullFloat64(item.VolumeML),
		item.DefaultQuantity,
		string(item.DefaultCarryStatus),
		item.IsDefault,
		toNullInt(item.DoseCount),
		toNullFloat64(item.Calories),
		toNullFloat64(item.CaloriesPerServing),
		toNullBool(item.RequiresWater),
		toNullString(item.Season),
		toNullString(item.Layer),
		toNullBool(item.Waterproof),
		toNullString(item.Size),
		toNullString(item.Color),
		toNullFloat64(item.CapacityPeople),
		toNullString(item.SeasonRating),
		toNullBool(item.Freestanding),
		toNullBool(item.HasFootprint),
		toNullFloat64(item.ComfortTempC),
		toNullString(item.FillType),
		toNullFloat64(item.RValue),
		toNullInt(item.CapacityMAH),
		toNullString(item.ChargePort),
		toNullBool(item.Rechargeable),
		toNullBytes(item.ImageBlob),
		toNullString(item.ImageMimeType),
		toNullInt(item.ImageSizeBytes),
		toNullInt(item.ImageWidthPX),
		toNullInt(item.ImageHeightPX),
		attributesJSON,
	).Scan(&item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}

	return nil
}

func (s *ItemStore) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM items WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected on delete item: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
