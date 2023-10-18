package generatedstore

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/Masterminds/semver/v3"
	apis "github.com/kubepug/kubepug/pkg/apis/v1alpha1"
	"github.com/kubepug/kubepug/pkg/utils"

	generatedapi "github.com/kubepug/kubepug/generator/deprecations"

	json "github.com/goccy/go-json"
)

type StoreConfig struct {
	// MinVerison defines the Kubernetes MinVersion that should be compared with this API
	MinVersion string
	// Path defines the path of the generated File
	Path string
	// internalPath defines the real path to be used on file location
	// this can be a temporary location in case of file being downloaded
	internalPath string
}

type GeneratedStore struct {
	db               apis.APIGroups
	config           StoreConfig
	requestedVersion *semver.Version
}

func NewGeneratedStore(config StoreConfig) (*GeneratedStore, error) {
	if config.Path == "" {
		return nil, fmt.Errorf("generated json location cannot be null")
	}

	// Set the internal location initially as the same value, if we are dealing with
	// a local file
	config.internalPath = config.Path

	urlLocation, err := url.Parse(config.Path)
	if err == nil && (urlLocation.Scheme == "http" || urlLocation.Scheme == "https") {
		config.internalPath, err = utils.DownloadGeneratedJSON(config.Path)
		if err != nil {
			return nil, err
		}
	}

	file, err := os.ReadFile(config.internalPath)
	if err != nil {
		return nil, err
	}

	return NewGeneratedStoreFromBytes(file, config)
}

// NewGeneratedStoreFromBytes allows setting a reader as a database. It should contain
// a valid Kubernetes generated data definition
func NewGeneratedStoreFromBytes(data []byte, config StoreConfig) (*GeneratedStore, error) {
	var parsedVersion *semver.Version
	var err error

	if config.MinVersion != "" && config.MinVersion != "master" {
		parsedVersion, err = semver.NewVersion(config.MinVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to parse min version: %s", err)
		}
	}

	db, err := newInternalDatabase(data)
	if err != nil {
		return nil, err
	}
	return &GeneratedStore{
		db:               db,
		config:           config,
		requestedVersion: parsedVersion,
	}, nil
}

func newInternalDatabase(data []byte) (apis.APIGroups, error) {
	defs := []generatedapi.APIDeprecation{}
	err := json.Unmarshal(data, &defs)
	if err != nil {
		return nil, fmt.Errorf("error parsing the JSON, file might be invalid: %v", err)
	}

	apigroup := make(apis.APIGroups)
	for k := range defs {
		group := defs[k].Group
		if group == "" {
			group = apis.CoreAPI // Special type for Core APIs
		}
		if _, ok := apigroup[group]; !ok {
			apigroup[group] = make(apis.APIKinds)
		}
		if _, ok := apigroup[group][defs[k].Kind]; !ok {
			apigroup[group][defs[k].Kind] = make(apis.APIVersion)
		}

		status := apis.APIVersionStatus{
			Description:        defs[k].Description,
			IntroducedVersion:  generateVersion(defs[k].IntroducedVersion.VersionMajor, defs[k].IntroducedVersion.VersionMinor),
			DeprecationVersion: generateVersion(defs[k].DeprecatedVersion.VersionMajor, defs[k].DeprecatedVersion.VersionMinor),
			DeletedVersion:     generateVersion(defs[k].RemovedVersion.VersionMajor, defs[k].RemovedVersion.VersionMinor),
		}
		if defs[k].Replacement.Version != "" && defs[k].Replacement.Kind != "" {
			status.Replacement = &apis.GroupVersionKind{
				Group:   defs[k].Replacement.Group,
				Version: defs[k].Replacement.Version,
				Kind:    defs[k].Replacement.Kind,
			}
		}

		apigroup[group][defs[k].Kind][defs[k].Version] = status
	}
	return apigroup, nil
}

func generateVersion(major, minor int) string {
	if major == 0 || minor == 0 {
		return ""
	}
	return fmt.Sprintf("%d.%d", major, minor)
}

func (s *GeneratedStore) GetAPIDefinition(_ context.Context, group, version, kind string) (result apis.APIVersionStatus, err error) {
	result = apis.APIVersionStatus{}

	if group == "" {
		group = apis.CoreAPI
	}

	// On this store, non found APIs means they are not Kubernetes APIs so we can skip it
	apigroup, ok := s.db[group]
	if !ok {
		return result, nil
	}

	apikind, ok := apigroup[kind]
	if !ok {
		return result, nil
	}

	apiversion, ok := apikind[version]
	if !ok {
		return result, nil
	}

	result.DeletedVersion = s.compareAndFill(apiversion.DeletedVersion)
	result.DeprecationVersion = s.compareAndFill(apiversion.DeprecationVersion)
	result.IntroducedVersion = s.compareAndFill(apiversion.IntroducedVersion)
	result.Description = apiversion.Description
	result.Replacement = apiversion.Replacement

	return result, nil
}

// compareAndFillVersion gets the requested version and compares with apiVersion
// If the requestedVersion is less than the detected version, it should be empty so the
// API won't be tagged (as deprecated or deleted)
func (s *GeneratedStore) compareAndFill(apiVersion string) string {
	// If empty or requestedVersion is null, it means we can use whatever is on apiVersion
	if s.requestedVersion == nil || apiVersion == "" {
		return apiVersion
	}

	// We don't expect errors to happen here as this apiVersion is generated from k8s API
	// but if it happens, we can just return a version that will always be deprecated.
	// Better a false positive and an issue report than hidding deprecated APIs.
	apiVersionVer, err := semver.NewVersion(apiVersion)
	if err != nil {
		return "invalid"
	}

	// If the requestedVersion is less than when API was marked as deprecated,
	// we should skip it
	if s.requestedVersion.LessThan(apiVersionVer) {
		return ""
	}

	return apiVersion
}
