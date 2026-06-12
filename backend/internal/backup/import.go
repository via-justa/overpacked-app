package backup

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/google/uuid"
)

// Import restores an archive into the database in a single transaction.
// In ModeReplace it first wipes user content (preserving the manufacturers/labels
// catalog and system item types); in ModeMerge it upserts on top of existing data.
func (s *Service) Import(ctx context.Context, zr *zip.Reader, mode Mode) (ImportResult, error) {
	if !mode.valid() {
		return ImportResult{}, ErrInvalidMode
	}

	snap, _, images, err := readArchive(zr)
	if err != nil {
		return ImportResult{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ImportResult{}, fmt.Errorf("begin import tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if mode == ModeReplace {
		if err = wipeUserContent(ctx, tx); err != nil {
			return ImportResult{}, err
		}
	}

	if err = applySnapshot(ctx, tx, snap, images); err != nil {
		return ImportResult{}, err
	}

	if err = tx.Commit(); err != nil {
		return ImportResult{}, fmt.Errorf("commit import tx: %w", err)
	}

	return ImportResult{Mode: mode, Counts: snap.counts()}, nil
}

func readArchive(zr *zip.Reader) (*Snapshot, *Manifest, map[string][]byte, error) {
	var snap *Snapshot
	var manifest *Manifest
	images := map[string][]byte{}

	for _, f := range zr.File {
		switch {
		case f.Name == dataFilename:
			s, err := readJSONEntry[Snapshot](f)
			if err != nil {
				return nil, nil, nil, err
			}
			snap = s
		case f.Name == manifestFilename:
			m, err := readJSONEntry[Manifest](f)
			if err != nil {
				return nil, nil, nil, err
			}
			manifest = m
		case strings.HasPrefix(f.Name, imagesDir+"/") && !f.FileInfo().IsDir():
			rc, err := f.Open()
			if err != nil {
				return nil, nil, nil, fmt.Errorf("%w: open %s", ErrInvalidArchive, f.Name)
			}
			data, err := io.ReadAll(rc)
			_ = rc.Close()
			if err != nil {
				return nil, nil, nil, fmt.Errorf("%w: read %s", ErrInvalidArchive, f.Name)
			}
			images[path.Base(f.Name)] = data
		}
	}

	if snap == nil {
		return nil, nil, nil, fmt.Errorf("%w: missing %s", ErrInvalidArchive, dataFilename)
	}
	if manifest != nil && manifest.FormatVersion > FormatVersion {
		return nil, nil, nil, fmt.Errorf("%w: archive is version %d, this build supports %d",
			ErrUnsupportedVersion, manifest.FormatVersion, FormatVersion)
	}

	return snap, manifest, images, nil
}

func readJSONEntry[T any](f *zip.File) (*T, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("%w: open %s", ErrInvalidArchive, f.Name)
	}
	defer rc.Close()

	var out T
	if err := json.NewDecoder(rc).Decode(&out); err != nil {
		return nil, fmt.Errorf("%w: decode %s: %v", ErrInvalidArchive, f.Name, err)
	}
	return &out, nil
}

func wipeUserContent(ctx context.Context, tx *sql.Tx) error {
	// User content only: the manufacturers/labels catalog and system item types are preserved.
	if _, err := tx.ExecContext(ctx, `
		TRUNCATE TABLE
			trip_person_items, trip_person_packs, trip_persons, trips,
			pack_items, packs,
			set_items, item_sets,
			item_labels, packing_list_labels,
			items, packing_lists, persons
		CASCADE`); err != nil {
		return fmt.Errorf("wipe user content: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM item_types WHERE is_system = FALSE`); err != nil {
		return fmt.Errorf("wipe custom item types: %w", err)
	}
	return nil
}

// applySnapshot inserts/upserts the snapshot in FK dependency order.
func applySnapshot(ctx context.Context, tx *sql.Tx, snap *Snapshot, images map[string][]byte) error {
	if err := applySettings(ctx, tx, snap.Settings); err != nil {
		return err
	}

	manuMap, err := applyManufacturers(ctx, tx, snap.Manufacturers)
	if err != nil {
		return err
	}
	labelMap, err := applyLabels(ctx, tx, snap.Labels)
	if err != nil {
		return err
	}
	if err := applyItemTypes(ctx, tx, snap.ItemTypes); err != nil {
		return err
	}
	if err := applyItemTypeFields(ctx, tx, snap.ItemTypeFields); err != nil {
		return err
	}
	if err := applyPersons(ctx, tx, snap.Persons); err != nil {
		return err
	}
	if err := applyItems(ctx, tx, snap.Items, manuMap, images); err != nil {
		return err
	}
	if err := applyItemLabels(ctx, tx, snap.ItemLabels, labelMap); err != nil {
		return err
	}
	if err := applySets(ctx, tx, snap.Sets); err != nil {
		return err
	}
	if err := applySetItems(ctx, tx, snap.SetItems); err != nil {
		return err
	}
	if err := applyPacks(ctx, tx, snap.Packs); err != nil {
		return err
	}
	if err := applyPackItems(ctx, tx, snap.PackItems); err != nil {
		return err
	}
	if err := applyPackingLists(ctx, tx, snap.PackingLists); err != nil {
		return err
	}
	if err := applyPackingListLabels(ctx, tx, snap.PackingListLabels, labelMap); err != nil {
		return err
	}
	if err := applyTrips(ctx, tx, snap.Trips); err != nil {
		return err
	}
	tripPersonMap, err := applyTripPersons(ctx, tx, snap.TripPersons)
	if err != nil {
		return err
	}
	if err := applyTripPersonPacks(ctx, tx, snap.TripPersonPacks, tripPersonMap); err != nil {
		return err
	}
	return applyTripPersonItems(ctx, tx, snap.TripPersonItems, tripPersonMap)
}

func applySettings(ctx context.Context, tx *sql.Tx, d *settingsDTO) error {
	if d == nil {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		UPDATE settings SET weight_unit=$1, distance_unit=$2, temperature_unit=$3,
			volume_unit=$4, currency=$5 WHERE id = 1`,
		d.WeightUnit, d.DistanceUnit, d.TemperatureUnit, d.VolumeUnit, d.Currency)
	if err != nil {
		return fmt.Errorf("apply settings: %w", err)
	}
	return nil
}

func applyManufacturers(ctx context.Context, tx *sql.Tx, ms []manufacturerDTO) (map[uuid.UUID]uuid.UUID, error) {
	out := make(map[uuid.UUID]uuid.UUID, len(ms))
	for _, d := range ms {
		var resolved uuid.UUID
		err := tx.QueryRowContext(ctx, `
			INSERT INTO manufacturers (id, name, website) VALUES ($1, $2, $3)
			ON CONFLICT (name) DO UPDATE SET website = EXCLUDED.website
			RETURNING id`, d.ID, d.Name, nullStr(d.Website)).Scan(&resolved)
		if err != nil {
			return nil, fmt.Errorf("apply manufacturer %q: %w", d.Name, err)
		}
		out[d.ID] = resolved
	}
	return out, nil
}

func applyLabels(ctx context.Context, tx *sql.Tx, ls []labelDTO) (map[uuid.UUID]uuid.UUID, error) {
	out := make(map[uuid.UUID]uuid.UUID, len(ls))
	for _, d := range ls {
		var resolved uuid.UUID
		err := tx.QueryRowContext(ctx, `
			INSERT INTO labels (id, name, color) VALUES ($1, $2, $3)
			ON CONFLICT (name) DO UPDATE SET color = EXCLUDED.color
			RETURNING id`, d.ID, d.Name, nullStr(d.Color)).Scan(&resolved)
		if err != nil {
			return nil, fmt.Errorf("apply label %q: %w", d.Name, err)
		}
		out[d.ID] = resolved
	}
	return out, nil
}

func applyItemTypes(ctx context.Context, tx *sql.Tx, ts []itemTypeDTO) error {
	for _, d := range ts {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO item_types (id, name, description, base_profile, is_system)
			VALUES ($1, $2, $3, $4, FALSE)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name,
				description = EXCLUDED.description, base_profile = EXCLUDED.base_profile,
				updated_at = NOW()`,
			d.ID, d.Name, nullStr(d.Description), nullStr(d.BaseProfile))
		if err != nil {
			return fmt.Errorf("apply item type %q: %w", d.ID, err)
		}
	}
	return nil
}

func applyItemTypeFields(ctx context.Context, tx *sql.Tx, fs []itemTypeFieldDTO) error {
	for _, d := range fs {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO item_type_fields (id, item_type_id, field_key, field_label, data_type,
				is_required, enum_options_json, min_value, max_value, unit, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (item_type_id, field_key) DO UPDATE SET field_label = EXCLUDED.field_label,
				data_type = EXCLUDED.data_type, is_required = EXCLUDED.is_required,
				enum_options_json = EXCLUDED.enum_options_json, min_value = EXCLUDED.min_value,
				max_value = EXCLUDED.max_value, unit = EXCLUDED.unit, sort_order = EXCLUDED.sort_order,
				updated_at = NOW()`,
			d.ID, d.ItemTypeID, d.FieldKey, d.FieldLabel, d.DataType, d.IsRequired,
			rawOrNull(d.EnumOptions), nullFloat(d.MinValue), nullFloat(d.MaxValue),
			nullStr(d.Unit), d.SortOrder)
		if err != nil {
			return fmt.Errorf("apply item type field %q.%q: %w", d.ItemTypeID, d.FieldKey, err)
		}
	}
	return nil
}

func applyPersons(ctx context.Context, tx *sql.Tx, ps []personDTO) error {
	for _, d := range ps {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO persons (id, name, gender, birthdate, body_weight_grams, conditioning_level)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, gender = EXCLUDED.gender,
				birthdate = EXCLUDED.birthdate, body_weight_grams = EXCLUDED.body_weight_grams,
				conditioning_level = EXCLUDED.conditioning_level, updated_at = NOW()`,
			d.ID, d.Name, nullStr(d.Gender), nullTime(d.Birthdate),
			nullFloat(d.BodyWeightGrams), nullStr(d.ConditioningLevel))
		if err != nil {
			return fmt.Errorf("apply person %q: %w", d.Name, err)
		}
	}
	return nil
}

func applyItems(ctx context.Context, tx *sql.Tx, items []itemDTO, manuMap map[uuid.UUID]uuid.UUID, images map[string][]byte) error {
	for _, d := range items {
		manufacturerID, ok := manuMap[d.ManufacturerID]
		if !ok {
			return fmt.Errorf("item %q references unknown manufacturer %s", d.Name, d.ManufacturerID)
		}

		var imageBlob []byte
		if d.ImageFile != nil {
			imageBlob = images[*d.ImageFile]
		}

		_, err := tx.ExecContext(ctx, `
			INSERT INTO items (id, manufacturer_id, type_id, name, is_active, description, source_url,
				price, weight_grams, volume_ml, default_quantity, default_carry_status, is_default,
				image_blob, image_mime_type, image_size_bytes, image_width_px, image_height_px, attributes_json)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
			ON CONFLICT (id) DO UPDATE SET manufacturer_id = EXCLUDED.manufacturer_id,
				type_id = EXCLUDED.type_id, name = EXCLUDED.name, is_active = EXCLUDED.is_active,
				description = EXCLUDED.description, source_url = EXCLUDED.source_url, price = EXCLUDED.price,
				weight_grams = EXCLUDED.weight_grams, volume_ml = EXCLUDED.volume_ml,
				default_quantity = EXCLUDED.default_quantity, default_carry_status = EXCLUDED.default_carry_status,
				is_default = EXCLUDED.is_default, image_blob = EXCLUDED.image_blob,
				image_mime_type = EXCLUDED.image_mime_type, image_size_bytes = EXCLUDED.image_size_bytes,
				image_width_px = EXCLUDED.image_width_px, image_height_px = EXCLUDED.image_height_px,
				attributes_json = EXCLUDED.attributes_json, updated_at = NOW()`,
			d.ID, manufacturerID, d.TypeID, d.Name, d.IsActive, nullStr(d.Description),
			nullStr(d.SourceURL), nullFloat(d.Price), nullFloat(d.WeightGrams), nullFloat(d.VolumeML),
			d.DefaultQuantity, d.DefaultCarryStatus, d.IsDefault, blobOrNil(imageBlob),
			nullStr(d.ImageMimeType), nullInt(d.ImageSizeBytes), nullInt(d.ImageWidthPX),
			nullInt(d.ImageHeightPX), attributesOrEmpty(d.Attributes))
		if err != nil {
			return fmt.Errorf("apply item %q: %w", d.Name, err)
		}
	}
	return nil
}

func applyItemLabels(ctx context.Context, tx *sql.Tx, ils []itemLabelDTO, labelMap map[uuid.UUID]uuid.UUID) error {
	for _, d := range ils {
		labelID, ok := labelMap[d.LabelID]
		if !ok {
			continue // label not in archive; skip dangling reference
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO item_labels (item_id, label_id) VALUES ($1, $2)
			ON CONFLICT (item_id, label_id) DO NOTHING`, d.ItemID, labelID)
		if err != nil {
			return fmt.Errorf("apply item label: %w", err)
		}
	}
	return nil
}

func applySets(ctx context.Context, tx *sql.Tx, sets []setDTO) error {
	for _, d := range sets {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO item_sets (id, name, description, is_active, set_category)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description,
				is_active = EXCLUDED.is_active, set_category = EXCLUDED.set_category, updated_at = NOW()`,
			d.ID, d.Name, nullStr(d.Description), d.IsActive, d.SetCategory)
		if err != nil {
			return fmt.Errorf("apply set %q: %w", d.Name, err)
		}
	}
	return nil
}

func applySetItems(ctx context.Context, tx *sql.Tx, sis []setItemDTO) error {
	for _, d := range sis {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO set_items (set_id, item_id, quantity, notes, sort_order)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (set_id, item_id) DO UPDATE SET quantity = EXCLUDED.quantity,
				notes = EXCLUDED.notes, sort_order = EXCLUDED.sort_order`,
			d.SetID, d.ItemID, d.Quantity, nullStr(d.Notes), d.SortOrder)
		if err != nil {
			return fmt.Errorf("apply set item: %w", err)
		}
	}
	return nil
}

func applyPacks(ctx context.Context, tx *sql.Tx, packs []packDTO) error {
	for _, d := range packs {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO packs (id, person_id, name, trip_type, notes, is_template)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET person_id = EXCLUDED.person_id, name = EXCLUDED.name,
				trip_type = EXCLUDED.trip_type, notes = EXCLUDED.notes, is_template = EXCLUDED.is_template,
				updated_at = NOW()`,
			d.ID, uuidPtrArg(d.PersonID), d.Name, nullStr(d.TripType), nullStr(d.Notes), d.IsTemplate)
		if err != nil {
			return fmt.Errorf("apply pack %q: %w", d.Name, err)
		}
	}
	return nil
}

