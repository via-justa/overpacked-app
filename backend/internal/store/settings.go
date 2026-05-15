package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

const singletonSettingsID = 1

type SettingsStore struct {
	db *sql.DB
}

func NewSettingsStore(db *sql.DB) *SettingsStore {
	return &SettingsStore{db: db}
}

func (s *SettingsStore) Get(ctx context.Context) (*domain.Settings, error) {
	query := `
		SELECT id, weight_unit, distance_unit, temperature_unit, volume_unit, currency
		FROM settings
		WHERE id = $1`

	var settings domain.Settings
	var weightUnit string
	var distanceUnit string
	var temperatureUnit string
	var volumeUnit string
	var currency string

	err := s.db.QueryRowContext(ctx, query, singletonSettingsID).Scan(
		&settings.ID,
		&weightUnit,
		&distanceUnit,
		&temperatureUnit,
		&volumeUnit,
		&currency,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}

	settings.WeightUnit = domain.WeightUnit(weightUnit)
	settings.DistanceUnit = domain.DistanceUnit(distanceUnit)
	settings.TemperatureUnit = domain.TemperatureUnit(temperatureUnit)
	settings.VolumeUnit = domain.VolumeUnit(volumeUnit)
	settings.Currency = domain.Currency(currency)

	return &settings, nil
}

func (s *SettingsStore) Update(ctx context.Context, settings *domain.Settings) error {
	query := `
		UPDATE settings
		SET weight_unit = $2,
			distance_unit = $3,
			temperature_unit = $4,
			volume_unit = $5,
			currency = $6
		WHERE id = $1`

	result, err := s.db.ExecContext(
		ctx,
		query,
		singletonSettingsID,
		string(settings.WeightUnit),
		string(settings.DistanceUnit),
		string(settings.TemperatureUnit),
		string(settings.VolumeUnit),
		string(settings.Currency),
	)
	if err != nil {
		return fmt.Errorf("update settings: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected on update settings: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	settings.ID = singletonSettingsID
	return nil
}

func (s *SettingsStore) StartFresh(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin start fresh tx: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Remove all user data while preserving system item types and schema rows.
	if _, err = tx.ExecContext(ctx, `
		TRUNCATE TABLE
			trip_persons,
			trip_sets,
			trip_items,
			trip_packs,
			trips,
			pack_items,
			packs,
			set_items,
			item_sets,
			item_labels,
			items,
			packing_list_labels,
			packing_lists,
			labels,
			manufacturers,
			persons
		RESTART IDENTITY CASCADE
	`); err != nil {
		return fmt.Errorf("truncate app data: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM item_types WHERE is_system = FALSE`); err != nil {
		return fmt.Errorf("delete custom item types: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE settings
		SET weight_unit = 'g',
			distance_unit = 'km',
			temperature_unit = 'c',
			volume_unit = 'ml',
			currency = 'eur'
		WHERE id = $1
	`, singletonSettingsID); err != nil {
		return fmt.Errorf("reset settings defaults: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit start fresh tx: %w", err)
	}

	return nil
}
