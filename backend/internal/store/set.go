package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

type SetStore struct {
	db *sql.DB
}

func NewSetStore(db *sql.DB) *SetStore {
	return &SetStore{db: db}
}

func (s *SetStore) Create(ctx context.Context, set *domain.ItemSet) error {
	query := `
		INSERT INTO item_sets (name, set_category, description, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	if err := s.db.QueryRowContext(ctx, query, set.Name, set.SetCategory, toNullString(set.Description), set.IsActive).Scan(
		&set.ID,
		&set.CreatedAt,
		&set.UpdatedAt,
	); err != nil {
		return fmt.Errorf("create set: %w", err)
	}

	return nil
}

func (s *SetStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.ItemSet, error) {
	query := `
		SELECT id, name, set_category, description, is_active, created_at, updated_at
		FROM item_sets
		WHERE id = $1`

	var set domain.ItemSet
	var description sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&set.ID,
		&set.Name,
		&set.SetCategory,
		&description,
		&set.IsActive,
		&set.CreatedAt,
		&set.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get set by id: %w", err)
	}

	set.Description = strPtr(description)
	return &set, nil
}

func (s *SetStore) List(ctx context.Context) ([]domain.ItemSet, error) {
	query := `
		SELECT id, name, set_category, description, is_active, created_at, updated_at
		FROM item_sets
		ORDER BY name ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list sets: %w", err)
	}
	defer rows.Close()

	sets := make([]domain.ItemSet, 0)
	for rows.Next() {
		var set domain.ItemSet
		var description sql.NullString

		if err := rows.Scan(
			&set.ID,
			&set.Name,
			&set.SetCategory,
			&description,
			&set.IsActive,
			&set.CreatedAt,
			&set.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan set: %w", err)
		}

		set.Description = strPtr(description)
		sets = append(sets, set)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sets: %w", err)
	}

	return sets, nil
}

func (s *SetStore) Update(ctx context.Context, set *domain.ItemSet) error {
	query := `
		UPDATE item_sets
		SET name = $2,
			set_category = $3,
			description = $4,
			is_active = $5,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := s.db.QueryRowContext(ctx, query, set.ID, set.Name, set.SetCategory, toNullString(set.Description), set.IsActive).Scan(&set.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update set: %w", err)
	}

	return nil
}

func (s *SetStore) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM item_sets WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete set: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected on delete set: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (s *SetStore) AddItem(ctx context.Context, item *domain.SetItem) error {
	query := `
		INSERT INTO set_items (set_id, item_id, quantity, notes, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	if err := s.db.QueryRowContext(ctx, query, item.SetID, item.ItemID, item.Quantity, toNullString(item.Notes), item.SortOrder).Scan(&item.ID); err != nil {
		return fmt.Errorf("add set item: %w", err)
	}

	return nil
}

func (s *SetStore) ListItems(ctx context.Context, setID uuid.UUID) ([]domain.SetItem, error) {
	query := `
		SELECT id, set_id, item_id, quantity, notes, sort_order
		FROM set_items
		WHERE set_id = $1
		ORDER BY sort_order ASC, id ASC`

	rows, err := s.db.QueryContext(ctx, query, setID)
	if err != nil {
		return nil, fmt.Errorf("list set items: %w", err)
	}
	defer rows.Close()

	out := make([]domain.SetItem, 0)
	for rows.Next() {
		var item domain.SetItem
		var notes sql.NullString
		if err := rows.Scan(&item.ID, &item.SetID, &item.ItemID, &item.Quantity, &notes, &item.SortOrder); err != nil {
			return nil, fmt.Errorf("scan set item: %w", err)
		}
		item.Notes = strPtr(notes)
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate set items: %w", err)
	}

	return out, nil
}

func (s *SetStore) UpdateItem(ctx context.Context, item *domain.SetItem) error {
	query := `
		UPDATE set_items
		SET quantity = $3,
			notes = $4,
			sort_order = $5
		WHERE set_id = $1 AND item_id = $2`

	result, err := s.db.ExecContext(ctx, query, item.SetID, item.ItemID, item.Quantity, toNullString(item.Notes), item.SortOrder)
	if err != nil {
		return fmt.Errorf("update set item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected on update set item: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (s *SetStore) RemoveItem(ctx context.Context, setID uuid.UUID, itemID uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM set_items WHERE set_id = $1 AND item_id = $2", setID, itemID)
	if err != nil {
		return fmt.Errorf("remove set item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected on remove set item: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
