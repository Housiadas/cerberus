package clock

import "time"

func Format(t *time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