func applyPackItems(ctx context.Context, tx *sql.Tx, pis []packItemDTO) error {
	for _, d := range pis {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO pack_items (pack_id, item_id, quantity, carry_status, notes)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (pack_id, item_id) DO UPDATE SET quantity = EXCLUDED.quantity,
				carry_status = EXCLUDED.carry_status, notes = EXCLUDED.notes, updated_at = NOW()`,
			d.PackID, d.ItemID, d.Quantity, d.CarryStatus, nullStr(d.Notes))
		if err != nil {
			return fmt.Errorf("apply pack item: %w", err)
		}
	}
	return nil
}

func applyPackingLists(ctx context.Context, tx *sql.Tx, pls []packingListDTO) error {
	for _, d := range pls {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO packing_lists (id, name, description) VALUES ($1, $2, $3)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description,
				updated_at = NOW()`,
			d.ID, d.Name, nullStr(d.Description))
		if err != nil {
			return fmt.Errorf("apply packing list %q: %w", d.Name, err)
		}
	}
	return nil
}

func applyPackingListLabels(ctx context.Context, tx *sql.Tx, plls []packingListLabelDTO, labelMap map[uuid.UUID]uuid.UUID) error {
	for _, d := range plls {
		labelID, ok := labelMap[d.LabelID]
		if !ok {
			continue
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO packing_list_labels (packing_list_id, label_id) VALUES ($1, $2)
			ON CONFLICT (packing_list_id, label_id) DO NOTHING`, d.PackingListID, labelID)
		if err != nil {
			return fmt.Errorf("apply packing list label: %w", err)
		}
	}
	return nil
}

func applyTrips(ctx context.Context, tx *sql.Tx, trips []tripDTO) error {
	for _, d := range trips {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO trips (id, name, trip_type, duration, notes, trip_komoot_url,
				trip_strava_url, trip_wanderer_url, total_distance_km)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, trip_type = EXCLUDED.trip_type,
				duration = EXCLUDED.duration, notes = EXCLUDED.notes,
				trip_komoot_url = EXCLUDED.trip_komoot_url, trip_strava_url = EXCLUDED.trip_strava_url,
				trip_wanderer_url = EXCLUDED.trip_wanderer_url, total_distance_km = EXCLUDED.total_distance_km,
				updated_at = NOW()`,
			d.ID, d.Name, d.TripType, nullStr(d.Duration), nullStr(d.Notes), nullStr(d.TripKomootURL),
			nullStr(d.TripStravaURL), nullStr(d.TripWandererURL), nullFloat(d.TotalDistanceKm))
		if err != nil {
			return fmt.Errorf("apply trip %q: %w", d.Name, err)
		}
	}
	return nil
}

