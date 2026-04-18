package mapper

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapSlice(t *testing.T) {
	tests := []struct {
		name string
		src  []int
		fn   func(int) string
		want []string
	}{
		{
			name: "empty slice",
			src:  []int{},
			fn:   strconv.Itoa,
			want: []string{},
		},
		{
			name: "single element",
			src:  []int{42},
			fn:   strconv.Itoa,
			want: []string{"42"},
		},
		{
			name: "multiple elements",
			src:  []int{1, 2, 3},
			fn:   strconv.Itoa,
			want: []string{"1", "2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapSlice(tt.src, tt.fn)
			assert.Equal(t, tt.want, got)
		})
	}
}
