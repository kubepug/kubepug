package mock

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMockStore(t *testing.T) {
	tests := []struct {
		name      string
		loadData  bool
		shoulderr bool
		want      *Store
	}{
		{
			name:      "should error return a mock with error enabled",
			shoulderr: true,
			want: &Store{
				shouldError: true,
			},
		},
		{
			name:      "should error and load data return a mock with error enabled",
			shoulderr: true,
			loadData:  true,
			want: &Store{
				shouldError: true,
				deprecated:  DeprecatedMock,
				deleted:     DeletedMock,
			},
		},
		{
			name:      "load data return a mock with loaded data",
			shoulderr: false,
			loadData:  true,
			want: &Store{
				deprecated: DeprecatedMock,
				deleted:    DeletedMock,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storer := NewMockStore(tt.loadData, tt.shoulderr)
			if !reflect.DeepEqual(storer, tt.want) {
				t.Errorf("NewMockStore() = %v, want %v", storer, tt.want)
			}
			deprecated, deleted, err := storer.GetDeprecations()
			if tt.shoulderr {
				require.Error(t, err)
				return
			}
			require.Equal(t, DeprecatedMock, deprecated)
			require.Equal(t, DeletedMock, deleted)
		})
	}
}
