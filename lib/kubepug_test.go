package lib

import (
	"reflect"
	"testing"
)

func TestNewKubepug(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		want    *Kubepug
		wantErr bool
	}{
		{
			name:    "null is an error",
			config:  nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:   "empty config should work",
			config: &Config{},
			want: &Kubepug{
				Config: &Config{},
			},
			wantErr: false,
		},
		{
			name: "some config should work",
			config: &Config{
				GeneratedStore: "https://lalala.com/data.json",
				K8sVersion:     "v1.21",
			},
			want: &Kubepug{
				Config: &Config{
					GeneratedStore: "https://lalala.com/data.json",
					K8sVersion:     "v1.21",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewKubepug(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKubepug() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKubepug() = %v, want %v", got, tt.want)
			}
		})
	}
}
