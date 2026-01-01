package web

import (
	"reflect"
	"testing"
)

func Test_calculateMetadata(t *testing.T) {
	type args struct {
		total int
		page  int
		rows  int
	}
	tests := []struct {
		name string
		args args
		want Metadata
	}{
		{
			name: "Metadata empty example",
			args: args{
				total: 0,
				page:  0,
				rows:  0,
			},
			want: Metadata{},
		},
		{
			name: "Metadata example, current page 1",
			args: args{
				total: 150,
				page:  1,
				rows:  8,
			},
			want: Metadata{
				FirstPage:   1,
				CurrentPage: 1,
				LastPage:    19,
				RowsPerPage: 8,
				Total:       150,
			},
		},
		{
			name: "Metadata example, current page 5",
			args: args{
				total: 150,
				page:  5,
				rows:  8,
			},
			want: Metadata{
				FirstPage:   1,
				CurrentPage: 5,
				LastPage:    19,
				RowsPerPage: 8,
				Total:       150,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateMetadata(tt.args.total, tt.args.page, tt.args.rows); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Page_Parse(t *testing.T) {
	type args struct {
		page        string
		rowsPerPage string
	}
	tests := []struct {
		name    string
		args    args
		want    Page
		wantErr bool
	}{
		{
			name: "Invalid numeric",
			args: args{
				page:        "test",
				rowsPerPage: "test",
			},
			want:    Page{},
			wantErr: true,
		},
		{
			name: "Valid numeric",
			args: args{
				page:        "2",
				rowsPerPage: "10",
			},
			want: Page{
				number: 2,
				rows:   10,
			},
			wantErr: false,
		},
		{
			name: "Page negative number",
			args: args{
				page:        "-2",
				rowsPerPage: "10",
			},
			want:    Page{},
			wantErr: true,
		},
		{
			name: "Rows number too big",
			args: args{
				page:        "1",
				rowsPerPage: "1000",
			},
			want:    Page{},
			wantErr: true,
		},
		{
			name: "Rows number too small",
			args: args{
				page:        "1",
				rowsPerPage: "-1",
			},
			want:    Page{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.page, tt.args.rowsPerPage)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}
