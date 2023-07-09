package formatter

import (
	jsonencoding "encoding/json"
	"reflect"
	"testing"

	"github.com/rikatz/kubepug/pkg/results"
)

func Test_json_Output(t *testing.T) {
	tests := []struct {
		name    string
		data    results.Result
		wantErr bool
	}{
		{
			name:    "some json data",
			data:    mockResult,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &json{}
			got, err := f.Output(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Output() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			roundTripData := results.Result{}
			err = jsonencoding.Unmarshal(got, &roundTripData)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(tt.data, roundTripData) {
				t.Errorf("json.Output() = %v, want %v", got, roundTripData)
			}
		})
	}
}
