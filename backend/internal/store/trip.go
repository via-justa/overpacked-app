package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

type TripStore struct {
	db *sql.DB
}

func NewTripStore(db *sql.DB) *TripStore {
	return &TripStore{db: db}
}

const errGetRowsAffected = "get rows affected: %w"

// Trip CRUD operations

func (s *TripStore) Create(ctx context.Context, trip *domain.Trip) error {
	query := `
		INSERT INTO trips (name, trip_type, duration, notes, trip_komoot_url, trip_strava_url, trip_wanderer_url, total_distance_km)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(
		ctx,
		query,
		trip.Name,
		trip.TripType,
		toNullString(trip.Duration),
		toNullString(trip.Notes),
		toNullString(trip.TripKomootURL),
		toNullString(trip.TripStravaURL),
		toNullString(trip.TripWandererURL),
		toNullFloat64(trip.TotalDistanceKm),
	).Scan(&trip.ID, &trip.CreatedAt, &trip.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create trip: %w", err)
	}

	return nil
}

func (s *TripStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Trip, error) {
	query := `
		SELECT id, name, trip_type, duration, notes, trip_komoot_url, trip_strava_url, trip_wanderer_url, total_distance_km, created_at, updated_at
		FROM trips
		WHERE id = $1`

	var trip domain.Trip
	var duration, notes, komootURL, stravaURL, wandererURL sql.NullString
	var distanceKm sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&trip.ID,
		&trip.Name,
		&trip.TripType,
		&duration,
		&notes,
		&komootURL,
		&stravaURL,
		&wandererURL,
		&distanceKm,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get trip by id: %w", err)
	}

	trip.Duration = strPtr(duration)
	trip.Notes = strPtr(notes)
	trip.TripKomootURL = strPtr(komootURL)
	trip.TripStravaURL = strPtr(stravaURL)
	trip.TripWandererURL = strPtr(wandererURL)
	trip.TotalDistanceKm = floatPtr(distanceKm)

	return &trip, nil
}

func (s *TripStore) List(ctx context.Context) ([]domain.Trip, error) {
	query := `
		SELECT id, name, trip_type, duration, notes, trip_komoot_url, trip_strava_url, trip_wanderer_url, total_distance_km, created_at, updated_at
		FROM trips
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list trips: %w", err)
	}
	defer rows.Close()

	var trips []domain.Trip
	for rows.Next() {
		var trip domain.Trip
		var duration, notes, komootURL, stravaURL, wandererURL sql.NullString
		var distanceKm sql.NullFloat64

		err := rows.Scan(
			&trip.ID,
			&trip.Name,
			&trip.TripType,
			&duration,
			&notes,
			&komootURL,
			&stravaURL,
			&wandererURL,
			&distanceKm,
			&trip.CreatedAt,
			&trip.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan trip row: %w", err)
		}

		trip.Duration = strPtr(duration)
		trip.Notes = strPtr(notes)
		trip.TripKomootURL = strPtr(komootURL)
		trip.TripStravaURL = strPtr(stravaURL)
		trip.TripWandererURL = strPtr(wandererURL)
		trip.TotalDistanceKm = floatPtr(distanceKm)

		trips = append(trips, trip)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate trip rows: %w", err)
	}

	return trips, nil
}

func (s *TripStore) Update(ctx context.Context, trip *domain.Trip) error {
	query := `
		UPDATE trips
		SET name = $1, trip_type = $2, duration = $3, notes = $4, trip_komoot_url = $5, trip_strava_url = $6, trip_wanderer_url = $7, total_distance_km = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9
		RETURNING updated_at`

	err := s.db.QueryRowContext(
		ctx,
		query,
		trip.Name,
		trip.TripType,
		toNullString(trip.Duration),
		toNullString(trip.Notes),
		toNullString(trip.TripKomootURL),
		toNullString(trip.TripStravaURL),
		toNullString(trip.TripWandererURL),
		toNullFloat64(trip.TotalDistanceKm),
		trip.ID,
	).Scan(&trip.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update trip: %w", err)
	}

	return nil
}

