package postgres

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/vortexui/vortexui/internal/domain"
)

// This file is the single boundary between pgx wire types (pgtype.*) and plain
// Go/domain types. Centralizing it keeps every repository method readable and
// the null-handling rules in one auditable place.

func timeToTS(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func unixToTime(sec int64) time.Time { return time.Unix(sec, 0).UTC() }

func ptrToTS(p *time.Time) pgtype.Timestamptz {
	if p == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *p, Valid: true}
}

func tsToPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	tt := t.Time
	return &tt
}

func ptrToInt8(p *int64) pgtype.Int8 {
	if p == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *p, Valid: true}
}

func int8ToPtr(v pgtype.Int8) *int64 {
	if !v.Valid {
		return nil
	}
	n := v.Int64
	return &n
}

// ptrToUUID encodes an optional UUID; nil becomes SQL NULL.
func ptrToUUID(p *uuid.UUID) pgtype.UUID {
	if p == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *p, Valid: true}
}

func uuidToPtr(v pgtype.UUID) *uuid.UUID {
	if !v.Valid {
		return nil
	}
	id := uuid.UUID(v.Bytes)
	return &id
}

// jsonbStrings marshals a string slice for a JSONB column, never emitting null.
func jsonbStrings(ss []string) []byte {
	if ss == nil {
		ss = []string{}
	}
	b, _ := json.Marshal(ss)
	return b
}

func stringsFromJSONB(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	var out []string
	_ = json.Unmarshal(b, &out)
	return out
}

func jsonbMap(m map[string]any) []byte {
	if m == nil {
		m = map[string]any{}
	}
	b, _ := json.Marshal(m)
	return b
}

func mapFromJSONB(b []byte) map[string]any {
	if len(b) == 0 {
		return nil
	}
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	return out
}

// geoPolicyToJSONB marshals an optional geo policy for a JSONB column. A nil
// policy becomes a nil slice so the column stays SQL NULL.
func geoPolicyToJSONB(p *domain.GeoPolicy) []byte {
	if p == nil {
		return nil
	}
	b, err := json.Marshal(p)
	if err != nil {
		return nil
	}
	return b
}

// geoPolicyFromJSONB decodes a geo policy from a JSONB column. Empty bytes,
// decode errors, or a policy with no allowed/blocked countries all map to nil
// (treated as "no policy").
func geoPolicyFromJSONB(b []byte) *domain.GeoPolicy {
	if len(b) == 0 {
		return nil
	}
	var p domain.GeoPolicy
	if err := json.Unmarshal(b, &p); err != nil {
		return nil
	}
	if len(p.AllowedCountries) == 0 && len(p.BlockedCountries) == 0 {
		return nil
	}
	return &p
}

// parseBucket turns a friendly bucket string ("1h", "30m", "1d", "7d") into a
// pgtype.Interval for date_bin. Days are normalized to hours since Go's
// time.ParseDuration has no day unit.
func parseBucket(s string) (pgtype.Interval, error) {
	if s == "" {
		s = "1h"
	}
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return pgtype.Interval{}, fmt.Errorf("bad bucket %q: %w", s, err)
		}
		return pgtype.Interval{Microseconds: int64(n) * 24 * int64(time.Hour) / 1000, Valid: true}, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return pgtype.Interval{}, fmt.Errorf("bad bucket %q: %w", s, err)
	}
	return pgtype.Interval{Microseconds: d.Microseconds(), Valid: true}, nil
}
