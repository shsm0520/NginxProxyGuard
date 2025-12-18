package repository

import (
	"database/sql"
	"time"
)

// SQL NULL type conversion helpers
// These functions reduce boilerplate when converting between Go types and SQL null types

// FromNullString converts sql.NullString to *string
func FromNullString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// FromNullTime converts sql.NullTime to *time.Time
func FromNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// FromNullInt64 converts sql.NullInt64 to *int64
func FromNullInt64(ni sql.NullInt64) *int64 {
	if ni.Valid {
		return &ni.Int64
	}
	return nil
}

// FromNullInt32 converts sql.NullInt32 to *int32
func FromNullInt32(ni sql.NullInt32) *int32 {
	if ni.Valid {
		return &ni.Int32
	}
	return nil
}

// FromNullFloat64 converts sql.NullFloat64 to *float64
func FromNullFloat64(nf sql.NullFloat64) *float64 {
	if nf.Valid {
		return &nf.Float64
	}
	return nil
}

// FromNullBool converts sql.NullBool to *bool
func FromNullBool(nb sql.NullBool) *bool {
	if nb.Valid {
		return &nb.Bool
	}
	return nil
}

// ToNullString converts *string to sql.NullString
func ToNullString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{String: *s, Valid: true}
	}
	return sql.NullString{}
}

// ToNullStringFromValue converts string to sql.NullString (empty string = invalid)
func ToNullStringFromValue(s string) sql.NullString {
	if s != "" {
		return sql.NullString{String: s, Valid: true}
	}
	return sql.NullString{}
}

// ToNullTime converts *time.Time to sql.NullTime
func ToNullTime(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}
	return sql.NullTime{}
}

// ToNullInt64 converts *int64 to sql.NullInt64
func ToNullInt64(i *int64) sql.NullInt64 {
	if i != nil {
		return sql.NullInt64{Int64: *i, Valid: true}
	}
	return sql.NullInt64{}
}

// ToNullInt32 converts *int32 to sql.NullInt32
func ToNullInt32(i *int32) sql.NullInt32 {
	if i != nil {
		return sql.NullInt32{Int32: *i, Valid: true}
	}
	return sql.NullInt32{}
}

// ToNullFloat64 converts *float64 to sql.NullFloat64
func ToNullFloat64(f *float64) sql.NullFloat64 {
	if f != nil {
		return sql.NullFloat64{Float64: *f, Valid: true}
	}
	return sql.NullFloat64{}
}

// ToNullBool converts *bool to sql.NullBool
func ToNullBool(b *bool) sql.NullBool {
	if b != nil {
		return sql.NullBool{Bool: *b, Valid: true}
	}
	return sql.NullBool{}
}

// StringValue returns the string value or empty string if nil
func StringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// Int64Value returns the int64 value or 0 if nil
func Int64Value(i *int64) int64 {
	if i != nil {
		return *i
	}
	return 0
}
