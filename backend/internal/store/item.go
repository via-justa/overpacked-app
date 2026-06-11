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
			image_path, image_mime_type, image_size_bytes, image_width_px, image_height_px,
			attributes_json
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18
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
		toNullString(item.ImagePath),
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
			is_default, image_path, image_mime_type, image_size_bytes, image_width_px, image_height_px, attributes_json,
			created_at, updated_at
		FROM items
		WHERE id = $1`

	var item domain.Item
	var description sql.NullString
	var sourceURL sql.NullString
	var price sql.NullFloat64
	var weightGrams sql.NullFloat64
	var volumeML sql.NullFloat64
	var imagePath sql.NullString
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
		&imagePath,
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
	item.ImagePath = strPtr(imagePath)
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
			is_default, image_path, image_mime_type, image_size_bytes, image_width_px, image_height_px, attributes_json,
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
		var imagePath sql.NullString
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
			&imagePath,
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
		item.ImagePath = strPtr(imagePath)
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
			image_path = $14,
			image_mime_type = $15,
			image_size_bytes = $16,
			image_width_px = $17,
			image_height_px = $18,
			attributes_json = $19,
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
		toNullString(item.ImagePath),
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

// Delete removes the item and returns the image path it referenced (if any) so
// the caller can clean up the orphaned file on disk.
func (s *ItemStore) Delete(ctx context.Context, id uuid.UUID) (oldPath *string, err error) {
	var imagePath sql.NullString
	err = s.db.QueryRowContext(ctx, "DELETE FROM items WHERE id = $1 RETURNING image_path", id).Scan(&imagePath)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("delete item: %w", err)
	}

	return strPtr(imagePath), nil
}

// SetImage records an item's image path and metadata, returning the previous
// image path (if any) so the caller can remove the replaced file.
func (s *ItemStore) SetImage(ctx context.Context, id uuid.UUID, path, mimeType string, sizeBytes, width, height int) (oldPath *string, err error) {
	query := `
		WITH old AS (SELECT image_path FROM items WHERE id = $1)
		UPDATE items
		SET image_path = $2,
			image_mime_type = $3,
			image_size_bytes = $4,
			image_width_px = $5,
			image_height_px = $6,
			updated_at = NOW()
		WHERE id = $1
		RETURNING (SELECT image_path FROM old)`

	var previous sql.NullString
	err = s.db.QueryRowContext(ctx, query, id, path, mimeType, sizeBytes, width, height).Scan(&previous)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("set item image: %w", err)
	}

	return strPtr(previous), nil
}

// ClearImage removes an item's image metadata, returning the previous image
// path (if any) so the caller can delete the file.
func (s *ItemStore) ClearImage(ctx context.Context, id uuid.UUID) (oldPath *string, err error) {
	query := `
		WITH old AS (SELECT image_path FROM items WHERE id = $1)
		UPDATE items
		SET image_path = NULL,
			image_mime_type = NULL,
			image_size_bytes = NULL,
			image_width_px = NULL,
			image_height_px = NULL,
			updated_at = NOW()
		WHERE id = $1
		RETURNING (SELECT image_path FROM old)`

	var previous sql.NullString
	err = s.db.QueryRowContext(ctx, query, id).Scan(&previous)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("clear item image: %w", err)
	}

	return strPtr(previous), nil
}
