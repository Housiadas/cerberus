package role_repo

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Housiadas/cerberus/internal/core/domain/role"
)

func applyFilter(filter role.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	wc := []string{"deleted_at IS NULL"}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", *filter.Name)

		wc = append(wc, "name LIKE :name")
	}

	buf.WriteString(" WHERE ")
	buf.WriteString(strings.Join(wc, " AND "))
}