func (s *TripStore) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM trips WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete trip: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(errGetRowsAffected, err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// TripPerson operations

func (s *TripStore) AddPerson(ctx context.Context, tripPerson *domain.TripPerson) error {
	query := `
		INSERT INTO trip_persons (trip_id, person_id)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(ctx, query, tripPerson.TripID, tripPerson.PersonID).Scan(
		&tripPerson.ID,
		&tripPerson.CreatedAt,
		&tripPerson.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("add person to trip: %w", err)
	}

	return nil
}

func (s *TripStore) GetTripPersonID(ctx context.Context, tripID, personID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT id FROM trip_persons WHERE trip_id = $1 AND person_id = $2`

	var tripPersonID uuid.UUID
	err := s.db.QueryRowContext(ctx, query, tripID, personID).Scan(&tripPersonID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, domain.ErrNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("get trip person id: %w", err)
	}

	return tripPersonID, nil
}

func (s *TripStore) ListPersons(ctx context.Context, tripID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT person_id
		FROM trip_persons
		WHERE trip_id = $1`

	rows, err := s.db.QueryContext(ctx, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("list trip persons: %w", err)
	}
	defer rows.Close()

	var personIDs []uuid.UUID
	for rows.Next() {
		var personID uuid.UUID
		if err := rows.Scan(&personID); err != nil {
			return nil, fmt.Errorf("scan person id: %w", err)
		}
		personIDs = append(personIDs, personID)
	}

	return personIDs, rows.Err()
}

func (s *TripStore) RemovePerson(ctx context.Context, tripID, personID uuid.UUID) error {
	query := `DELETE FROM trip_persons WHERE trip_id = $1 AND person_id = $2`

	result, err := s.db.ExecContext(ctx, query, tripID, personID)
	if err != nil {
		return fmt.Errorf("remove person from trip: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(errGetRowsAffected, err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// TripPersonPack operations

func (s *TripStore) AddPersonPack(ctx context.Context, tripPersonPack *domain.TripPersonPack) error {
	query := `
		INSERT INTO trip_person_packs (trip_person_id, pack_id)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(ctx, query, tripPersonPack.TripPersonID, tripPersonPack.PackID).Scan(
		&tripPersonPack.ID,
		&tripPersonPack.CreatedAt,
		&tripPersonPack.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("add pack to person in trip: %w", err)
	}

	return nil
}

func (s *TripStore) ListPersonPacks(ctx context.Context, tripPersonID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT pack_id
		FROM trip_person_packs
		WHERE trip_person_id = $1`

	rows, err := s.db.QueryContext(ctx, query, tripPersonID)
	if err != nil {
		return nil, fmt.Errorf("list person packs: %w", err)
	}
	defer rows.Close()

	var packIDs []uuid.UUID
	for rows.Next() {
		var packID uuid.UUID
		if err := rows.Scan(&packID); err != nil {
			return nil, fmt.Errorf("scan pack id: %w", err)
		}
		packIDs = append(packIDs, packID)
	}

	return packIDs, rows.Err()
}

func (s *TripStore) RemovePersonPack(ctx context.Context, tripPersonID, packID uuid.UUID) error {
	query := `DELETE FROM trip_person_packs WHERE trip_person_id = $1 AND pack_id = $2`

	result, err := s.db.ExecContext(ctx, query, tripPersonID, packID)
	if err != nil {
		return fmt.Errorf("remove pack from person: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(errGetRowsAffected, err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// TripPersonItem operations

func (s *TripStore) AddPersonItem(ctx context.Context, tripPersonItem *domain.TripPersonItem) error {
	query := `
		INSERT INTO trip_person_items (trip_person_id, item_id, quantity, carry_status, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(
		ctx,
		query,
		tripPersonItem.TripPersonID,
		tripPersonItem.ItemID,
		tripPersonItem.Quantity,
		tripPersonItem.CarryStatus,
		toNullString(tripPersonItem.Notes),
	).Scan(&tripPersonItem.ID, &tripPersonItem.CreatedAt, &tripPersonItem.UpdatedAt)
	if err != nil {
		return fmt.Errorf("add item to person in trip: %w", err)
	}

	return nil
}

func (s *TripStore) GetPersonItemByID(ctx context.Context, id uuid.UUID) (*domain.TripPersonItem, error) {
	query := `
		SELECT id, trip_person_id, item_id, quantity, carry_status, notes, created_at, updated_at
		FROM trip_person_items
		WHERE id = $1`

	var tripPersonItem domain.TripPersonItem
	var notes sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&tripPersonItem.ID,
		&tripPersonItem.TripPersonID,
		&tripPersonItem.ItemID,
		&tripPersonItem.Quantity,
		&tripPersonItem.CarryStatus,
		&notes,
		&tripPersonItem.CreatedAt,
		&tripPersonItem.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get person item: %w", err)
	}

	tripPersonItem.Notes = strPtr(notes)

	return &tripPersonItem, nil
}

func (s *TripStore) ListPersonItems(ctx context.Context, tripPersonID uuid.UUID) ([]domain.TripPersonItem, error) {
	query := `
		SELECT id, trip_person_id, item_id, quantity, carry_status, notes, created_at, updated_at
		FROM trip_person_items
		WHERE trip_person_id = $1
		ORDER BY created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, tripPersonID)
	if err != nil {
		return nil, fmt.Errorf("list person items: %w", err)
	}
	defer rows.Close()

	var tripPersonItems []domain.TripPersonItem
	for rows.Next() {
		var tripPersonItem domain.TripPersonItem
		var notes sql.NullString

		err := rows.Scan(
			&tripPersonItem.ID,
			&tripPersonItem.TripPersonID,
			&tripPersonItem.ItemID,
			&tripPersonItem.Quantity,
			&tripPersonItem.CarryStatus,
			&notes,
			&tripPersonItem.CreatedAt,
			&tripPersonItem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan person item: %w", err)
		}

		tripPersonItem.Notes = strPtr(notes)
		tripPersonItems = append(tripPersonItems, tripPersonItem)
	}

	return tripPersonItems, rows.Err()
}

func (s *TripStore) UpdatePersonItem(ctx context.Context, tripPersonItem *domain.TripPersonItem) error {
	query := `
		UPDATE trip_person_items
		SET quantity = $1, carry_status = $2, notes = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING updated_at`

	err := s.db.QueryRowContext(
		ctx,
		query,
		tripPersonItem.Quantity,
		tripPersonItem.CarryStatus,
		toNullString(tripPersonItem.Notes),
		tripPersonItem.ID,
	).Scan(&tripPersonItem.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update person item: %w", err)
	}

	return nil
}

func (s *TripStore) RemovePersonItem(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM trip_person_items WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("remove item from person: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(errGetRowsAffected, err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
