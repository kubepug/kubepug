package swaggerstore

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	apis "github.com/rikatz/kubepug/pkg/apis/v1alpha1"
	pugerrors "github.com/rikatz/kubepug/pkg/errors"

	json "github.com/goccy/go-json"

	log "github.com/sirupsen/logrus"
)

type SwaggerStore struct {
	db apis.APIGroups
}

func NewSwaggerStoreFromFile(swaggerfile string) (*SwaggerStore, error) {
	file, err := os.ReadFile(swaggerfile)
	if err != nil {
		return nil, err
	}

	return NewSwaggerStoreFromBytes(file)
}

// NewSwaggerStoreFromReader allows setting a reader as a database. It should contain
// a valid Kubernetes swagger definition
func NewSwaggerStoreFromBytes(data []byte) (*SwaggerStore, error) {
	db, err := newInternalDatabase(data)
	if err != nil {
		return nil, err
	}
	return &SwaggerStore{
		db: db,
	}, nil
}

func newInternalDatabase(data []byte) (apis.APIGroups, error) {
	defs := &definitionsJSON{}
	err := json.Unmarshal(data, defs)
	if err != nil {
		return nil, fmt.Errorf("error parsing the JSON, file might be invalid: %v", err)
	}

	apigroup := make(apis.APIGroups)
	for _, definition := range defs.Definitions {
		var deprecated bool
		if strings.Contains(strings.ToLower(definition.Description), "deprecated") {
			log.Debugf("API Definition does not contains the word DEPRECATED in its description, skipping")
			deprecated = true
		}
		for _, groups := range definition.GroupVersionKind {
			if groups.Group == "" {
				groups.Group = apis.CoreAPI // Special type for Core APIs
			}
			if _, ok := apigroup[groups.Group]; !ok {
				apigroup[groups.Group] = make(apis.APIKinds)
			}
			if _, ok := apigroup[groups.Group][groups.Kind]; !ok {
				apigroup[groups.Group][groups.Kind] = make(apis.APIVersion)
			}

			apigroup[groups.Group][groups.Kind][groups.Version] = apis.APIVersionStatus{
				Description:        definition.Description,
				DeprecationVersion: fmt.Sprintf("%t", deprecated),
			}
		}
	}
	return apigroup, nil
}

func (s *SwaggerStore) GetAPIDefinition(_ context.Context, group, version, kind string) (result apis.APIVersionStatus, err error) {

	result = apis.APIVersionStatus{}

	if group == "" {
		group = apis.CoreAPI
	}

	apigroup, ok := s.db[group]
	if !ok {
		return apis.APIVersionStatus{
			DeletedVersion: internalStatusVersion,
		}, pugerrors.ErrAPINotFound
	}

	apikind, ok := apigroup[kind]
	if !ok {
		return apis.APIVersionStatus{
			DeletedVersion: internalStatusVersion,
		}, pugerrors.ErrAPINotFound
	}

	apiversion, ok := apikind[version]
	if !ok {
		return apis.APIVersionStatus{
			DeletedVersion: internalStatusVersion,
		}, pugerrors.ErrAPINotFound
	}

	deprecatedAPI, err := strconv.ParseBool(apiversion.DeprecationVersion)
	if err != nil || deprecatedAPI {
		result.DeprecationVersion = internalStatusVersion
	}

	result.Description = apiversion.Description

	return result, nil
}
