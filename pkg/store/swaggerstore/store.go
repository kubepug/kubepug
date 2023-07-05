package swaggerstore

import (
	"context"
	"fmt"
	"os"
	"strings"

	pugerrors "github.com/rikatz/kubepug/pkg/errors"

	json "github.com/goccy/go-json"

	log "github.com/sirupsen/logrus"
)

type SwaggerStore struct {
	db APIGroups
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

func newInternalDatabase(data []byte) (APIGroups, error) {
	defs := &definitionsJSON{}
	err := json.Unmarshal(data, defs)
	if err != nil {
		return nil, fmt.Errorf("error parsing the JSON, file might be invalid: %v", err)
	}

	apigroup := make(APIGroups)
	for _, definition := range defs.Definitions {
		var deprecated bool
		if strings.Contains(strings.ToLower(definition.Description), "deprecated") {
			log.Debugf("API Definition does not contains the word DEPRECATED in its description, skipping")
			deprecated = true
		}
		for _, groups := range definition.GroupVersionKind {
			if groups.Group == "" {
				groups.Group = CoreAPI // Special type for Core APIs
			}
			if _, ok := apigroup[groups.Group]; !ok {
				apigroup[groups.Group] = make(APIKinds)
			}
			if _, ok := apigroup[groups.Group][groups.Kind]; !ok {
				apigroup[groups.Group][groups.Kind] = make(APIVersion)
			}

			apigroup[groups.Group][groups.Kind][groups.Version] = APIVersionStatus{
				Description: definition.Description,
				Deprecated:  deprecated,
			}
		}
	}
	return apigroup, nil
}

func (s *SwaggerStore) GetAPIDefinition(_ context.Context, group, version, kind string) (description string, deprecated bool, err error) {
	if group == "" {
		group = CoreAPI
	}

	apigroup, ok := s.db[group]
	if !ok {
		return "", false, pugerrors.ErrAPINotFound
	}

	apikind, ok := apigroup[kind]
	if !ok {
		return "", false, pugerrors.ErrAPINotFound
	}

	apiversion, ok := apikind[version]
	if !ok {
		return "", false, pugerrors.ErrAPINotFound
	}

	return apiversion.Description, apiversion.Deprecated, nil
}
