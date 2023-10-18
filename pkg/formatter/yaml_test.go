package formatter

import (
	"reflect"
	"testing"

	yamlencoder "gopkg.in/yaml.v3"

	"github.com/kubepug/kubepug/pkg/results"
)

func Test_yaml_Output(t *testing.T) {
	tests := []struct {
		name    string
		f       *yaml
		data    results.Result
		wantErr bool
	}{
		{
			name:    "some yaml data",
			data:    mockResult,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &yaml{}
			got, err := f.Output(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("yaml.Output() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			roundTripData := results.Result{}
			err = yamlencoder.Unmarshal(got, &roundTripData)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(tt.data, roundTripData) {
				t.Errorf("yaml.Output() = %v, want %v", got, roundTripData)
			}
		})
	}
}
