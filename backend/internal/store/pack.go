package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

type PackStore struct {
	db *sql.DB
}

func NewPackStore(db *sql.DB) *PackStore {
	return &PackStore{db: db}
}

func (s *PackStore) Create(ctx context.Context, pack *domain.Pack) error {
	query := `
		INSERT INTO packs (person_id, name, trip_type, notes, is_template)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	var personID any
	if pack.PersonID != nil {
		personID = *pack.PersonID
	}

	var tripType sql.NullString
	if pack.TripType != nil {
		tripType = sql.NullString{String: string(*pack.TripType), Valid: true}
	}

	err := s.db.QueryRowContext(
		ctx,
		query,
		personID,
		pack.Name,
		tripType,
		toNullString(pack.Notes),
		pack.IsTemplate,
	).Scan(&pack.ID, &pack.CreatedAt, &pack.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create pack: %w", err)
	}

	return nil
}

func (s *PackStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Pack, error) {
	query := `
		SELECT id, person_id, name, trip_type, notes, is_template, created_at, updated_at
		FROM packs
		WHERE id = $1`

	var pack domain.Pack
	var personID sql.NullString
	var tripType sql.NullString
	var notes sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&pack.ID,
		&personID,
		&pack.Name,
		&tripType,
		&notes,
		&pack.IsTemplate,
		&pack.CreatedAt,
		&pack.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get pack by id: %w", err)
	}

	if personID.Valid {
		parsed, parseErr := uuid.Parse(personID.String)
		if parseErr != nil {
			return nil, fmt.Errorf("parse pack person_id: %w", parseErr)
		}
		pack.PersonID = &parsed
	}
	if tripType.Valid {
		t := domain.TripType(tripType.String)
		pack.TripType = &t
	}
	pack.Notes = strPtr(notes)

	return &pack, nil
}

func (s *PackStore) List(ctx context.Context) ([]domain.Pack, error) {
	query := `
		SELECT id, person_id, name, trip_type, notes, is_template, created_at, updated_at
		FROM packs
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list packs: %w", err)
	}
	defer rows.Close()

	packs := make([]domain.Pack, 0)
	for rows.Next() {
		var pack domain.Pack
		var personID sql.NullString
		var tripType sql.NullString
		var notes sql.NullString

		if err := rows.Scan(
			&pack.ID,
			&personID,
			&pack.Name,
			&tripType,
			&notes,
			&pack.IsTemplate,
			&pack.CreatedAt,
			&pack.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pack: %w", err)
		}

		if personID.Valid {
			parsed, parseErr := uuid.Parse(personID.String)
			if parseErr != nil {
				return nil, fmt.Errorf("parse pack person_id: %w", parseErr)
			}
			pack.PersonID = &parsed
		}
		if tripType.Valid {
			t := domain.TripType(tripType.String)
			pack.TripType = &t
		}
		pack.Notes = strPtr(notes)
		packs = append(packs, pack)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate packs: %w", err)
	}

	return packs, nil
}

func (s *PackStore) Update(ctx context.Context, pack *domain.Pack) error {
	query := `
		UPDATE packs
		SET person_id = $2,
			name = $3,
			trip_type = $4,
			notes = $5,
			is_template = $6,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	var personID any
	if pack.PersonID != nil {
		personID = *pack.PersonID
	}

	var tripType sql.NullString
	if pack.TripType != nil {
		tripType = sql.NullString{String: string(*pack.TripType), Valid: true}
	}

	err := s.db.QueryRowContext(
		ctx,
		query,
		pack.ID,
		personID,
		pack.Name,
		tripType,
		toNullString(pack.Notes),
		pack.IsTemplate,
	).Scan(&pack.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update pack: %w", err)
	}

	return nil
}

func (s *PackStore) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM packs WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete pack: %w", err)
	}

	return rowsAffectedOrNotFound(result, "rows affected on delete pack")
}

func (s *PackStore) AddItem(ctx context.Context, item *domain.PackItem) error {
	query := `
		INSERT INTO pack_items (pack_id, item_id, quantity, carry_status, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(
		ctx,
		query,
		item.PackID,
		item.ItemID,
		item.Quantity,
		string(item.CarryStatus),
		toNullString(item.Notes),
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("add pack item: %w", err)
	}

	return nil
}

func (s *PackStore) ListItems(ctx context.Context, packID uuid.UUID) ([]domain.PackItem, error) {
	query := `
		SELECT id, pack_id, item_id, quantity, carry_status, notes, created_at, updated_at
		FROM pack_items
		WHERE pack_id = $1
		ORDER BY created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, packID)
	if err != nil {
		return nil, fmt.Errorf("list pack items: %w", err)
	}
	defer rows.Close()

	out := make([]domain.PackItem, 0)
	for rows.Next() {
		var item domain.PackItem
		var carryStatus string
		var notes sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.PackID,
			&item.ItemID,
			&item.Quantity,
			&carryStatus,
			&notes,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pack item: %w", err)
		}

		item.CarryStatus = domain.CarryStatus(carryStatus)
		item.Notes = strPtr(notes)
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pack items: %w", err)
	}

	return out, nil
}

func (s *PackStore) GetItemByID(ctx context.Context, packID uuid.UUID, itemID uuid.UUID) (*domain.PackItem, error) {
	query := `
		SELECT id, pack_id, item_id, quantity, carry_status, notes, created_at, updated_at
		FROM pack_items
		WHERE pack_id = $1 AND item_id = $2`

	var item domain.PackItem
	var carryStatus string
	var notes sql.NullString

	err := s.db.QueryRowContext(ctx, query, packID, itemID).Scan(
		&item.ID,
		&item.PackID,
		&item.ItemID,
		&item.Quantity,
		&carryStatus,
		&notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get pack item: %w", err)
	}

	item.CarryStatus = domain.CarryStatus(carryStatus)
	item.Notes = strPtr(notes)

	return &item, nil
}

func (s *PackStore) UpdateItem(ctx context.Context, item *domain.PackItem) error {
	query := `
		UPDATE pack_items
		SET quantity = $3,
			carry_status = $4,
			notes = $5,
			updated_at = NOW()
		WHERE pack_id = $1 AND item_id = $2
		RETURNING updated_at`

	err := s.db.QueryRowContext(
		ctx,
		query,
		item.PackID,
		item.ItemID,
		item.Quantity,
		string(item.CarryStatus),
		toNullString(item.Notes),
	).Scan(&item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update pack item: %w", err)
	}

	return nil
}

func (s *PackStore) RemoveItem(ctx context.Context, packID uuid.UUID, itemID uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM pack_items WHERE pack_id = $1 AND item_id = $2", packID, itemID)
	if err != nil {
		return fmt.Errorf("remove pack item: %w", err)
	}

	return rowsAffectedOrNotFound(result, "rows affected on remove pack item")
}
