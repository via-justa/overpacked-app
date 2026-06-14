package seeds

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

//go:embed data/*.yaml
var seedFiles embed.FS

const fileErrorFormat = "%s: %w"

// SeedFile represents a parsed seed YAML file
type SeedFile struct {
	Filename string
	Table    string
	Records  []map[string]any
}

// SeedData is the structure of a seed YAML file
type SeedData struct {
	Table   string           `yaml:"table"`
	Records []map[string]any `yaml:"records"`
}

// LoadSeedFiles loads all seed YAML files from the embedded filesystem
func LoadSeedFiles() ([]SeedFile, error) {
	var result []SeedFile

	entries, err := fs.ReadDir(seedFiles, "data")
	if err != nil {
		return nil, fmt.Errorf("read seed data directory: %w", err)
	}

	var filenames []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			filenames = append(filenames, entry.Name())
		}
	}

	// Sort to ensure consistent execution order
	sort.Strings(filenames)

	for _, filename := range filenames {
		content, err := fs.ReadFile(seedFiles, filepath.Join("data", filename))
		if err != nil {
			return nil, fmt.Errorf("read seed file %s: %w", filename, err)
		}

		var data SeedData
		if err := yaml.Unmarshal(content, &data); err != nil {
			return nil, fmt.Errorf("parse seed file %s: %w", filename, err)
		}

		if data.Table == "" {
			return nil, fmt.Errorf("seed file %s: missing 'table' field", filename)
		}

		if len(data.Records) == 0 {
			return nil, fmt.Errorf("seed file %s: no records found", filename)
		}

		result = append(result, SeedFile{
			Filename: filename,
			Table:    data.Table,
			Records:  data.Records,
		})
	}

	return result, nil
}

// ValidateSeedFile validates a seed file based on its target table
func ValidateSeedFile(file SeedFile) error {
	switch file.Table {
	case "labels":
		return validateSeedRecords(file, ValidateLabelsRecord)
	case "manufacturers":
		return validateSeedRecords(file, ValidateManufacturersRecord)
	default:
		return fmt.Errorf("%s: unsupported table '%s'", file.Filename, file.Table)
	}
}

// validateSeedRecords enforces name/id uniqueness and runs a per-record
// validator, wrapping any failure with the file name for context.
func validateSeedRecords(file SeedFile, validate func(record map[string]any, index int) error) error {
	if err := CheckDuplicates(file.Records, "name"); err != nil {
		return fmt.Errorf(fileErrorFormat, file.Filename, err)
	}
	if err := CheckDuplicates(file.Records, "id"); err != nil {
		return fmt.Errorf(fileErrorFormat, file.Filename, err)
	}
	for i, record := range file.Records {
		if err := validate(record, i); err != nil {
			return fmt.Errorf(fileErrorFormat, file.Filename, err)
		}
	}
	return nil
}
