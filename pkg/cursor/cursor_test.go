package cursor

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		cursor    string
		limit     string
		wantErr   bool
		wantLimit int
		hasCursor bool
	}{
		{
			name:      "empty cursor and limit defaults",
			cursor:    "",
			limit:     "",
			wantErr:   false,
			wantLimit: 10,
			hasCursor: false,
		},
		{
			name:      "custom limit",
			cursor:    "",
			limit:     "25",
			wantErr:   false,
			wantLimit: 25,
			hasCursor: false,
		},
		{
			name:    "limit too small",
			cursor:  "",
			limit:   "0",
			wantErr: true,
		},
		{
			name:    "limit too large",
			cursor:  "",
			limit:   "200",
			wantErr: true,
		},
		{
			name:    "invalid limit",
			cursor:  "",
			limit:   "abc",
			wantErr: true,
		},
		{
			name:      "valid cursor",
			cursor:    Encode("2024-01-01", "abc-123"),
			limit:     "10",
			wantErr:   false,
			wantLimit: 10,
			hasCursor: true,
		},
		{
			name:    "invalid base64 cursor",
			cursor:  "not-valid-base64!!!",
			limit:   "10",
			wantErr: true,
		},
		{
			name:    "invalid json cursor",
			cursor:  "bm90LWpzb24=", // "not-json" in base64
			limit:   "10",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.cursor, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got.Limit() != tt.wantLimit {
				t.Errorf("Limit() = %d, want %d", got.Limit(), tt.wantLimit)
			}

			if got.HasCursor() != tt.hasCursor {
				t.Errorf("HasCursor() = %v, want %v", got.HasCursor(), tt.hasCursor)
			}
		})
	}
}

func TestEncodeDecode(t *testing.T) {
	encoded := Encode("some-value", "some-id")

	cur, err := Parse(encoded, "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cur.FieldValue() != "some-value" {
		t.Errorf("FieldValue() = %q, want %q", cur.FieldValue(), "some-value")
	}

	if cur.ID() != "some-id" {
		t.Errorf("ID() = %q, want %q", cur.ID(), "some-id")
	}
}

func TestNewResult(t *testing.T) {
	type item struct {
		ID   string
		Name string
	}

	idFn := func(i item) string { return i.ID }
	fieldFn := func(i item) any { return i.Name }

	t.Run("empty data", func(t *testing.T) {
		cur, _ := Parse("", "10")
		result := NewResult([]item{}, 10, cur, orderBy(), idFn, fieldFn)

		if len(result.Data) != 0 {
			t.Error("expected empty data")
		}

		if result.Metadata.HasMore {
			t.Error("expected HasMore=false")
		}
	})

	t.Run("has more", func(t *testing.T) {
		items := make([]item, 11)
		for i := range items {
			items[i] = item{ID: string(rune('a' + i)), Name: "name"}
		}

		cur, _ := Parse("", "10")
		result := NewResult(items, 10, cur, orderBy(), idFn, fieldFn)

		if len(result.Data) != 10 {
			t.Errorf("expected 10 items, got %d", len(result.Data))
		}

		if !result.Metadata.HasMore {
			t.Error("expected HasMore=true")
		}

		if result.Metadata.NextCursor == "" {
			t.Error("expected NextCursor to be set")
		}
	})

	t.Run("prev cursor set when cursor provided", func(t *testing.T) {
		items := []item{{ID: "a", Name: "x"}}
		cur, _ := Parse(Encode("val", "prev-id"), "10")
		result := NewResult(items, 10, cur, orderBy(), idFn, fieldFn)

		if result.Metadata.PrevCursor == "" {
			t.Error("expected PrevCursor to be set")
		}
	})
}

func orderBy() struct {
	Field     string
	Direction string
} {
	return struct {
		Field     string
		Direction string
	}{Field: "name", Direction: "ASC"}
}
