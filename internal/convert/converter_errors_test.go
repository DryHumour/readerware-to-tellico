package convert

import (
	"errors"
	"testing"
)

func TestRowError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  RowError
		want string
	}{
		{
			name: "with column",
			err:  RowError{Line: 5, Column: "title", Err: errors.New("invalid")},
			want: "5:title: invalid",
		},
		{
			name: "without column",
			err:  RowError{Line: 10, Err: errors.New("parse error")},
			want: "10: parse error",
		},
		{
			name: "with record",
			err:  RowError{Record: 4, Line: 6, Err: errors.New("parse error")},
			want: "record 4 (line 6): parse error",
		},
		{
			name: "with record and column",
			err:  RowError{Record: 4, Line: 6, Column: "title", Err: errors.New("invalid")},
			want: "record 4 (line 6):title: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("RowError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRowError_Unwrap(t *testing.T) {
	t.Parallel()

	inner := errors.New("inner error")
	err := RowError{Line: 1, Err: inner}

	if got := err.Unwrap(); got != inner { //nolint:errorlint // we expect exactly the inner error
		t.Errorf("RowError.Unwrap() = %v, want %v", got, inner)
	}
}

func TestWrapColumnErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		line int
		err  error
		want error
	}{
		{
			name: "nil error",
			line: 1,
			err:  nil,
			want: nil,
		},
		{
			name: "single ColumnError",
			line: 5,
			err:  ColumnError{Column: "title", Err: errors.New("invalid")},
			want: RowError{Line: 5, Column: "title", Err: ColumnError{Column: "title", Err: errors.New("invalid")}},
		},
		{
			name: "non-ColumnError",
			line: 10,
			err:  errors.New("other error"),
			want: errors.New("other error"),
		},
		{
			name: "multi-error with ColumnErrors",
			line: 3,
			err: errors.Join(
				ColumnError{Column: "title", Err: errors.New("invalid")},
				ColumnError{Column: "author", Err: errors.New("missing")},
				errors.New("other"),
			),
			want: errors.Join(
				RowError{Line: 3, Column: "title", Err: ColumnError{Column: "title", Err: errors.New("invalid")}},
				RowError{Line: 3, Column: "author", Err: ColumnError{Column: "author", Err: errors.New("missing")}},
				errors.New("other"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := wrapColumnErrors(0, tt.line, tt.err)
			if (got == nil) != (tt.want == nil) {
				t.Errorf("wrapColumnErrors() = %v, want %v", got, tt.want)
				return
			}
			if got == nil || tt.want == nil {
				return
			}
			if got.Error() != tt.want.Error() {
				t.Errorf("wrapColumnErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}
