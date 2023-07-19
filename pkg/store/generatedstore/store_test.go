package generatedstore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Masterminds/semver/v3"
	apis "github.com/rikatz/kubepug/pkg/apis/v1alpha1"
	"github.com/stretchr/testify/require"
)

var (
	mockValidData = `
	[
    {
        "group": "extensions",
        "version": "v1beta1",
        "kind": "DaemonSet",
        "description": "DEPRECATED - This group version of DaemonSet is deprecated by apps/v1beta2/DaemonSet. See the release notes for\nmore information.\nDaemonSet represents the configuration of a daemon set.",
        "introduced_version": {
            "version_major": 1,
            "version_minor": 1
        },
        "deprecated_version": {
            "version_major": 1,
            "version_minor": 8
        },
        "removed_version": {
            "version_major": 1,
            "version_minor": 16
        },
        "replacement": {
            "group": "apps",
            "version": "v1",
            "kind": "DaemonSet"
        }
    },
	{
        "group": "",
        "version": "v1",
        "kind": "BlahPod",
        "description": "BlahPod is a deprecated Pod.",
        "introduced_version": {
            "version_major": 1,
            "version_minor": 1
        },
        "deprecated_version": {
            "version_major": 1,
            "version_minor": 8
        },
        "removed_version": {
            "version_major": 1,
            "version_minor": 16
        },
        "replacement": {
            "group": "",
            "version": "v1",
            "kind": "Pod"
        }
    },
    {
        "group": "discovery.k8s.io",
        "version": "v1beta1",
        "kind": "EndpointSliceList",
        "description": "EndpointSliceList represents a list of endpoint slices",
        "introduced_version": {
            "version_major": 1,
            "version_minor": 16
        },
        "deprecated_version": {
            "version_major": 1,
            "version_minor": 21
        },
        "removed_version": {
            "version_major": 1,
            "version_minor": 25
        },
        "replacement": {
            "group": "discovery.k8s.io",
            "version": "v1",
            "kind": "EndpointSlice"
        }
    },
    {
        "group": "admission.k8s.io",
        "version": "v1beta1",
        "kind": "AdmissionReview",
        "description": "AdmissionReview describes an admission review request/response.",
        "introduced_version": {
            "version_major": 1,
            "version_minor": 9
        },
        "deprecated_version": {
            "version_major": 1,
            "version_minor": 19
        },
        "removed_version": {
            "version_major": 1,
            "version_minor": 22
        },
        "replacement": {
            "group": "admission.k8s.io",
            "version": "v1",
            "kind": "AdmissionReview"
        }
    }
]`

	mockInvalidData = `
[
    {
        "group": "extensions",
        "version": "v1beta1",
        "kind": "DaemonSet",
        "description": "DEPRECATED - This group version of DaemonSet is deprecated by apps/v1beta2/DaemonSet. See the release notes for\nmore information.\nDaemonSet represents the configuration of a daemon set.",
        "introduced_version": {
            "version_major": 1,
            "version_minor": 1
        },
        "replacement": {
            "group": "apps",
            "version": "v1",
            "kind": "DaemonSet"
        `
	mockNoVersions = `
[
    {
        "group": "extensions",
        "version": "v1beta1",
        "kind": "DaemonSet",
        "description": "DEPRECATED - This group version of DaemonSet is deprecated by apps/v1beta2/DaemonSet. See the release notes for\nmore information.\nDaemonSet represents the configuration of a daemon set.",
        "deprecated_version": {
            "version_major": 0,
            "version_minor": 5
        },
        "removed_version": {
            "version_major": 7,
            "version_minor": 0
        }
    }
]`
)

