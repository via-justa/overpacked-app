package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

type LabelStore struct {
	db *sql.DB
}

func NewLabelStore(db *sql.DB) *LabelStore {
	return &LabelStore{db: db}
}

func (s *LabelStore) Create(ctx context.Context, label *domain.Label) error {
	query := `
		INSERT INTO labels (name, color)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	if err := s.db.QueryRowContext(ctx, query, label.Name, toNullString(label.Color)).Scan(
		&label.ID,
		&label.CreatedAt,
		&label.UpdatedAt,
	); err != nil {
		return fmt.Errorf("create label: %w", err)
	}

	return nil
}

func (s *LabelStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Label, error) {
	query := `
		SELECT id, name, color, created_at, updated_at
		FROM labels
		WHERE id = $1`

	var label domain.Label
	var color sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&label.ID,
		&label.Name,
		&color,
		&label.CreatedAt,
		&label.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get label by id: %w", err)
	}

	label.Color = strPtr(color)
	return &label, nil
}

func (s *LabelStore) List(ctx context.Context) ([]domain.Label, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, color, created_at, updated_at
		FROM labels
		ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list labels: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Label, 0)
	for rows.Next() {
		var l domain.Label
		var color sql.NullString
		if err := rows.Scan(&l.ID, &l.Name, &color, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan label: %w", err)
		}
		l.Color = strPtr(color)
		out = append(out, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate labels: %w", err)
	}

	return out, nil
}

func (s *LabelStore) Update(ctx context.Context, label *domain.Label) error {
	query := `
		UPDATE labels
		SET name = $2,
			color = $3,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	if err := s.db.QueryRowContext(ctx, query, label.ID, label.Name, toNullString(label.Color)).Scan(&label.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("update label: %w", err)
	}

	return nil
}

func (s *LabelStore) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM labels WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete label: %w", err)
	}

	return rowsAffectedOrNotFound(result, "check delete label result")
}

// ListItemLabels returns all labels for a given item.
func (s *LabelStore) ListItemLabels(ctx context.Context, itemID uuid.UUID) ([]domain.Label, error) {
	query := `
		SELECT l.id, l.name, l.color, l.created_at, l.updated_at
		FROM labels l
		JOIN item_labels il ON l.id = il.label_id
		WHERE il.item_id = $1
		ORDER BY l.name ASC`

	rows, err := s.db.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, fmt.Errorf("list item labels: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Label, 0)
	for rows.Next() {
		var l domain.Label
		var color sql.NullString
		if err := rows.Scan(&l.ID, &l.Name, &color, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan item label: %w", err)
		}
		l.Color = strPtr(color)
		out = append(out, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate item labels: %w", err)
	}

	return out, nil
}

// AddItemLabel adds a label to an item.
func (s *LabelStore) AddItemLabel(ctx context.Context, itemID, labelID uuid.UUID) error {
	query := `
		INSERT INTO item_labels (item_id, label_id)
		VALUES ($1, $2)
		ON CONFLICT (item_id, label_id) DO NOTHING`

	if _, err := s.db.ExecContext(ctx, query, itemID, labelID); err != nil {
		return fmt.Errorf("add item label: %w", err)
	}

	return nil
}

// RemoveItemLabel removes a label from an item.
func (s *LabelStore) RemoveItemLabel(ctx context.Context, itemID, labelID uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM item_labels
		WHERE item_id = $1 AND label_id = $2`, itemID, labelID)
	if err != nil {
		return fmt.Errorf("remove item label: %w", err)
	}

	return rowsAffectedOrNotFound(result, "check remove item label result")
}
