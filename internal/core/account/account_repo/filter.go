package account_repo

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Housiadas/cerberus/internal/core/account"
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

func applyFilter(filter account.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	wc := []string{"deleted_at IS NULL"}

	if filter.ID != nil {
		data["id"] = *filter.ID

		wc = append(wc, "id = :id")
	}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", *filter.Name)

		wc = append(wc, "name LIKE :name")
	}

	if filter.StripeCustomerID != nil {
		data["stripe_customer_id"] = *filter.StripeCustomerID

		wc = append(wc, "stripe_customer_id = :stripe_customer_id")
	}

	buf.WriteString(" WHERE ")
	buf.WriteString(strings.Join(wc, " AND "))
}
