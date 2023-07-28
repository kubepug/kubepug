package kubepug

import (
	"reflect"
	"testing"

	"github.com/rikatz/kubepug/pkg/results"
)

func TestGetDeprecations(t *testing.T) {
	tests := []struct {
		name       string
		d          Deprecator
		wantResult results.Result
		wantErr    bool
	}{
		// TODO: Implement Mock deprecations and add tests.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := GetDeprecations(tt.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeprecations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("GetDeprecations() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
