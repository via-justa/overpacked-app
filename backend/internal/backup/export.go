package backup

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"path"
)

// BuildArchive streams a full backup ZIP (manifest.json, data.json, images/) to w.
func (s *Service) BuildArchive(ctx context.Context, w io.Writer) error {
	snapshot, images, err := s.collect(ctx)
	if err != nil {
		return err
	}

	zw := zip.NewWriter(w)

	if err := writeJSONEntry(zw, dataFilename, snapshot); err != nil {
		return err
	}

	for name, blob := range images {
		f, err := zw.Create(path.Join(imagesDir, name))
		if err != nil {
			return fmt.Errorf("create image entry %s: %w", name, err)
		}
		if _, err := f.Write(blob); err != nil {
			return fmt.Errorf("write image entry %s: %w", name, err)
		}
	}

	manifest := Manifest{
		FormatVersion: FormatVersion,
		CreatedAt:     timeNow(),
		Counts:        snapshot.counts(),
	}
	if err := writeJSONEntry(zw, manifestFilename, manifest); err != nil {
		return err
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("finalize archive: %w", err)
	}
	return nil
}

func writeJSONEntry(zw *zip.Writer, name string, payload any) error {
	f, err := zw.Create(name)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		return fmt.Errorf("encode %s: %w", name, err)
	}
	return nil
}

// collect reads all user data into a Snapshot and returns the image blobs keyed
// by their archive filename (images/<filename>).
func (s *Service) collect(ctx context.Context) (*Snapshot, map[string][]byte, error) {
	snap := &Snapshot{}
	images := map[string][]byte{}

	var err error
	if snap.Settings, err = s.querySettings(ctx); err != nil {
		return nil, nil, err
	}
	if snap.Manufacturers, err = s.queryManufacturers(ctx); err != nil {
		return nil, nil, err
	}
	if snap.Labels, err = s.queryLabels(ctx); err != nil {
		return nil, nil, err
	}
	if snap.ItemTypes, err = s.queryItemTypes(ctx); err != nil {
		return nil, nil, err
	}
	if snap.ItemTypeFields, err = s.queryItemTypeFields(ctx); err != nil {
		return nil, nil, err
	}
	if snap.Persons, err = s.queryPersons(ctx); err != nil {
		return nil, nil, err
	}
	if snap.Items, err = s.queryItems(ctx, images); err != nil {
		return nil, nil, err
	}
	if snap.ItemLabels, err = s.queryItemLabels(ctx); err != nil {
		return nil, nil, err
	}
	if snap.Sets, err = s.querySets(ctx); err != nil {
		return nil, nil, err
	}
	if snap.SetItems, err = s.querySetItems(ctx); err != nil {
		return nil, nil, err
	}
	if snap.Packs, err = s.queryPacks(ctx); err != nil {
		return nil, nil, err
	}
	if snap.PackItems, err = s.queryPackItems(ctx); err != nil {
		return nil, nil, err
	}
	if snap.PackingLists, err = s.queryPackingLists(ctx); err != nil {
		return nil, nil, err
	}
	if snap.PackingListLabels, err = s.queryPackingListLabels(ctx); err != nil {
		return nil, nil, err
	}
	if snap.Trips, err = s.queryTrips(ctx); err != nil {
		return nil, nil, err
	}
	if snap.TripPersons, err = s.queryTripPersons(ctx); err != nil {
		return nil, nil, err
	}
	if snap.TripPersonPacks, err = s.queryTripPersonPacks(ctx); err != nil {
		return nil, nil, err
	}
	if snap.TripPersonItems, err = s.queryTripPersonItems(ctx); err != nil {
		return nil, nil, err
	}

	return snap, images, nil
}

func (s *Service) querySettings(ctx context.Context) (*settingsDTO, error) {
	var d settingsDTO
	err := s.db.QueryRowContext(ctx, `
		SELECT weight_unit, distance_unit, temperature_unit, volume_unit, currency
		FROM settings WHERE id = 1`).
		Scan(&d.WeightUnit, &d.DistanceUnit, &d.TemperatureUnit, &d.VolumeUnit, &d.Currency)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query settings: %w", err)
	}
	return &d, nil
}

