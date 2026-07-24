package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// CursorPage represents a page of results with cursor-based pagination.
type CursorPage[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Total      int    `json:"total,omitempty"`
}

// Cursor encodes the pagination position.
type Cursor struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Direction string    `json:"direction"` // "next" or "prev"
}

// EncodeCursor serializes a cursor to a URL-safe base64 string.
func EncodeCursor(c Cursor) string {
	data, _ := json.Marshal(c)
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeCursor deserializes a cursor from a base64 string.
func DecodeCursor(s string) (*Cursor, error) {
	if s == "" {
		return nil, nil
	}
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor encoding: %w", err)
	}
	var c Cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}
	return &c, nil
}

// CursorPaginateParams holds pagination parameters from the request.
type CursorPaginateParams struct {
	Cursor   string `query:"cursor"`
	Limit    int    `query:"limit"`
	MaxLimit int    // server-enforced maximum
}

// NormalizeLimit ensures the limit is within bounds.
func (p *CursorPaginateParams) NormalizeLimit() int {
	if p.MaxLimit == 0 {
		p.MaxLimit = 100
	}
	if p.Limit <= 0 {
		return 20 // default page size
	}
	if p.Limit > p.MaxLimit {
		return p.MaxLimit
	}
	return p.Limit
}

// BuildCursorSQL generates SQL WHERE clause + ORDER BY for cursor pagination.
// tablAlias is the table alias (or empty string).
// Returns the WHERE condition and the sort order.
func BuildCursorSQL(cursor *Cursor, tableAlias string) (whereClause string, orderBy string) {
	prefix := ""
	if tableAlias != "" {
		prefix = tableAlias + "."
	}

	orderBy = fmt.Sprintf("%screated_at DESC, %sid DESC", prefix, prefix)

	if cursor == nil {
		return "", orderBy
	}

	if cursor.Direction == "prev" {
		whereClause = fmt.Sprintf(
			"(%screated_at > $cursor_time OR (%screated_at = $cursor_time AND %sid > $cursor_id))",
			prefix, prefix, prefix)
	} else {
		whereClause = fmt.Sprintf(
			"(%screated_at < $cursor_time OR (%screated_at = $cursor_time AND %sid < $cursor_id))",
			prefix, prefix, prefix)
	}

	return whereClause, orderBy
}
