package mock

import (
	"fmt"

	"github.com/rikatz/kubepug/pkg/apis/v1alpha1"
	"github.com/rikatz/kubepug/pkg/results"
)

var (
	DeprecatedMock = []results.ResultItem{
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

	DeletedMock = []results.ResultItem{
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

type Store struct {
	deprecated  []results.ResultItem
	deleted     []results.ResultItem
	shouldError bool
}

func NewMockStore(loadData, shoulderr bool) *Store {
	m := &Store{}
	if loadData {
		m.deleted = DeletedMock
		m.deprecated = DeprecatedMock
	}

	m.shouldError = shoulderr
	return m
}

func (m *Store) GetDeprecations() (deprecated, deleted []results.ResultItem, err error) {
	if m.shouldError {
		return nil, nil, fmt.Errorf("something weird happened")
	}
	return m.deprecated, m.deleted, nil
}
