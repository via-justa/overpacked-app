package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

type PersonStore struct {
	db *sql.DB
}

func NewPersonStore(db *sql.DB) *PersonStore {
	return &PersonStore{db: db}
}

func (s *PersonStore) Create(ctx context.Context, person *domain.Person) error {
	query := `
		INSERT INTO persons (name, gender, birthdate, body_weight_grams, conditioning_level)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	gender := sql.NullString{}
	if person.Gender != nil {
		gender = sql.NullString{String: string(*person.Gender), Valid: true}
	}

	var birthdate sql.NullTime
	if person.Birthdate != nil {
		birthdate = sql.NullTime{Time: *person.Birthdate, Valid: true}
	}

	conditioningLevel := sql.NullString{}
	if person.ConditioningLevel != nil {
		conditioningLevel = sql.NullString{String: string(*person.ConditioningLevel), Valid: true}
	}

	if err := s.db.QueryRowContext(ctx, query, person.Name, gender, birthdate, toNullFloat64(person.BodyWeightGrams), conditioningLevel).Scan(
		&person.ID,
		&person.CreatedAt,
		&person.UpdatedAt,
	); err != nil {
		return fmt.Errorf("create person: %w", err)
	}

	return nil
}

func (s *PersonStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Person, error) {
	query := `
		SELECT id, name, gender, birthdate, body_weight_grams, conditioning_level, created_at, updated_at
		FROM persons
		WHERE id = $1`

	var person domain.Person
	var gender sql.NullString
	var birthdate sql.NullTime
	var bodyWeight sql.NullFloat64
	var conditioningLevel sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&person.ID,
		&person.Name,
		&gender,
		&birthdate,
		&bodyWeight,
		&conditioningLevel,
		&person.CreatedAt,
		&person.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get person by id: %w", err)
	}

	if gender.Valid {
		g := domain.Gender(gender.String)
		person.Gender = &g
	}
	if conditioningLevel.Valid {
		cl := domain.ConditioningLevel(conditioningLevel.String)
		person.ConditioningLevel = &cl
	}
	person.Birthdate = timePtr(birthdate)
	person.BodyWeightGrams = floatPtr(bodyWeight)

	return &person, nil
}

func (s *PersonStore) List(ctx context.Context) ([]domain.Person, error) {
	query := `
		SELECT id, name, gender, birthdate, body_weight_grams, conditioning_level, created_at, updated_at
		FROM persons
		ORDER BY name ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list persons: %w", err)
	}
	defer rows.Close()

	persons := make([]domain.Person, 0)
	for rows.Next() {
		var person domain.Person
		var gender sql.NullString
		var birthdate sql.NullTime
		var bodyWeight sql.NullFloat64
		var conditioningLevel sql.NullString

		if err := rows.Scan(
			&person.ID,
			&person.Name,
			&gender,
			&birthdate,
			&bodyWeight,
			&conditioningLevel,
			&person.CreatedAt,
			&person.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan person: %w", err)
		}

		if gender.Valid {
			g := domain.Gender(gender.String)
			person.Gender = &g
		}
		if conditioningLevel.Valid {
			cl := domain.ConditioningLevel(conditioningLevel.String)
			person.ConditioningLevel = &cl
		}
		person.Birthdate = timePtr(birthdate)
		person.BodyWeightGrams = floatPtr(bodyWeight)

		persons = append(persons, person)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate persons: %w", err)
	}

	return persons, nil
}

func (s *PersonStore) Update(ctx context.Context, person *domain.Person) error {
	query := `
		UPDATE persons
		SET name = $2,
			gender = $3,
			birthdate = $4,
			body_weight_grams = $5,
			conditioning_level = $6,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	gender := sql.NullString{}
	if person.Gender != nil {
		gender = sql.NullString{String: string(*person.Gender), Valid: true}
	}

	var birthdate sql.NullTime
	if person.Birthdate != nil {
		birthdate = sql.NullTime{Time: *person.Birthdate, Valid: true}
	}

	conditioningLevel := sql.NullString{}
	if person.ConditioningLevel != nil {
		conditioningLevel = sql.NullString{String: string(*person.ConditioningLevel), Valid: true}
	}

	err := s.db.QueryRowContext(ctx, query, person.ID, person.Name, gender, birthdate, toNullFloat64(person.BodyWeightGrams), conditioningLevel).Scan(&person.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update person: %w", err)
	}

	return nil
}

func (s *PersonStore) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM persons WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete person: %w", err)
	}

	return rowsAffectedOrNotFound(result, "rows affected on delete person")
}
