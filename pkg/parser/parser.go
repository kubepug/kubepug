package parser

import (
	"fmt"
	"os"
	"strings"

	json "github.com/goccy/go-json"

	log "github.com/sirupsen/logrus"
)

// PopulateKubeAPIMap Converts a Swagger Definition into a map of KubeAPIs["group/version/kind"]
func (kubeAPIs KubernetesAPIs) PopulateKubeAPIMap(swaggerfile string) (err error) {
	// Open our jsonFile
	log.Debugf("Opening the swagger file for reading: %s", swaggerfile)
	byteValue, err := os.ReadFile(swaggerfile)
	if err != nil {
		return err
	}

	defs := &definitionsJson{}
	err = json.Unmarshal(byteValue, defs)
	if err != nil {
		return fmt.Errorf("error parsing the JSON, file might be invalid: %v", err)
	}

	log.Debugf("Iteracting through %d definitions", len(defs.Definitions))
	for k, value := range defs.Definitions {
		log.Debugf("Getting API values from %s", k)

		if value.Description == "" || len(value.GroupVersionKind) == 0 {
			continue
		}

		// We need also non deprecated APIs for the APIWalk/deleted APIs
		var deprecated bool
		if strings.Contains(strings.ToLower(value.Description), "deprecated") {
			log.Debugf("API Definition does not contains the word DEPRECATED in its description, skipping")
			deprecated = true
		}

		var name string
		for _, gvk := range value.GroupVersionKind {
			if gvk.Group != "" {
				name = fmt.Sprintf("%s/%s/%s", gvk.Group, gvk.Version, gvk.Kind)
			} else {
				name = fmt.Sprintf("%s/%s", gvk.Version, gvk.Kind)
			}

			log.Debugf("Adding %s to map.", name)
			kubeAPIs[name] = KubeAPI{
				Description: value.Description,
				Group:       gvk.Group,
				Kind:        gvk.Kind,
				Version:     gvk.Version,
				Deprecated:  deprecated,
			}
		}
	}
	return nil
}
