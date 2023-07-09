package errors

import (
	"errors"
	"testing"
)

func TestIsErrAPINotFound(t *testing.T) {
	tests := []struct {
		name string
		e    error
		want bool
	}{
		{
			name: "IsNotFoundErr true",
			e:    ErrAPINotFound,
			want: true,
		},
		{
			name: "IsNotFoundErr false",
			e:    errors.New("API not found"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsErrAPINotFound(tt.e); got != tt.want {
				t.Errorf("IsErrAPINotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
