package results

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCreateItem(t *testing.T) {
	tests := []struct {
		name    string
		group   string
		version string
		kind    string
		items   []Item
		want    ResultItem
	}{
		{
			name:    "return valid item",
			group:   "somegroup",
			version: "someversion",
			kind:    "SomeKind",
			items: []Item{
				{
					Scope:      "Object",
					ObjectName: "myobj1",
					Namespace:  "somens",
					Location:   "/some/location",
				},
				{
					Scope:      "Global",
					ObjectName: "myobj2",
					Location:   "/some/location2",
				},
			},
			want: ResultItem{
				Group:   "somegroup",
				Version: "someversion",
				Kind:    "SomeKind",
				Items: []Item{
					{
						Scope:      "Object",
						ObjectName: "myobj1",
						Namespace:  "somens",
						Location:   "/some/location",
					},
					{
						Scope:      "Global",
						ObjectName: "myobj2",
						Location:   "/some/location2",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateItem(tt.group, tt.version, tt.kind, tt.items); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListObjects(t *testing.T) {
	tests := []struct {
		name                string
		items               []unstructured.Unstructured
		wantDeprecatedItems []Item
	}{
		{
			name: "test items",
			items: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"name":      "scoped",
							"namespace": "somens",
						},
					},
				},
				{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"name": "globalobj",
						},
					},
				},
			},
			wantDeprecatedItems: []Item{
				{
					ObjectName: "scoped",
					Namespace:  "somens",
					Scope:      namespacedObject,
				},
				{
					ObjectName: "globalobj",
					Scope:      clusterObject,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDeprecatedItems := ListObjects(tt.items); !reflect.DeepEqual(gotDeprecatedItems, tt.wantDeprecatedItems) {
				t.Errorf("ListObjects() = %v, want %v", gotDeprecatedItems, tt.wantDeprecatedItems)
			}
		})
	}
}
