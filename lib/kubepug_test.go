package lib

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/rikatz/kubepug/pkg/store/mock"
	"github.com/stretchr/testify/require"
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

func TestGetDeprecated(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/data.json" {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mock.MockValidData)) //nolint: errcheck
			}
			if r.URL.Path == "/notfound.json" {
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	defer ts.Close()

	t.Run("nil config should fail", func(t *testing.T) {
		pug := &Kubepug{}
		result, err := pug.GetDeprecated()
		require.ErrorContains(t, err, "config cannot be null")
		require.Nil(t, result)
	})

	t.Run("empty store should fail", func(t *testing.T) {
		pug := &Kubepug{
			Config: &Config{
				GeneratedStore: "",
			},
		}
		result, err := pug.GetDeprecated()
		require.Error(t, err)
		require.ErrorContains(t, err, "a database path should be provided")
		require.Nil(t, result)
	})

	t.Run("local file not found should fail", func(t *testing.T) {
		pug := &Kubepug{
			Config: &Config{
				GeneratedStore: "/tmp123/lalala",
			},
		}
		result, err := pug.GetDeprecated()
		require.Error(t, err)
		require.ErrorContains(t, err, "open /tmp123/lalala: no such file or directory")
		require.Nil(t, result)
	})

	t.Run("remote file not found should fail", func(t *testing.T) {
		pug := &Kubepug{
			Config: &Config{
				GeneratedStore: ts.URL + "/notfound.json",
			},
		}
		result, err := pug.GetDeprecated()
		require.Error(t, err)
		require.ErrorContains(t, err, "could not download the data file")
		require.Nil(t, result)
	})

	t.Run("invalid file input should fail", func(t *testing.T) {
		pug := &Kubepug{
			Config: &Config{
				GeneratedStore: ts.URL + "/data.json",
				K8sVersion:     "v1.22",
				Input:          "/tmp123/lslslasd",
			},
		}

		result, err := pug.GetDeprecated()
		require.Error(t, err)
		require.ErrorContains(t, err, "error reading file input: input location /tmp123/lslslasd does not exist")
		require.Nil(t, result)
	})

	t.Run("empty k8s config should fail", func(t *testing.T) {
		pug := &Kubepug{
			Config: &Config{
				GeneratedStore: ts.URL + "/data.json",
				K8sVersion:     "v1.22",
			},
		}

		result, err := pug.GetDeprecated()
		require.Error(t, err)
		require.ErrorContains(t, err, "k8s config cannot be null when k8s is being used")
		require.Nil(t, result)
	})
}
