package parser

import (
	"fmt"
	"os"
	"strings"

	json "github.com/goccy/go-json"
	"github.com/rikatz/kubepug/pkg/results"

	log "github.com/sirupsen/logrus"
)

func NewAPIGroupsFromSwaggerFile(swaggerfile string) (APIGroups, error) {
	defs, err := newDefinitionsFromFile(swaggerfile)
	if err != nil {
		return nil, err
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

func newDefinitionsFromFile(swaggerfile string) (*definitionsJSON, error) {
	log.Debugf("Opening the swagger file for reading: %s", swaggerfile)
	byteValue, err := os.ReadFile(swaggerfile)
	if err != nil {
		return nil, err
	}

	defs := &definitionsJSON{}
	err = json.Unmarshal(byteValue, defs)
	if err != nil {
		return nil, fmt.Errorf("error parsing the JSON, file might be invalid: %v", err)
	}
	return defs, nil
}

type ListerFunc func(group, version, resource, kind string) (results.ResultItem, error)

// CheckForItem verifies for an item inside a database and returns:
// An item (may be null case not found/not deprecated)
// A bool that indicate if the field was deleted. False indicates it was deprecated
// An error
func (db APIGroups) CheckForItem(group, version, kind, resource string, itemlister ListerFunc) (*results.ResultItem, bool, error) {
	apigroup, ok := db[group]
	if !ok {
		result, err := itemlister(group, version, resource, kind)
		if err != nil {
			return nil, false, err
		}
		if len(result.Items) == 0 {
			return nil, false, nil
		}
		return &result, true, nil
	}

	apikind, ok := apigroup[kind]
	if !ok {
		result, err := itemlister(group, version, resource, kind)
		if err != nil {
			return nil, false, err
		}
		if len(result.Items) == 0 {
			return nil, false, nil
		}
		return &result, true, nil
	}

	apiversion, ok := apikind[version]
	if !ok {
		result, err := itemlister(group, version, resource, kind)
		if err != nil {
			return nil, false, err
		}
		if len(result.Items) == 0 {
			return nil, false, nil
		}
		return &result, true, nil
	}

	if apiversion.Deprecated {
		result, err := itemlister(group, version, resource, kind)
		if err != nil {
			return nil, false, err
		}
		result.Description = apiversion.Description
		if len(result.Items) == 0 {
			return nil, false, nil
		}
		return &result, false, nil
	}

	return nil, false, nil
}
