package tax_rate_repo

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Housiadas/cerberus/internal/core/domain/tax_rate"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
)

func applyCursor(cur cursor.Cursor, orderBy order.By, data map[string]any, buf *bytes.Buffer) {
	if !cur.HasCursor() {
		return
	}

	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return
	}

	data["cursor_value"] = cur.FieldValue()
	data["cursor_id"] = cur.ID()

	op := ">"
	if orderBy.Direction == order.DESC {
		op = "<"
	}

	fmt.Fprintf(buf, " AND (%s, id) %s (:cursor_value, :cursor_id)", by, op)
}

func applyFilter(filter tax_rate.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	wc := make([]string, 0)

	if filter.ID != nil {
		data["id"] = *filter.ID

		wc = append(wc, "id = :id")
	}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", *filter.Name)

		wc = append(wc, "name LIKE :name")
	}

	if filter.Country != nil {
		data["country"] = filter.Country.String()

		wc = append(wc, "country = :country")
	}

	if filter.IsActive != nil {
		data["is_active"] = *filter.IsActive

		wc = append(wc, "is_active = :is_active")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
