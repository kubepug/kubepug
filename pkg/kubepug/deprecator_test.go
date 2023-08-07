package kubepug

import (
	"fmt"
	"testing"

	"github.com/rikatz/kubepug/pkg/apis/v1alpha1"
	"github.com/rikatz/kubepug/pkg/results"
	"github.com/stretchr/testify/require"
)

var (
	deprecatedMock = []results.ResultItem{
		{
			Group:   "something.anything",
			Version: "v1beta1",
			Kind:    "blah",
			Replacement: &v1alpha1.GroupVersionKind{
				Group:   "something.anything",
				Version: "v1",
				Kind:    "blah",
			},
			K8sVersion: "v1.22",
			Items: []results.Item{
				{
					Scope:      "Object",
					Namespace:  "default",
					ObjectName: "bloh",
				},
			},
		},
	}

	deletedMock = []results.ResultItem{
		{
			Group:   "something.anything",
			Version: "v1alpha1",
			Kind:    "blah",
			Replacement: &v1alpha1.GroupVersionKind{
				Group:   "something.anything",
				Version: "v1",
				Kind:    "blah",
			},
			K8sVersion: "v1.19",
			Items: []results.Item{
				{
					Scope:      "Object",
					Namespace:  "default",
					ObjectName: "bloh",
				},
			},
		},
	}
)

type mockStore struct {
	deprecated  []results.ResultItem
	deleted     []results.ResultItem
	shouldError bool
}

func (m *mockStore) GetDeprecations() (deprecated, deleted []results.ResultItem, err error) {
	if m.shouldError {
		return nil, nil, fmt.Errorf("something weird happened")
	}
	return m.deprecated, m.deleted, nil
}

func TestGetDeprecations(t *testing.T) {
	t.Run("should return an error", func(t *testing.T) {
		store := &mockStore{shouldError: true}
		result, err := GetDeprecations(store)
		require.Error(t, err)
		require.Empty(t, result.DeletedAPIs)
		require.Empty(t, result.DeprecatedAPIs)
	})

	t.Run("should return correctly", func(t *testing.T) {
		store := &mockStore{
			shouldError: false,
			deprecated:  deprecatedMock,
			deleted:     deletedMock,
		}
		result, err := GetDeprecations(store)
		require.NoError(t, err)
		require.Equal(t, deletedMock, result.DeletedAPIs)
		require.Equal(t, deprecatedMock, result.DeprecatedAPIs)
	})
}