func applyTripPersons(ctx context.Context, tx *sql.Tx, tps []tripPersonDTO) (map[uuid.UUID]uuid.UUID, error) {
	out := make(map[uuid.UUID]uuid.UUID, len(tps))
	for _, d := range tps {
		var resolved uuid.UUID
		err := tx.QueryRowContext(ctx, `
			INSERT INTO trip_persons (id, trip_id, person_id) VALUES ($1, $2, $3)
			ON CONFLICT (trip_id, person_id) DO UPDATE SET updated_at = NOW()
			RETURNING id`, d.ID, d.TripID, d.PersonID).Scan(&resolved)
		if err != nil {
			return nil, fmt.Errorf("apply trip person: %w", err)
		}
		out[d.ID] = resolved
	}
	return out, nil
}

func applyTripPersonPacks(ctx context.Context, tx *sql.Tx, tpps []tripPersonPackDTO, tripPersonMap map[uuid.UUID]uuid.UUID) error {
	for _, d := range tpps {
		tripPersonID, ok := tripPersonMap[d.TripPersonID]
		if !ok {
			continue
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO trip_person_packs (trip_person_id, pack_id) VALUES ($1, $2)
			ON CONFLICT (trip_person_id, pack_id) DO NOTHING`, tripPersonID, d.PackID)
		if err != nil {
			return fmt.Errorf("apply trip person pack: %w", err)
		}
	}
	return nil
}

func applyTripPersonItems(ctx context.Context, tx *sql.Tx, tpis []tripPersonItemDTO, tripPersonMap map[uuid.UUID]uuid.UUID) error {
	for _, d := range tpis {
		tripPersonID, ok := tripPersonMap[d.TripPersonID]
		if !ok {
			continue
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO trip_person_items (trip_person_id, item_id, quantity, carry_status, notes)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (trip_person_id, item_id) DO UPDATE SET quantity = EXCLUDED.quantity,
				carry_status = EXCLUDED.carry_status, notes = EXCLUDED.notes`,
			tripPersonID, d.ItemID, d.Quantity, nullStr(d.CarryStatus), nullStr(d.Notes))
		if err != nil {
			return fmt.Errorf("apply trip person item: %w", err)
		}
	}
	return nil
}
