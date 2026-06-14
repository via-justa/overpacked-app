package backup

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

func timeNow() time.Time { return time.Now().UTC() }

func uuidParse(s string) (uuid.UUID, error) { return uuid.Parse(s) }

func strPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func floatPtr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	f := v.Float64
	return &f
}

func intPtr(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int64)
	return &i
}

func timePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}

func rawOrNil(b []byte) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	return append(json.RawMessage(nil), b...)
}

// nullStr builds a sql.NullString from an optional pointer.
func nullStr(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

func nullInt(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}

func nullFloat(v *float64) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *v, Valid: true}
}

func nullTime(v *time.Time) sql.NullTime {
	if v == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *v, Valid: true}
}

// rawOrNull returns a JSONB-insertable value, mapping empty to a SQL NULL.
func rawOrNull(b json.RawMessage) any {
	if len(b) == 0 {
		return nil
	}
	return []byte(b)
}

// attributesOrEmpty returns a JSONB-insertable value for the NOT NULL attributes
// column, defaulting empty to an empty object.
func attributesOrEmpty(b json.RawMessage) []byte {
	if len(b) == 0 {
		return []byte("{}")
	}
	return []byte(b)
}

// blobOrNil maps an empty image blob to a SQL NULL for the BYTEA column.
func blobOrNil(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

// uuidPtrArg maps an optional UUID to a nullable SQL argument.
func uuidPtrArg(v *uuid.UUID) any {
	if v == nil {
		return nil
	}
	return *v
}

// imageExtForMime maps an image MIME type to a file extension for archive filenames.
func imageExtForMime(mime string) string {
	switch mime {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	default:
		return "bin"
	}
}
