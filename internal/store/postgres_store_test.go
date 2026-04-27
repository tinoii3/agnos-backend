package store

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "postgres unique violation",
			err: &pgconn.PgError{
				Code: "23505",
			},
			want: true,
		},
		{
			name: "wrapped postgres unique violation",
			err:  fmt.Errorf("insert failed: %w", &pgconn.PgError{Code: "23505"}),
			want: true,
		},
		{
			name: "other postgres error code",
			err: &pgconn.PgError{
				Code: "23503",
			},
			want: false,
		},
		{
			name: "non postgres error",
			err:  errors.New("some generic error"),
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsUniqueViolation(tc.err)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
