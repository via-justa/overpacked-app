package seeds

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

const (
	seedsErrLoadFailed     = "failed to load seed files"
	seedsErrValidateFailed = "failed to validate seed files"
	seedsErrSeedFailed     = "failed to seed database"
)

// Run executes all seed files in order
func Run(ctx context.Context, db *sql.DB) error {
	fmt.Println("Loading seed files...")
	seedFiles, err := LoadSeedFiles()
	if err != nil {
		return fmt.Errorf("%s: %w", seedsErrLoadFailed, err)
	}

	fmt.Printf("Found %d seed file(s)\n", len(seedFiles))

	// Validate all files before seeding
	fmt.Println("Validating seed files...")
	for _, file := range seedFiles {
		if err := ValidateSeedFile(file); err != nil {
			return fmt.Errorf("%s: %w", seedsErrValidateFailed, err)
		}
	}

	fmt.Println("All seed files validated successfully")

	// Seed each table
	for _, file := range seedFiles {
		fmt.Printf("Seeding %s from %s...\n", file.Table, file.Filename)

		var inserted, skipped int
		switch file.Table {
		case "labels":
			inserted, skipped, err = seedLabels(ctx, db, file.Records)
		case "manufacturers":
			inserted, skipped, err = seedManufacturers(ctx, db, file.Records)
		default:
			return fmt.Errorf("unsupported table: %s", file.Table)
		}

		if err != nil {
			return fmt.Errorf("%s %s: %w", seedsErrSeedFailed, file.Table, err)
		}

		fmt.Printf("  ✓ Inserted: %d, Skipped (already exists): %d\n", inserted, skipped)
	}

	fmt.Println("\n✓ Seeding completed successfully")
	return nil
}

// seedLabels seeds the labels table
func seedLabels(ctx context.Context, db *sql.DB, records []map[string]any) (inserted, skipped int, err error) {
	for _, record := range records {
		name := record["name"].(string)

		// Check if label already exists
		var exists bool
		err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM labels WHERE name = $1)", name).Scan(&exists)
		if err != nil {
			return inserted, skipped, fmt.Errorf("check existing label '%s': %w", name, err)
		}

		if exists {
			skipped++
			continue
		}

		// Insert new label
		id := uuid.New()
		var color *string
		if colorVal, ok := record["color"]; ok && colorVal != nil {
			colorStr := colorVal.(string)
			color = &colorStr
		}

		_, err = db.ExecContext(ctx,
			"INSERT INTO labels (id, name, color) VALUES ($1, $2, $3)",
			id, name, color,
		)
		if err != nil {
			return inserted, skipped, fmt.Errorf("insert label '%s': %w", name, err)
		}

		inserted++
	}

	return inserted, skipped, nil
}

// seedManufacturers seeds the manufacturers table
func seedManufacturers(ctx context.Context, db *sql.DB, records []map[string]any) (inserted, skipped int, err error) {
	for _, record := range records {
		name := record["name"].(string)

		// Check if manufacturer already exists
		var exists bool
		err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM manufacturers WHERE name = $1)", name).Scan(&exists)
		if err != nil {
			return inserted, skipped, fmt.Errorf("check existing manufacturer '%s': %w", name, err)
		}

		if exists {
			skipped++
			continue
		}

		// Insert new manufacturer
		id := uuid.New()
		var website *string
		if websiteVal, ok := record["website"]; ok && websiteVal != nil {
			websiteStr := websiteVal.(string)
			website = &websiteStr
		}

		_, err = db.ExecContext(ctx,
			"INSERT INTO manufacturers (id, name, website) VALUES ($1, $2, $3)",
			id, name, website,
		)
		if err != nil {
			return inserted, skipped, fmt.Errorf("insert manufacturer '%s': %w", name, err)
		}

		inserted++
	}

	return inserted, skipped, nil
}
