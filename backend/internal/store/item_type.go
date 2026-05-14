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

const itemTypesErrSystemDeleteForbidden = "system item types cannot be deleted"

type ItemTypeStore struct {
	db *sql.DB
}

func NewItemTypeStore(db *sql.DB) *ItemTypeStore {
	return &ItemTypeStore{db: db}
}

func (s *ItemTypeStore) List(ctx context.Context) ([]domain.ItemType, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description, base_profile, is_system, created_at, updated_at
		FROM item_types
		ORDER BY is_system DESC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list item types: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ItemType, 0)
	for rows.Next() {
		var it domain.ItemType
		var description sql.NullString
		var baseProfile sql.NullString
		if err := rows.Scan(&it.ID, &it.Name, &description, &baseProfile, &it.IsSystem, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan item type: %w", err)
		}
		it.Description = strPtr(description)
		it.BaseProfile = strPtr(baseProfile)
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate item types: %w", err)
	}

	return out, nil
}

func (s *ItemTypeStore) Create(ctx context.Context, itemType *domain.ItemType) error {
	query := `
		INSERT INTO item_types (id, name, description, base_profile, is_system)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`

	if err := s.db.QueryRowContext(
		ctx,
		query,
		itemType.ID,
		itemType.Name,
		toNullString(itemType.Description),
		toNullString(itemType.BaseProfile),
		itemType.IsSystem,
	).Scan(&itemType.CreatedAt, &itemType.UpdatedAt); err != nil {
		return fmt.Errorf("create item type: %w", err)
	}

	return nil
}

func (s *ItemTypeStore) GetByID(ctx context.Context, id string) (*domain.ItemType, error) {
	query := `
		SELECT id, name, description, base_profile, is_system, created_at, updated_at
		FROM item_types
		WHERE id = $1`

	var it domain.ItemType
	var description sql.NullString
	var baseProfile sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(&it.ID, &it.Name, &description, &baseProfile, &it.IsSystem, &it.CreatedAt, &it.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get item type by id: %w", err)
	}

	it.Description = strPtr(description)
	it.BaseProfile = strPtr(baseProfile)
	return &it, nil
}

func (s *ItemTypeStore) Update(ctx context.Context, itemType *domain.ItemType) error {
	query := `
		UPDATE item_types
		SET name = $2,
			description = $3,
			base_profile = $4,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := s.db.QueryRowContext(ctx, query, itemType.ID, itemType.Name, toNullString(itemType.Description), toNullString(itemType.BaseProfile)).Scan(&itemType.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update item type: %w", err)
	}

	return nil
}

func (s *ItemTypeStore) Delete(ctx context.Context, id string) error {
	itemType, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if itemType.IsSystem {
		return domain.ValidationError{Message: itemTypesErrSystemDeleteForbidden}
	}

	result, err := s.db.ExecContext(ctx, `DELETE FROM item_types WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete item type: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected on delete item type: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (s *ItemTypeStore) ListFields(ctx context.Context, itemTypeID string) ([]domain.ItemTypeField, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, item_type_id, field_key, field_label, data_type, is_required,
		       enum_options_json, min_value, max_value, unit, sort_order, created_at, updated_at
		FROM item_type_fields
		WHERE item_type_id = $1
		ORDER BY sort_order ASC, field_key ASC`, itemTypeID)
	if err != nil {
		return nil, fmt.Errorf("list item type fields: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ItemTypeField, 0)
	for rows.Next() {
		var field domain.ItemTypeField
		var enumOptions []byte
		var minValue sql.NullFloat64
		var maxValue sql.NullFloat64
		var unit sql.NullString

		if err := rows.Scan(
			&field.ID,
			&field.ItemTypeID,
			&field.FieldKey,
			&field.FieldLabel,
			&field.DataType,
			&field.IsRequired,
			&enumOptions,
			&minValue,
			&maxValue,
			&unit,
			&field.SortOrder,
			&field.CreatedAt,
			&field.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan item type field: %w", err)
		}

		if len(enumOptions) > 0 {
			if err := json.Unmarshal(enumOptions, &field.EnumOptions); err != nil {
				return nil, fmt.Errorf("decode item type field enum options: %w", err)
			}
		}
		field.MinValue = floatPtr(minValue)
		field.MaxValue = floatPtr(maxValue)
		field.Unit = strPtr(unit)
		out = append(out, field)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate item type fields: %w", err)
	}

	return out, nil
}

func (s *ItemTypeStore) ReplaceFields(ctx context.Context, itemTypeID string, fields []domain.ItemTypeField) ([]domain.ItemTypeField, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin replace item type fields: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM item_type_fields WHERE item_type_id = $1`, itemTypeID); err != nil {
		return nil, fmt.Errorf("delete existing item type fields: %w", err)
	}

	for i := range fields {
		f := &fields[i]
		if f.ID == uuid.Nil {
			f.ID = uuid.New()
		}
		f.ItemTypeID = itemTypeID

		var enumOptions []byte
		if len(f.EnumOptions) > 0 {
			enumOptions, err = json.Marshal(f.EnumOptions)
			if err != nil {
				return nil, fmt.Errorf("encode item type field enum options: %w", err)
			}
		}

		if err := tx.QueryRowContext(
			ctx,
			`INSERT INTO item_type_fields (
				id, item_type_id, field_key, field_label, data_type, is_required,
				enum_options_json, min_value, max_value, unit, sort_order
			) VALUES (
				$1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10, $11
			)
			RETURNING created_at, updated_at`,
			f.ID,
			f.ItemTypeID,
			f.FieldKey,
			f.FieldLabel,
			f.DataType,
			f.IsRequired,
			nullableJSON(enumOptions),
			toNullFloat64(f.MinValue),
			toNullFloat64(f.MaxValue),
			toNullString(f.Unit),
			f.SortOrder,
		).Scan(&f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("insert item type field: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit replace item type fields: %w", err)
	}

	return fields, nil
}

func nullableJSON(v []byte) any {
	if len(v) == 0 {
		return nil
	}
	return v
}
