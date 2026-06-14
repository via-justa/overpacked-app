package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

// PackingListStore handles database operations for packing lists
type PackingListStore struct {
	db *sql.DB
}

// NewPackingListStore creates a new packing list store
func NewPackingListStore(db *sql.DB) *PackingListStore {
	return &PackingListStore{db: db}
}

// Create creates a new packing list
func (s *PackingListStore) Create(ctx context.Context, name string, description *string) (*domain.PackingList, error) {
	query := `
		INSERT INTO packing_lists (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at`

	var pl domain.PackingList
	var desc sql.NullString
	err := s.db.QueryRowContext(ctx, query, name, description).Scan(
		&pl.ID, &pl.Name, &desc, &pl.CreatedAt, &pl.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	pl.Description = strPtr(desc)
	return &pl, nil
}

// GetByID retrieves a packing list by ID
func (s *PackingListStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.PackingList, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM packing_lists
		WHERE id = $1`

	var pl domain.PackingList
	var desc sql.NullString
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&pl.ID, &pl.Name, &desc, &pl.CreatedAt, &pl.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	pl.Description = strPtr(desc)
	return &pl, nil
}

// List returns all packing lists
func (s *PackingListStore) List(ctx context.Context) ([]domain.PackingList, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM packing_lists
		ORDER BY name ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []domain.PackingList
	for rows.Next() {
		var pl domain.PackingList
		var desc sql.NullString
		if err := rows.Scan(&pl.ID, &pl.Name, &desc, &pl.CreatedAt, &pl.UpdatedAt); err != nil {
			return nil, err
		}
		pl.Description = strPtr(desc)
		lists = append(lists, pl)
	}
	return lists, rows.Err()
}

// Update updates an existing packing list
func (s *PackingListStore) Update(ctx context.Context, id uuid.UUID, name *string, description *string) (*domain.PackingList, error) {
	// Check if packing list exists
	existing, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Use existing values if not provided
	if name == nil {
		name = &existing.Name
	}
	if description == nil {
		description = existing.Description
	}

	query := `
		UPDATE packing_lists
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, description, created_at, updated_at`

	var pl domain.PackingList
	var desc sql.NullString
	err = s.db.QueryRowContext(ctx, query, name, description, id).Scan(
		&pl.ID, &pl.Name, &desc, &pl.CreatedAt, &pl.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	pl.Description = strPtr(desc)
	return &pl, nil
}

// Delete deletes a packing list
func (s *PackingListStore) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM packing_lists WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return rowsAffectedOrNotFound(result, "delete packing list")
}

// ListLabels returns all labels for a packing list
func (s *PackingListStore) ListLabels(ctx context.Context, packingListID uuid.UUID) ([]domain.Label, error) {
	query := `
		SELECT l.id, l.name, l.color, l.created_at, l.updated_at
		FROM labels l
		JOIN packing_list_labels pll ON l.id = pll.label_id
		WHERE pll.packing_list_id = $1
		ORDER BY l.name ASC`

	rows, err := s.db.QueryContext(ctx, query, packingListID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []domain.Label
	for rows.Next() {
		var l domain.Label
		var color sql.NullString
		if err := rows.Scan(&l.ID, &l.Name, &color, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		l.Color = strPtr(color)
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

// AddLabel adds a label to a packing list
func (s *PackingListStore) AddLabel(ctx context.Context, packingListID, labelID uuid.UUID) error {
	// Verify packing list exists
	if _, err := s.GetByID(ctx, packingListID); err != nil {
		return err
	}

	query := `
		INSERT INTO packing_list_labels (packing_list_id, label_id)
		VALUES ($1, $2)
		ON CONFLICT (packing_list_id, label_id) DO NOTHING`

	_, err := s.db.ExecContext(ctx, query, packingListID, labelID)
	return err
}

// RemoveLabel removes a label from a packing list
func (s *PackingListStore) RemoveLabel(ctx context.Context, packingListID, labelID uuid.UUID) error {
	// Verify packing list exists
	if _, err := s.GetByID(ctx, packingListID); err != nil {
		return err
	}

	query := `
		DELETE FROM packing_list_labels
		WHERE packing_list_id = $1 AND label_id = $2`

	result, err := s.db.ExecContext(ctx, query, packingListID, labelID)
	if err != nil {
		return err
	}

	return rowsAffectedOrNotFound(result, "remove packing list label")
}
