package seeds

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// validateID ensures a record carries a parseable UUID 'id'. Seed ids are stable
// so backups can reference seed rows by id across instances.
func validateID(record map[string]any, index int) error {
	id, ok := record["id"]
	if !ok {
		return fmt.Errorf("record %d: missing required field 'id'", index)
	}

	idStr, ok := id.(string)
	if !ok {
		return fmt.Errorf("record %d: field 'id' must be a string", index)
	}

	if _, err := uuid.Parse(idStr); err != nil {
		return fmt.Errorf("record %d: field 'id' must be a valid UUID: %w", index, err)
	}

	return nil
}

// ValidateLabelsRecord validates a single labels table record
func ValidateLabelsRecord(record map[string]any, index int) error {
	if err := validateID(record, index); err != nil {
		return err
	}

	// name is required
	name, ok := record["name"]
	if !ok {
		return fmt.Errorf("record %d: missing required field 'name'", index)
	}

	nameStr, ok := name.(string)
	if !ok {
		return fmt.Errorf("record %d: field 'name' must be a string", index)
	}

	if nameStr == "" {
		return fmt.Errorf("record %d: field 'name' cannot be empty", index)
	}

	// color is optional but must be valid hex if present
	if color, exists := record["color"]; exists && color != nil {
		colorStr, ok := color.(string)
		if !ok {
			return fmt.Errorf("record %d: field 'color' must be a string", index)
		}

		if colorStr != "" && !hexColorPattern.MatchString(colorStr) {
			return fmt.Errorf("record %d: field 'color' must be a valid hex color (e.g., #FF5733)", index)
		}
	}

	return nil
}

// ValidateManufacturersRecord validates a single manufacturers table record
func ValidateManufacturersRecord(record map[string]any, index int) error {
	if err := validateID(record, index); err != nil {
		return err
	}

	// name is required
	name, ok := record["name"]
	if !ok {
		return fmt.Errorf("record %d: missing required field 'name'", index)
	}

	nameStr, ok := name.(string)
	if !ok {
		return fmt.Errorf("record %d: field 'name' must be a string", index)
	}

	if nameStr == "" {
		return fmt.Errorf("record %d: field 'name' cannot be empty", index)
	}

	// website is optional but must be string if present
	if website, exists := record["website"]; exists && website != nil {
		_, ok := website.(string)
		if !ok {
			return fmt.Errorf("record %d: field 'website' must be a string", index)
		}
	}

	return nil
}

// CheckDuplicates checks for duplicate values in a specific field across all records
func CheckDuplicates(records []map[string]any, fieldName string) error {
	seen := make(map[string]int)

	for i, record := range records {
		if value, exists := record[fieldName]; exists {
			if valueStr, ok := value.(string); ok && valueStr != "" {
				if prevIndex, found := seen[valueStr]; found {
					return fmt.Errorf("duplicate %s '%s' found at records %d and %d", fieldName, valueStr, prevIndex, i)
				}
				seen[valueStr] = i
			}
		}
	}

	return nil
}