// queryManufacturers exports only manufacturers referenced by an item.
func (s *Service) queryManufacturers(ctx context.Context) ([]manufacturerDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT m.id, m.name, m.website
		FROM manufacturers m
		JOIN items i ON i.manufacturer_id = m.id
		ORDER BY m.name`)
	if err != nil {
		return nil, fmt.Errorf("query manufacturers: %w", err)
	}
	defer rows.Close()

	out := []manufacturerDTO{}
	for rows.Next() {
		var d manufacturerDTO
		var website sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &website); err != nil {
			return nil, fmt.Errorf("scan manufacturer: %w", err)
		}
		d.Website = strPtr(website)
		out = append(out, d)
	}
	return out, rows.Err()
}

// queryLabels exports only labels referenced by an item or packing list.
func (s *Service) queryLabels(ctx context.Context) ([]labelDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, color FROM labels
		WHERE id IN (SELECT label_id FROM item_labels
		             UNION SELECT label_id FROM packing_list_labels)
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query labels: %w", err)
	}
	defer rows.Close()

	out := []labelDTO{}
	for rows.Next() {
		var d labelDTO
		var color sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &color); err != nil {
			return nil, fmt.Errorf("scan label: %w", err)
		}
		d.Color = strPtr(color)
		out = append(out, d)
	}
	return out, rows.Err()
}

// queryItemTypes exports custom (non-system) item types only.
func (s *Service) queryItemTypes(ctx context.Context) ([]itemTypeDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description, base_profile
		FROM item_types WHERE is_system = FALSE ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query item types: %w", err)
	}
	defer rows.Close()

	out := []itemTypeDTO{}
	for rows.Next() {
		var d itemTypeDTO
		var description, baseProfile sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &description, &baseProfile); err != nil {
			return nil, fmt.Errorf("scan item type: %w", err)
		}
		d.Description = strPtr(description)
		d.BaseProfile = strPtr(baseProfile)
		out = append(out, d)
	}
	return out, rows.Err()
}

// queryItemTypeFields exports fields for custom item types only.
func (s *Service) queryItemTypeFields(ctx context.Context) ([]itemTypeFieldDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT f.id, f.item_type_id, f.field_key, f.field_label, f.data_type, f.is_required,
		       f.enum_options_json, f.min_value, f.max_value, f.unit, f.sort_order
		FROM item_type_fields f
		JOIN item_types t ON t.id = f.item_type_id
		WHERE t.is_system = FALSE
		ORDER BY f.item_type_id, f.sort_order`)
	if err != nil {
		return nil, fmt.Errorf("query item type fields: %w", err)
	}
	defer rows.Close()

	out := []itemTypeFieldDTO{}
	for rows.Next() {
		var d itemTypeFieldDTO
		var enumOptions []byte
		var minValue, maxValue sql.NullFloat64
		var unit sql.NullString
		if err := rows.Scan(&d.ID, &d.ItemTypeID, &d.FieldKey, &d.FieldLabel, &d.DataType,
			&d.IsRequired, &enumOptions, &minValue, &maxValue, &unit, &d.SortOrder); err != nil {
			return nil, fmt.Errorf("scan item type field: %w", err)
		}
		d.EnumOptions = rawOrNil(enumOptions)
		d.MinValue = floatPtr(minValue)
		d.MaxValue = floatPtr(maxValue)
		d.Unit = strPtr(unit)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) queryPersons(ctx context.Context) ([]personDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, gender, birthdate, body_weight_grams, conditioning_level
		FROM persons ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query persons: %w", err)
	}
	defer rows.Close()

	out := []personDTO{}
	for rows.Next() {
		var d personDTO
		var gender, conditioning sql.NullString
		var birthdate sql.NullTime
		var bodyWeight sql.NullFloat64
		if err := rows.Scan(&d.ID, &d.Name, &gender, &birthdate, &bodyWeight, &conditioning); err != nil {
			return nil, fmt.Errorf("scan person: %w", err)
		}
		d.Gender = strPtr(gender)
		d.Birthdate = timePtr(birthdate)
		d.BodyWeightGrams = floatPtr(bodyWeight)
		d.ConditioningLevel = strPtr(conditioning)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) queryItems(ctx context.Context, images map[string][]byte) ([]itemDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, manufacturer_id, type_id, name, is_active, description, source_url,
		       price, weight_grams, volume_ml, default_quantity, default_carry_status, is_default,
		       image_blob, image_mime_type, image_size_bytes, image_width_px, image_height_px, attributes_json
		FROM items ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	out := []itemDTO{}
	for rows.Next() {
		var d itemDTO
		var description, sourceURL, mimeType sql.NullString
		var price, weight, volume sql.NullFloat64
		var sizeBytes, widthPX, heightPX sql.NullInt64
		var imageBlob, attributes []byte
		if err := rows.Scan(&d.ID, &d.ManufacturerID, &d.TypeID, &d.Name, &d.IsActive,
			&description, &sourceURL, &price, &weight, &volume, &d.DefaultQuantity,
			&d.DefaultCarryStatus, &d.IsDefault, &imageBlob, &mimeType, &sizeBytes,
			&widthPX, &heightPX, &attributes); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		d.Description = strPtr(description)
		d.SourceURL = strPtr(sourceURL)
		d.Price = floatPtr(price)
		d.WeightGrams = floatPtr(weight)
		d.VolumeML = floatPtr(volume)
		d.ImageMimeType = strPtr(mimeType)
		d.ImageSizeBytes = intPtr(sizeBytes)
		d.ImageWidthPX = intPtr(widthPX)
		d.ImageHeightPX = intPtr(heightPX)
		d.Attributes = rawOrNil(attributes)

		if len(imageBlob) > 0 && mimeType.Valid {
			name := fmt.Sprintf("%s.%s", d.ID, imageExtForMime(mimeType.String))
			images[name] = append([]byte(nil), imageBlob...)
			d.ImageFile = &name
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) queryItemLabels(ctx context.Context) ([]itemLabelDTO, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT item_id, label_id FROM item_labels`)
	if err != nil {
		return nil, fmt.Errorf("query item labels: %w", err)
	}
	defer rows.Close()

	out := []itemLabelDTO{}
	for rows.Next() {
		var d itemLabelDTO
		if err := rows.Scan(&d.ItemID, &d.LabelID); err != nil {
			return nil, fmt.Errorf("scan item label: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) querySets(ctx context.Context) ([]setDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description, is_active, set_category FROM item_sets ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query sets: %w", err)
	}
	defer rows.Close()

	out := []setDTO{}
	for rows.Next() {
		var d setDTO
		var description sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &description, &d.IsActive, &d.SetCategory); err != nil {
			return nil, fmt.Errorf("scan set: %w", err)
		}
		d.Description = strPtr(description)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) querySetItems(ctx context.Context) ([]setItemDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT set_id, item_id, quantity, notes, sort_order FROM set_items`)
	if err != nil {
		return nil, fmt.Errorf("query set items: %w", err)
	}
	defer rows.Close()

	out := []setItemDTO{}
	for rows.Next() {
		var d setItemDTO
		var notes sql.NullString
		if err := rows.Scan(&d.SetID, &d.ItemID, &d.Quantity, &notes, &d.SortOrder); err != nil {
			return nil, fmt.Errorf("scan set item: %w", err)
		}
		d.Notes = strPtr(notes)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) queryPacks(ctx context.Context) ([]packDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, person_id, name, trip_type, notes, is_template FROM packs ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query packs: %w", err)
	}
	defer rows.Close()

	out := []packDTO{}
	for rows.Next() {
		var d packDTO
		var personID sql.NullString
		var tripType, notes sql.NullString
		if err := rows.Scan(&d.ID, &personID, &d.Name, &tripType, &notes, &d.IsTemplate); err != nil {
			return nil, fmt.Errorf("scan pack: %w", err)
		}
		if personID.Valid {
			if id, perr := uuidParse(personID.String); perr == nil {
				d.PersonID = &id
			}
		}
		d.TripType = strPtr(tripType)
		d.Notes = strPtr(notes)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) queryPackItems(ctx context.Context) ([]packItemDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT pack_id, item_id, quantity, carry_status, notes FROM pack_items`)
	if err != nil {
		return nil, fmt.Errorf("query pack items: %w", err)
	}
	defer rows.Close()

	out := []packItemDTO{}
	for rows.Next() {
		var d packItemDTO
		var notes sql.NullString
		if err := rows.Scan(&d.PackID, &d.ItemID, &d.Quantity, &d.CarryStatus, &notes); err != nil {
			return nil, fmt.Errorf("scan pack item: %w", err)
		}
		d.Notes = strPtr(notes)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) queryPackingLists(ctx context.Context) ([]packingListDTO, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, description FROM packing_lists ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query packing lists: %w", err)
	}
	defer rows.Close()

	out := []packingListDTO{}
	for rows.Next() {
		var d packingListDTO
		var description sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &description); err != nil {
			return nil, fmt.Errorf("scan packing list: %w", err)
		}
		d.Description = strPtr(description)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) queryPackingListLabels(ctx context.Context) ([]packingListLabelDTO, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT packing_list_id, label_id FROM packing_list_labels`)
	if err != nil {
		return nil, fmt.Errorf("query packing list labels: %w", err)
	}
	defer rows.Close()

	out := []packingListLabelDTO{}
	for rows.Next() {
		var d packingListLabelDTO
		if err := rows.Scan(&d.PackingListID, &d.LabelID); err != nil {
			return nil, fmt.Errorf("scan packing list label: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) queryTrips(ctx context.Context) ([]tripDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, trip_type, duration, notes, trip_komoot_url, trip_strava_url,
		       trip_wanderer_url, total_distance_km
		FROM trips ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query trips: %w", err)
	}
	defer rows.Close()

	out := []tripDTO{}
	for rows.Next() {
		var d tripDTO
		var duration, notes, komoot, strava, wanderer sql.NullString
		var distance sql.NullFloat64
		if err := rows.Scan(&d.ID, &d.Name, &d.TripType, &duration, &notes, &komoot,
			&strava, &wanderer, &distance); err != nil {
			return nil, fmt.Errorf("scan trip: %w", err)
		}
		d.Duration = strPtr(duration)
		d.Notes = strPtr(notes)
		d.TripKomootURL = strPtr(komoot)
		d.TripStravaURL = strPtr(strava)
		d.TripWandererURL = strPtr(wanderer)
		d.TotalDistanceKm = floatPtr(distance)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) queryTripPersons(ctx context.Context) ([]tripPersonDTO, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, trip_id, person_id FROM trip_persons`)
	if err != nil {
		return nil, fmt.Errorf("query trip persons: %w", err)
	}
	defer rows.Close()

	out := []tripPersonDTO{}
	for rows.Next() {
		var d tripPersonDTO
		if err := rows.Scan(&d.ID, &d.TripID, &d.PersonID); err != nil {
			return nil, fmt.Errorf("scan trip person: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) queryTripPersonPacks(ctx context.Context) ([]tripPersonPackDTO, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT trip_person_id, pack_id FROM trip_person_packs`)
	if err != nil {
		return nil, fmt.Errorf("query trip person packs: %w", err)
	}
	defer rows.Close()

	out := []tripPersonPackDTO{}
	for rows.Next() {
		var d tripPersonPackDTO
		if err := rows.Scan(&d.TripPersonID, &d.PackID); err != nil {
			return nil, fmt.Errorf("scan trip person pack: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) queryTripPersonItems(ctx context.Context) ([]tripPersonItemDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT trip_person_id, item_id, quantity, carry_status, notes FROM trip_person_items`)
	if err != nil {
		return nil, fmt.Errorf("query trip person items: %w", err)
	}
	defer rows.Close()

	out := []tripPersonItemDTO{}
	for rows.Next() {
		var d tripPersonItemDTO
		var carryStatus, notes sql.NullString
		if err := rows.Scan(&d.TripPersonID, &d.ItemID, &d.Quantity, &carryStatus, &notes); err != nil {
			return nil, fmt.Errorf("scan trip person item: %w", err)
		}
		d.CarryStatus = strPtr(carryStatus)
		d.Notes = strPtr(notes)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (snap *Snapshot) counts() map[string]int {
	return map[string]int{
		"manufacturers":       len(snap.Manufacturers),
		"labels":              len(snap.Labels),
		"item_types":          len(snap.ItemTypes),
		"item_type_fields":    len(snap.ItemTypeFields),
		"persons":             len(snap.Persons),
		"items":               len(snap.Items),
		"item_labels":         len(snap.ItemLabels),
		"item_sets":           len(snap.Sets),
		"set_items":           len(snap.SetItems),
		"packs":               len(snap.Packs),
		"pack_items":          len(snap.PackItems),
		"packing_lists":       len(snap.PackingLists),
		"packing_list_labels": len(snap.PackingListLabels),
		"trips":               len(snap.Trips),
		"trip_persons":        len(snap.TripPersons),
		"trip_person_packs":   len(snap.TripPersonPacks),
		"trip_person_items":   len(snap.TripPersonItems),
	}
}
