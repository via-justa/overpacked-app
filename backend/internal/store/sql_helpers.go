package store

import (
	"database/sql"
	"time"
)

func toNullString(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

func toNullFloat64(v *float64) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *v, Valid: true}
}

func toNullInt(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}

func toNullBool(v *bool) sql.NullBool {
	if v == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: *v, Valid: true}
}

func toNullBytes(v []byte) []byte {
	if len(v) == 0 {
		return nil
	}

	return append([]byte(nil), v...)
}

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

func intPtrFromNull(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int64)
	return &i
}

func boolPtrFromNull(v sql.NullBool) *bool {
	if !v.Valid {
		return nil
	}
	b := v.Bool
	return &b
}

func timePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}
