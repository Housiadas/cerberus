package page

import (
	"fmt"
	"strconv"
)

const (
	rowsPerPageDefault = 10
	rowsPerPageMax     = 100
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

// Page represents the requested page and rows per page.
type Page struct {
	number int
	rows   int
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

	rows := rowsPerPageDefault

	if rowsPerPage != "" {
		var err error

		rows, err = strconv.Atoi(rowsPerPage)
		if err != nil {
			return Page{}, fmt.Errorf("rows conversion: %w", err)
		}
	}

	if number <= 0 {
		return Page{}, ErrPageTooSmall
	}

	if rows <= 0 {
		return Page{}, ErrRowsTooSmall
	}

	if rows > rowsPerPageMax {
		return Page{}, ErrRowsTooLarge
	}

	p := Page{
		number: number,
		rows:   rows,
	}

	return p, nil
}

// MustParse creates a paging value for testing.
func MustParse(page string, rowsPerPage string) Page {
	pg, err := Parse(page, rowsPerPage)
	if err != nil {
		panic(err)
	}

	return pg
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