func Test_generateVersion(t *testing.T) {
	tests := []struct {
		name  string
		major int
		minor int
		want  string
	}{
		{
			name:  "major 0 should return empty",
			major: 0,
			minor: 5,
			want:  "",
		},
		{
			name:  "minor 0 should return empty",
			major: 7,
			minor: 0,
			want:  "",
		},
		{
			name:  "should return the right version",
			major: 1,
			minor: 25,
			want:  "1.25",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateVersion(tt.major, tt.minor); got != tt.want {
				t.Errorf("generateVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPopulateStruct(t *testing.T) {
	t.Run("with invalid json file", func(t *testing.T) {
		_, err := newInternalDatabase([]byte(mockInvalidData))
		require.Error(t, err)
	})

	t.Run("with valid json file", func(t *testing.T) {
		v, err := newInternalDatabase([]byte(mockValidData))
		require.NoError(t, err)

		require.Equal(t, v["extensions"]["DaemonSet"]["v1beta1"].DeprecationVersion, "1.8")
		require.Equal(t, v["extensions"]["DaemonSet"]["v1beta1"].DeletedVersion, "1.16")
		require.Equal(t, v["extensions"]["DaemonSet"]["v1beta1"].IntroducedVersion, "1.1")

		require.Equal(t,
			"DEPRECATED - This group version of DaemonSet is deprecated by apps/v1beta2/DaemonSet. See the release notes for\nmore information.\nDaemonSet represents the configuration of a daemon set.",
			v["extensions"]["DaemonSet"]["v1beta1"].Description)

		require.Equal(t,
			&apis.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"},
			v["extensions"]["DaemonSet"]["v1beta1"].Replacement)
	})

	t.Run("without versions and replacements json file", func(t *testing.T) {
		v, err := newInternalDatabase([]byte(mockNoVersions))
		require.NoError(t, err)

		require.Equal(t, v["extensions"]["DaemonSet"]["v1beta1"].DeprecationVersion, "")
		require.Equal(t, v["extensions"]["DaemonSet"]["v1beta1"].DeletedVersion, "")
		require.Equal(t,
			"DEPRECATED - This group version of DaemonSet is deprecated by apps/v1beta2/DaemonSet. See the release notes for\nmore information.\nDaemonSet represents the configuration of a daemon set.",
			v["extensions"]["DaemonSet"]["v1beta1"].Description)

		require.Nil(t,
			v["extensions"]["DaemonSet"]["v1beta1"].Replacement)
	})
}

func TestNewStoreFromHTTP(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/data.json" {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockValidData))
			}
			if r.URL.Path == "/datainvalid.json" {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockInvalidData))
			}
			if r.URL.Path == "/notfound.json" {
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	defer ts.Close()

	t.Run("with invalid path should fail", func(t *testing.T) {
		_, err := NewGeneratedStore(StoreConfig{Path: ts.URL + "/notfound.json"})
		require.Error(t, err)
	})
	t.Run("with invalid file content should fail", func(t *testing.T) {
		_, err := NewGeneratedStore(StoreConfig{Path: ts.URL + "/datainvalid.json"})
		require.Error(t, err)
	})

	t.Run("with valid remote file content should parse", func(t *testing.T) {
		v, err := NewGeneratedStore(StoreConfig{Path: ts.URL + "/data.json"})
		require.NoError(t, err)

		require.Equal(t, v.db["extensions"]["DaemonSet"]["v1beta1"].DeprecationVersion, "1.8")
		require.Equal(t, v.db["extensions"]["DaemonSet"]["v1beta1"].DeletedVersion, "1.16")
		require.Equal(t, v.db["extensions"]["DaemonSet"]["v1beta1"].IntroducedVersion, "1.1")

		require.Equal(t,
			"DEPRECATED - This group version of DaemonSet is deprecated by apps/v1beta2/DaemonSet. See the release notes for\nmore information.\nDaemonSet represents the configuration of a daemon set.",
			v.db["extensions"]["DaemonSet"]["v1beta1"].Description)

		require.Equal(t,
			&apis.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"},
			v.db["extensions"]["DaemonSet"]["v1beta1"].Replacement)

		require.Equal(t, v.db["admission.k8s.io"]["AdmissionReview"]["v1beta1"].DeprecationVersion, "1.19")
		require.Equal(t, v.db["admission.k8s.io"]["AdmissionReview"]["v1beta1"].DeletedVersion, "1.22")
		require.Equal(t, v.db["admission.k8s.io"]["AdmissionReview"]["v1beta1"].IntroducedVersion, "1.9")

		require.Equal(t,
			"AdmissionReview describes an admission review request/response.",
			v.db["admission.k8s.io"]["AdmissionReview"]["v1beta1"].Description)

		require.Equal(t,
			&apis.GroupVersionKind{Group: "admission.k8s.io", Version: "v1", Kind: "AdmissionReview"},
			v.db["admission.k8s.io"]["AdmissionReview"]["v1beta1"].Replacement)
	})

}

func TestNewStoreFromFile(t *testing.T) {

	t.Run("with no file path should return an error", func(t *testing.T) {
		_, err := NewGeneratedStore(StoreConfig{Path: ""})
		require.Error(t, err)
	})
	t.Run("with invalid file path should return an error", func(t *testing.T) {
		_, err := NewGeneratedStore(StoreConfig{Path: "/xpto/blabla/123"})
		require.Error(t, err)
	})

	t.Run("with valid path should be able to parse it", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", "")
		require.NoError(t, err)
		err = os.WriteFile(tmp+"/testfile", []byte(mockValidData), 0o600)
		require.NoError(t, err)
		v, err := NewGeneratedStore(StoreConfig{Path: tmp + "/testfile", MinVersion: "v1.20"})
		require.NoError(t, err)

		results, err := v.GetAPIDefinition(context.TODO(), "admission.k8s.io", "v1beta1", "AdmissionReview")
		require.NoError(t, err)
		require.Equal(t, "", results.DeletedVersion)
		require.Equal(t, "1.19", results.DeprecationVersion)
		require.Equal(t, "1.9", results.IntroducedVersion)
		require.Equal(t,
			results.Replacement,
			&apis.GroupVersionKind{Group: "admission.k8s.io", Version: "v1", Kind: "AdmissionReview"})
		require.Equal(t,
			"AdmissionReview describes an admission review request/response.", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "", "v1", "BlahPod")
		require.NoError(t, err)
		require.Equal(t, "1.16", results.DeletedVersion)
		require.Equal(t, "1.8", results.DeprecationVersion)
		require.Equal(t, "1.1", results.IntroducedVersion)
		require.Equal(t,
			results.Replacement,
			&apis.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"})
		require.Equal(t,
			"BlahPod is a deprecated Pod.", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "networking.k8s.io", "v1", "NetworkPolicy")
		require.NoError(t, err)
		require.Equal(t, "", results.DeletedVersion)
		require.Equal(t, "", results.DeprecationVersion)
		require.Equal(t, "", results.IntroducedVersion)
		require.Equal(t,
			"", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "extensions", "v2", "Unknown")
		require.NoError(t, err)
		require.Equal(t, "", results.DeletedVersion)
		require.Equal(t, "", results.DeprecationVersion)
		require.Equal(t, "", results.IntroducedVersion)
		require.Equal(t,
			"", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "extensions", "v2", "DaemonSet")
		require.NoError(t, err)
		require.Equal(t, "", results.DeletedVersion)
		require.Equal(t, "", results.DeprecationVersion)
		require.Equal(t, "", results.IntroducedVersion)
		require.Equal(t,
			"", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "extensions", "v2", "DaemonSet")
		require.NoError(t, err)
		require.Equal(t, "", results.DeletedVersion)
		require.Equal(t, "", results.DeprecationVersion)
		require.Equal(t, "", results.IntroducedVersion)
		require.Equal(t,
			"", results.Description)

	})
}

func TestNewStoreFromBytes(t *testing.T) {
	t.Run("with invalid bytes should return an error", func(t *testing.T) {
		_, err := NewGeneratedStoreFromBytes([]byte(mockInvalidData), StoreConfig{})
		require.Error(t, err)
	})

	t.Run("with invalid version should return an error", func(t *testing.T) {
		_, err := NewGeneratedStoreFromBytes([]byte(mockValidData), StoreConfig{MinVersion: "xxxx"})
		require.Error(t, err)
	})

	t.Run("with valid bytes should be able to parse it", func(t *testing.T) {
		v, err := NewGeneratedStoreFromBytes([]byte(mockValidData), StoreConfig{})
		require.NoError(t, err)
		require.Equal(t, v.db["extensions"]["DaemonSet"]["v1beta1"].DeprecationVersion, "1.8")
		require.Equal(t, v.db["extensions"]["DaemonSet"]["v1beta1"].DeletedVersion, "1.16")
		require.Equal(t, v.db["extensions"]["DaemonSet"]["v1beta1"].IntroducedVersion, "1.1")

		require.Equal(t,
			"DEPRECATED - This group version of DaemonSet is deprecated by apps/v1beta2/DaemonSet. See the release notes for\nmore information.\nDaemonSet represents the configuration of a daemon set.",
			v.db["extensions"]["DaemonSet"]["v1beta1"].Description)

		require.Equal(t,
			&apis.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"},
			v.db["extensions"]["DaemonSet"]["v1beta1"].Replacement)

		require.Equal(t, v.db["admission.k8s.io"]["AdmissionReview"]["v1beta1"].DeprecationVersion, "1.19")
		require.Equal(t, v.db["admission.k8s.io"]["AdmissionReview"]["v1beta1"].DeletedVersion, "1.22")
		require.Equal(t, v.db["admission.k8s.io"]["AdmissionReview"]["v1beta1"].IntroducedVersion, "1.9")

		require.Equal(t,
			"AdmissionReview describes an admission review request/response.",
			v.db["admission.k8s.io"]["AdmissionReview"]["v1beta1"].Description)

		require.Equal(t,
			&apis.GroupVersionKind{Group: "admission.k8s.io", Version: "v1", Kind: "AdmissionReview"},
			v.db["admission.k8s.io"]["AdmissionReview"]["v1beta1"].Replacement)

		results, err := v.GetAPIDefinition(context.TODO(), "admission.k8s.io", "v1beta1", "AdmissionReview")
		require.NoError(t, err)
		require.Equal(t, "1.22", results.DeletedVersion)
		require.Equal(t, "1.19", results.DeprecationVersion)
		require.Equal(t, "1.9", results.IntroducedVersion)
		require.Equal(t,
			results.Replacement,
			&apis.GroupVersionKind{Group: "admission.k8s.io", Version: "v1", Kind: "AdmissionReview"})
		require.Equal(t,
			"AdmissionReview describes an admission review request/response.", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "", "v1", "BlahPod")
		require.NoError(t, err)
		require.Equal(t, "1.16", results.DeletedVersion)
		require.Equal(t, "1.8", results.DeprecationVersion)
		require.Equal(t, "1.1", results.IntroducedVersion)
		require.Equal(t,
			results.Replacement,
			&apis.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"})
		require.Equal(t,
			"BlahPod is a deprecated Pod.", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "networking.k8s.io", "v1", "NetworkPolicy")
		require.NoError(t, err)
		require.Equal(t, "", results.DeletedVersion)
		require.Equal(t, "", results.DeprecationVersion)
		require.Equal(t, "", results.IntroducedVersion)
		require.Equal(t,
			"", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "extensions", "v2", "Unknown")
		require.NoError(t, err)
		require.Equal(t, "", results.DeletedVersion)
		require.Equal(t, "", results.DeprecationVersion)
		require.Equal(t, "", results.IntroducedVersion)
		require.Equal(t,
			"", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "extensions", "v2", "DaemonSet")
		require.NoError(t, err)
		require.Equal(t, "", results.DeletedVersion)
		require.Equal(t, "", results.DeprecationVersion)
		require.Equal(t, "", results.IntroducedVersion)
		require.Equal(t,
			"", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "extensions", "v2", "DaemonSet")
		require.NoError(t, err)
		require.Equal(t, "", results.DeletedVersion)
		require.Equal(t, "", results.DeprecationVersion)
		require.Equal(t, "", results.IntroducedVersion)
		require.Equal(t,
			"", results.Description)

	})
}

func TestGeneratedStore_compareAndFill(t *testing.T) {

	tests := []struct {
		name             string
		apiVersion       string
		requestedVersion *semver.Version
		want             string
	}{
		{
			name:             "null semver should return the requested apiVersion",
			apiVersion:       "v1.5.6",
			want:             "v1.5.6",
			requestedVersion: nil,
		},
		{
			name:             "empty apiVersion should return empty",
			apiVersion:       "",
			want:             "",
			requestedVersion: semver.MustParse("v1.7.8"),
		},
		{
			name:             "should return the apiVersion if requested version is bigger than the deprecation",
			apiVersion:       "1.2",
			want:             "1.2",
			requestedVersion: semver.MustParse("1.3"),
		},
		{
			name:             "should return the apiVersion if requested version is bigger than the deprecation and contains v prefix",
			apiVersion:       "1.2",
			want:             "1.2",
			requestedVersion: semver.MustParse("v1.3"),
		},
		{
			name:             "should return the apiVersion if requested version is bigger than the deprecation and contains patch",
			apiVersion:       "1.2",
			want:             "1.2",
			requestedVersion: semver.MustParse("1.3.5"),
		},
		{
			name:             "should return the apiVersion if requested version is equal than the deprecation and contains patch",
			apiVersion:       "1.2",
			want:             "1.2",
			requestedVersion: semver.MustParse("1.2.5"),
		},
		{
			name:             "should return the apiVersion if requested version is equal than the deprecation and does not contains patch",
			apiVersion:       "1.2",
			want:             "1.2",
			requestedVersion: semver.MustParse("1.2"),
		},
		{
			name:             "should not return the apiVersion if requested version is less than the deprecation",
			apiVersion:       "1.2",
			want:             "",
			requestedVersion: semver.MustParse("1.1"),
		},
		{
			name:             "should return invalid if requested apiVersion is invalid",
			apiVersion:       "xxxxx",
			want:             "invalid",
			requestedVersion: semver.MustParse("1.1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GeneratedStore{
				requestedVersion: tt.requestedVersion,
			}
			if got := s.compareAndFill(tt.apiVersion); got != tt.want {
				t.Errorf("GeneratedStore.compareAndFill() = %v, want %v", got, tt.want)
			}
		})
	}
}
