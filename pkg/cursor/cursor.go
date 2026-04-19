// Package cursor provides support for cursor-based (keyset) pagination.
package cursor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Housiadas/cerberus/pkg/order"
)

const (
	limitDefault = 10
	limitMax     = 100
)

// encodedCursor holds the values serialized inside the opaque cursor string.
type encodedCursor struct {
	FieldValue string `json:"v"`
	ID         string `json:"id"`
}

// Cursor represents a parsed, validated cursor pagination request.
type Cursor struct {
	fieldValue string
	id         string
	limit      int
	hasCursor  bool
}

// Parse decodes the cursor string and limit, returning a validated Cursor.
func Parse(cursorStr string, limitStr string) (Cursor, error) {
	limit := limitDefault

	if limitStr != "" {
		var err error

		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return Cursor{}, fmt.Errorf("limit conversion: %w", err)
		}
	}

	if limit <= 0 {
		return Cursor{}, ErrLimitTooSmall
	}

	if limit > limitMax {
		return Cursor{}, ErrLimitTooLarge
	}

	c := Cursor{
		limit: limit,
	}

	if cursorStr != "" {
		raw, err := base64.URLEncoding.DecodeString(cursorStr)
		if err != nil {
			return Cursor{}, fmt.Errorf("%w: %w", ErrInvalidEncoded, err)
		}

		var ec encodedCursor

		err = json.Unmarshal(raw, &ec)
		if err != nil {
			return Cursor{}, fmt.Errorf("%w: %w", ErrInvalidCursor, err)
		}

		if ec.ID == "" {
			return Cursor{}, ErrInvalidCursor
		}

		c.fieldValue = ec.FieldValue
		c.id = ec.ID
		c.hasCursor = true
	}

	return c, nil
}

// Encode produces an opaque cursor string from a sort field value and row ID.
func Encode(fieldValue any, id string) string {
	ec := encodedCursor{
		FieldValue: fmt.Sprintf("%v", fieldValue),
		ID:         id,
	}

	raw, _ := json.Marshal(ec) //nolint:errchkjson

	return base64.URLEncoding.EncodeToString(raw)
}

// HasCursor returns true if a cursor was provided (i.e. not the first page).
func (c Cursor) HasCursor() bool {
	return c.hasCursor
}

// FieldValue returns the sort column value from the cursor.
func (c Cursor) FieldValue() string {
	return c.fieldValue
}

// ID returns the tiebreaking row ID from the cursor.
func (c Cursor) ID() string {
	return c.id
}

// Limit returns the number of rows per page.
func (c Cursor) Limit() int {
	return c.limit
}

// Metadata holds the pagination metadata for a cursor-based response.
type Metadata struct {
	NextCursor string `json:"nextCursor,omitempty"`
	PrevCursor string `json:"prevCursor,omitempty"`
	HasMore    bool   `json:"hasMore"`
	Limit      int    `json:"limit"`
}

// Result is the data model used when returning a cursor-paginated query result.
type Result[T any] struct {
	Data     []T      `json:"data"`
	Metadata Metadata `json:"metadata"`
}

// NewResult constructs a cursor-paginated result from the data returned by the repo.
// The repo should fetch limit+1 rows to detect whether more data exists.
func NewResult[T any](
	data []T,
	limit int,
	cur Cursor,
	_ order.By,
	idExtractor func(T) string,
	fieldExtractor func(T) any,
) Result[T] {
	if len(data) == 0 {
		return Result[T]{
			Data: []T{},
			Metadata: Metadata{
				Limit: limit,
			},
		}
	}

	hasMore := len(data) > limit
	if hasMore {
		data = data[:limit]
	}

	meta := Metadata{
		HasMore: hasMore,
		Limit:   limit,
	}

	if hasMore {
		last := data[len(data)-1]
		meta.NextCursor = Encode(fieldExtractor(last), idExtractor(last))
	}

	if cur.HasCursor() {
		first := data[0]
		meta.PrevCursor = Encode(fieldExtractor(first), idExtractor(first))
	}

	return Result[T]{
		Data:     data,
		Metadata: meta,
	}
}
