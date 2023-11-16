package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	someRandomContent = "testing download"
	dataJSON          = "/data.json"
)

func TestDownloadGeneratedJSON(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == dataJSON {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(someRandomContent)) //nolint: errcheck
			}
			if r.URL.Path == "/notfound.json" {
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	defer ts.Close()

	t.Run("force bad tempdir", func(t *testing.T) {
		// Force OSEnv to wrong tmpdir
		err := os.Setenv("TMPDIR", "/xpto123lalala")
		t.Cleanup(func() {
			err = os.Setenv("TMPDIR", "")
			require.NoError(t, err)
		})
		require.NoError(t, err)
		f, err := DownloadGeneratedJSON(ts.URL + dataJSON)
		require.Error(t, err)
		require.Empty(t, f)
	})

	t.Run("download valid data", func(t *testing.T) {
		f, err := DownloadGeneratedJSON(ts.URL + dataJSON)
		require.NoError(t, err)
		require.Contains(t, f, "kubepug")
		require.Contains(t, f, "data.json")
	})

	t.Run("error on invalid data", func(t *testing.T) {
		f, err := DownloadGeneratedJSON(ts.URL + "/notfound.json")
		require.Error(t, err)
		require.Empty(t, f)
	})

	t.Run("error on invalid url", func(t *testing.T) {
		f, err := DownloadGeneratedJSON("http://127.0.0.1:xpto1")
		require.Error(t, err)
		require.Empty(t, f)
	})

	t.Run("try to create file on a forbidden place", func(t *testing.T) {
		err := downloadFile("/tmp123/xpto", ts.URL+dataJSON)
		require.Error(t, err)
	})
}
