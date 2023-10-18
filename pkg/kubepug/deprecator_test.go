package kubepug

import (
	"testing"

	mock "github.com/kubepug/kubepug/pkg/store/mock"
	"github.com/stretchr/testify/require"
)

func TestGetDeprecations(t *testing.T) {
	t.Run("should return an error", func(t *testing.T) {
		store := mock.NewMockStore(true, true)
		result, err := GetDeprecations(store)
		require.Error(t, err)
		require.Empty(t, result.DeletedAPIs)
		require.Empty(t, result.DeprecatedAPIs)
	})

	t.Run("should return correctly", func(t *testing.T) {
		store := mock.NewMockStore(true, false)
		result, err := GetDeprecations(store)
		require.NoError(t, err)
		require.Equal(t, mock.DeletedMock, result.DeletedAPIs)
		require.Equal(t, mock.DeprecatedMock, result.DeprecatedAPIs)
	})
}
