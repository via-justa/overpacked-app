package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

type ManufacturerStore struct {
	db *sql.DB
}

func NewManufacturerStore(db *sql.DB) *ManufacturerStore {
	return &ManufacturerStore{db: db}
}

func (s *ManufacturerStore) Create(ctx context.Context, manufacturer *domain.Manufacturer) error {
	query := `
		INSERT INTO manufacturers (name, website)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	if err := s.db.QueryRowContext(ctx, query, manufacturer.Name, toNullString(manufacturer.Website)).Scan(
		&manufacturer.ID,
		&manufacturer.CreatedAt,
		&manufacturer.UpdatedAt,
	); err != nil {
		return fmt.Errorf("create manufacturer: %w", err)
	}

	return nil
}

func (s *ManufacturerStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Manufacturer, error) {
	query := `
		SELECT id, name, website, created_at, updated_at
		FROM manufacturers
		WHERE id = $1`

	var manufacturer domain.Manufacturer
	var website sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&manufacturer.ID,
		&manufacturer.Name,
		&website,
		&manufacturer.CreatedAt,
		&manufacturer.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get manufacturer by id: %w", err)
	}

	manufacturer.Website = strPtr(website)
	return &manufacturer, nil
}

func (s *ManufacturerStore) List(ctx context.Context) ([]domain.Manufacturer, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, website, created_at, updated_at
		FROM manufacturers
		ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list manufacturers: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Manufacturer, 0)
	for rows.Next() {
		var m domain.Manufacturer
		var website sql.NullString
		if err := rows.Scan(&m.ID, &m.Name, &website, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan manufacturer: %w", err)
		}
		m.Website = strPtr(website)
		out = append(out, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate manufacturers: %w", err)
	}

	return out, nil
}

func (s *ManufacturerStore) Update(ctx context.Context, manufacturer *domain.Manufacturer) error {
	query := `
		UPDATE manufacturers
		SET name = $2,
			website = $3,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := s.db.QueryRowContext(ctx, query, manufacturer.ID, manufacturer.Name, toNullString(manufacturer.Website)).Scan(&manufacturer.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update manufacturer: %w", err)
	}

	return nil
}

func (s *ManufacturerStore) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM manufacturers WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete manufacturer: %w", err)
	}

	return rowsAffectedOrNotFound(result, "rows affected on delete manufacturer")
}
