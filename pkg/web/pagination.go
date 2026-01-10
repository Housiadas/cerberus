package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// The Result is the data model used when returning a query result.
type Result[T any] struct {
	Data     []T      `json:"data"`
	Metadata Metadata `json:"metadata"`
}

// NewResult constructs a result value to return query results.
func NewResult[T any](data []T, total int, page Page) Result[T] {
	metadata := calculateMetadata(total, page.Number(), page.RowsPerPage())

	return Result[T]{
		Data:     data,
		Metadata: metadata,
	}
}

// Encode implements the encoder interface.
func (r Result[T]) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)

	return data, "application/json", fmt.Errorf("page result encode error: %w", err)
}

// ============================================================================

// Metadata a struct for holding the pagination metadata.
type Metadata struct {
	FirstPage   int `json:"firstPage,omitempty"`
	CurrentPage int `json:"currentPage,omitempty"`
	LastPage    int `json:"lastPage,omitempty"`
	RowsPerPage int `json:"rowsPerPage,omitempty"`
	Total       int `json:"total,omitempty"`
}

// calculateMetadata function calculates the appropriate pagination metadata
// values given the total number of records, current page, and page size values. Note
// that the last page value is calculated using the math.Ceil() function, which rounds
// up a float to the nearest integer. So, for example, if there were 12 records in total
// and a page size of 5, the last page value would be math.Ceil(12/5) = 3.
func calculateMetadata(total, page, rows int) Metadata {
	if total == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage: page,
		RowsPerPage: rows,
		FirstPage:   1,
		LastPage:    int(math.Ceil(float64(total) / float64(rows))),
		Total:       total,
	}
}

// ============================================================================

// Page represents the requested page and rows per page.
type Page struct {
	number int
	rows   int
}

// String implements the stringer interface.
func (p Page) String() string {
	return fmt.Sprintf("page: %d rows: %d", p.number, p.rows)
}

// Number returns the page number.
func (p Page) Number() int {
	return p.number
}

// RowsPerPage returns the rows per page.
func (p Page) RowsPerPage() int {
	return p.rows
}

// Parse parses the strings and validates the values are in reason.
func Parse(page string, rowsPerPage string) (Page, error) {
	number := 1

	if page != "" {
		var err error

		number, err = strconv.Atoi(page)
		if err != nil {
			return Page{}, fmt.Errorf("page conversion: %w", err)
		}
	}

	rows := 10

	if rowsPerPage != "" {
		var err error

		rows, err = strconv.Atoi(rowsPerPage)
		if err != nil {
			return Page{}, fmt.Errorf("rows conversion: %w", err)
		}
	}

	if number <= 0 {
		return Page{}, errors.New("page value too small, must be larger than 0")
	}

	if rows <= 0 {
		return Page{}, errors.New("rows value too small, must be larger than 0")
	}

	if rows > 100 {
		return Page{}, errors.New("rows value too large, must be less than 100")
	}

	p := Page{
		number: number,
		rows:   rows,
	}

	return p, nil
}

// PageMustParse creates a paging value for testing.
func PageMustParse(page string, rowsPerPage string) Page {
	pg, err := Parse(page, rowsPerPage)
	if err != nil {
		panic(err)
	}

	return pg
}
