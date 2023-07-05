package utils

import (
	"testing"
)

func Test_ShouldParse(t *testing.T) {
	tests := []struct {
		name         string
		group        string
		ignoregroup  []string
		includegroup []string
		want         bool
	}{
		{
			name: "core apis should not be ignored",
			want: true,
		},
		{
			name:  "core apis should not be ignored",
			group: "",
			want:  true,
		},
		{
			name:         "ignored apis should be ignored",
			group:        "something.k8s.io",
			ignoregroup:  []string{"something.k8s.io"},
			includegroup: []string{"k8s.io"},
			want:         false,
		},
		{
			name:         "no group configured should not be ignored",
			group:        "something.k8s.io",
			ignoregroup:  []string{},
			includegroup: []string{},
			want:         true,
		},
		{
			name:         "included group should not be ignored",
			group:        "bla.random.api",
			ignoregroup:  []string{},
			includegroup: []string{"random.api"},
			want:         true,
		},
		{
			name:         "api outside included group should not be parsed",
			group:        "other.api",
			ignoregroup:  []string{},
			includegroup: []string{"random.api"},
			want:         false,
		},
		{
			name:         "multiple APIs to be ignored should result on ignored API",
			group:        "gateway.x-k8s.io",
			ignoregroup:  []string{"externaldns.k8s.io", "x-k8s.io"},
			includegroup: []string{".k8s.io"},
			want:         false,
		},
		{
			name:         "v1 internal groups should not be ignored",
			group:        "apps",
			ignoregroup:  []string{"externaldns.k8s.io", "x-k8s.io"},
			includegroup: []string{".k8s.io"},
			want:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldParse(tt.group, tt.ignoregroup, tt.includegroup); got != tt.want {
				t.Errorf("ShouldParse() = %v, want %v", got, tt.want)
			}
		})
	}
}
