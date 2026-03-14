package invoice_repo

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Housiadas/cerberus/internal/core/domain/invoice"
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

func applyFilter(filter invoice.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	wc := make([]string, 0)

	if filter.ID != nil {
		data["id"] = *filter.ID

		wc = append(wc, "id = :id")
	}

	if filter.AccountID != nil {
		data["account_id"] = *filter.AccountID

		wc = append(wc, "account_id = :account_id")
	}

	if filter.Status != nil {
		data["status"] = filter.Status.String()

		wc = append(wc, "status = :status")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
