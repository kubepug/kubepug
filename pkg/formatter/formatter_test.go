package formatter

import (
	"reflect"
	"testing"

	"github.com/rikatz/kubepug/pkg/results"
)

var mockResult = results.Result{
	DeprecatedAPIs: []results.ResultItem{
		{
			Group:   "somegroup",
			Version: "v3",
			Kind:    "SomeKind",
			Items: []results.Item{
				{
					Scope:      "Object",
					ObjectName: "myobj",
					Namespace:  "somens",
					Location:   "/some/location",
				},
			},
		},
	},
	DeletedAPIs: []results.ResultItem{
		{
			Group:   "somegroup2",
			Version: "v4",
			Kind:    "SomeKind1",
			Items: []results.Item{
				{
					Scope:      "Object",
					ObjectName: "myobj2",
					Namespace:  "somens3",
					Location:   "/some/location3",
				},
			},
		},
	},
}

func TestNewFormatterWithError(t *testing.T) {
	tests := []struct {
		name          string
		formattertype string
		want          Formatter
		wantErr       bool
	}{
		{
			name:          "invalid is an error",
			formattertype: "bla",
			want:          nil,
			wantErr:       true,
		},
		{
			name:          "stdout is valid",
			formattertype: "stdout",
			want:          &stdout{},
			wantErr:       false,
		},
		{
			name:          "plain is valid",
			formattertype: "plain",
			want:          &plain{},
			wantErr:       false,
		},
		{
			name:          "json is valid",
			formattertype: "json",
			want:          &json{},
			wantErr:       false,
		},
		{
			name:          "yaml is valid",
			formattertype: "yaml",
			want:          &yaml{},
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFormatterWithError(tt.formattertype)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFormatterWithError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFormatterWithError() = %v, want %v", got, tt.want)
			}
		})
	}
}
