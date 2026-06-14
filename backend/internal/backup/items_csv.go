package backup

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"time"
)

const itemsCSVFilename = "items.csv"

var itemsCSVHeader = []string{
	"name", "manufacturer", "item_type", "description", "source_url", "price",
	"weight_grams", "volume_ml", "default_quantity", "default_carry_status",
	"is_active", "is_default", "labels", "attributes", "image_file",
	"image_mime_type", "image_width_px", "image_height_px", "created_at", "updated_at",
}

// ExportItemsCSV writes a generic, items-only denormalized export (deliverable B).
// Without images it writes a CSV directly to w; with images it writes a ZIP
// containing items.csv plus an images/ folder referenced by the image_file column.
func (s *Service) ExportItemsCSV(ctx context.Context, w io.Writer, includeImages bool) error {
	if !includeImages {
		cw := csv.NewWriter(w)
		if _, err := s.writeItemRows(ctx, cw, false); err != nil {
			return err
		}
		cw.Flush()
		return cw.Error()
	}

	zw := zip.NewWriter(w)
	csvFile, err := zw.Create(itemsCSVFilename)
	if err != nil {
		return fmt.Errorf("create items.csv entry: %w", err)
	}
	cw := csv.NewWriter(csvFile)
	images, err := s.writeItemRows(ctx, cw, true)
	if err != nil {
		return err
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return err
	}

	// Write image entries only after the CSV entry is complete: archive/zip
	// permits a single open entry at a time, so each zw.Create closes the prior
	// one. Interleaving image entries with CSV writes corrupts the CSV stream.
	for _, img := range images {
		f, err := zw.Create(path.Join(imagesDir, img.name))
		if err != nil {
			return fmt.Errorf("create image entry %s: %w", img.name, err)
		}
		if _, err := f.Write(img.blob); err != nil {
			return fmt.Errorf("write image entry %s: %w", img.name, err)
		}
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("finalize export archive: %w", err)
	}
	return nil
}

// csvImage is an item image collected during a CSV export, to be written into
// the archive after the CSV entry is closed.
type csvImage struct {
	name string
	blob []byte
}

// writeItemRows streams item rows to cw and, when collectImages is set, records
// each item's image (filename in the image_file column) for the caller to write
// into the archive afterwards.
func (s *Service) writeItemRows(ctx context.Context, cw *csv.Writer, collectImages bool) ([]csvImage, error) {
	if err := cw.Write(itemsCSVHeader); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT i.name, m.name, t.name, i.description, i.source_url, i.price,
			i.weight_grams, i.volume_ml, i.default_quantity, i.default_carry_status,
			i.is_active, i.is_default,
			COALESCE(string_agg(DISTINCT l.name, '; ' ORDER BY l.name), '') AS labels,
			i.attributes_json, i.image_blob, i.image_mime_type, i.image_width_px,
			i.image_height_px, i.created_at, i.updated_at
		FROM items i
		JOIN manufacturers m ON m.id = i.manufacturer_id
		JOIN item_types t ON t.id = i.type_id
		LEFT JOIN item_labels il ON il.item_id = i.id
		LEFT JOIN labels l ON l.id = il.label_id
		GROUP BY i.id, m.name, t.name
		ORDER BY i.name`)
	if err != nil {
		return nil, fmt.Errorf("query items for csv: %w", err)
	}
	defer rows.Close()

	usedNames := map[string]int{}
	var images []csvImage
	for rows.Next() {
		var name, manufacturer, itemType, labels, carryStatus string
		var description, sourceURL, mimeType sql.NullString
		var price, weight, volume sql.NullFloat64
		var widthPX, heightPX sql.NullInt64
		var defaultQty int
		var isActive, isDefault bool
		var attributes, imageBlob []byte
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&name, &manufacturer, &itemType, &description, &sourceURL, &price,
			&weight, &volume, &defaultQty, &carryStatus, &isActive, &isDefault, &labels,
			&attributes, &imageBlob, &mimeType, &widthPX, &heightPX, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan item row for csv: %w", err)
		}

		imageFile := ""
		if collectImages && len(imageBlob) > 0 && mimeType.Valid {
			imageFile = uniqueImageName(usedNames, name, imageExtForMime(mimeType.String))
			images = append(images, csvImage{name: imageFile, blob: append([]byte(nil), imageBlob...)})
		}

		record := []string{
			name, manufacturer, itemType,
			nullToStr(description), nullToStr(sourceURL),
			floatToStr(price), floatToStr(weight), floatToStr(volume),
			strconv.Itoa(defaultQty), carryStatus,
			strconv.FormatBool(isActive), strconv.FormatBool(isDefault),
			labels, attributesToStr(attributes), imageFile,
			nullToStr(mimeType), intToStr(widthPX), intToStr(heightPX),
			createdAt.Format(time.RFC3339), updatedAt.Format(time.RFC3339),
		}
		if err := cw.Write(record); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}

	return images, rows.Err()
}

func uniqueImageName(used map[string]int, itemName, ext string) string {
	base := sanitizeFilename(itemName)
	if base == "" {
		base = "item"
	}
	used[base]++
	return fmt.Sprintf("%s-%d.%s", base, used[base], ext)
}

func sanitizeFilename(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func nullToStr(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return v.String
}

func floatToStr(v sql.NullFloat64) string {
	if !v.Valid {
		return ""
	}
	return strconv.FormatFloat(v.Float64, 'f', -1, 64)
}

func intToStr(v sql.NullInt64) string {
	if !v.Valid {
		return ""
	}
	return strconv.FormatInt(v.Int64, 10)
}

func attributesToStr(b []byte) string {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "{}" {
		return ""
	}
	return s
}
