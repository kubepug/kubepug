package kubepug

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

var definitionsMap map[string]interface{}

func getGroupVersionKind(value map[string]interface{}) (group, version, kind string) {
	for k, v := range value {
		switch k {
		case "group":
			group = v.(string)
		case "version":
			version = v.(string)
		case "kind":
			kind = v.(string)
		}
	}
	return group, version, kind
}

func getKubeAPIValues(value map[string]interface{}, config *rest.Config) (KubeAPI, bool) {
	var valid, deprecated bool
	var description, group, version, kind, resourceName string

	disco, err := discovery.NewDiscoveryClientForConfig(config)

	if err != nil {
		log.Fatalf("Failed to create the K8s Discovery client")
	}

	gvk, valid, err := unstructured.NestedSlice(value, "x-kubernetes-group-version-kind")

	if !valid || err != nil {
		return KubeAPI{}, false
	}

	gvkMap := gvk[0]
	group, version, kind = getGroupVersionKind(gvkMap.(map[string]interface{}))

	if resourceName = DiscoverResourceName(disco, group, version, kind); resourceName == "" {
		// If no ResourceName is found in the API Server this Resource does not exists and should
		// be ignored
		log.Debugf("Marking the resource as invalid because it doesn't exists in the APIServer")
		valid = false
	}

	description, found, err := unstructured.NestedString(value, "description")

	if !found || err != nil {
		log.Debugf("Marking the resource as invalid because it doesn't contain a description")
		return KubeAPI{}, false
	}

	if strings.Contains(strings.ToLower(description), "deprecated") {
		log.Debugf("API Definition contains the word DEPRECATED in its description")
		deprecated = true
	}

	if valid {
		return KubeAPI{
			description: description,
			group:       group,
			kind:        kind,
			version:     version,
			name:        resourceName,
			deprecated:  deprecated,
		}, true
	}

	return KubeAPI{}, false
}

// PopulateKubeAPIMap Converts an API Definition into a map of KubeAPIs["group/version/name"]
func (KubeAPIs KubernetesAPIs) PopulateKubeAPIMap(config *rest.Config, swaggerfile string) (err error) {
	// Open our jsonFile
	log.Debugf("Opening the swagger file for reading: %s", swaggerfile)
	jsonFile, err := os.Open(swaggerfile)
	// if we os.Open returns an error then handle it
	if err != nil {
		return err
	}
	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = jsonFile.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(byteValue, &definitionsMap)
	if err != nil {
		log.Warning("Error parsing the JSON, file might me invalid")
		return err
	}
	definitions := definitionsMap["definitions"].(map[string]interface{})

	log.Debugf("Iteracting through %d definitions", len(definitions))
	for k, value := range definitions {
		val := value.(map[string]interface{})
		log.Debugf("Getting API values from %s", k)
		if kubeapivalue, valid := getKubeAPIValues(val, config); valid {
			log.Debugf("Valid API object found for %s", k)
			var name string
			if kubeapivalue.group != "" {
				name = fmt.Sprintf("%s/%s/%s", kubeapivalue.group, kubeapivalue.version, kubeapivalue.name)
			} else {
				name = fmt.Sprintf("%s/%s", kubeapivalue.version, kubeapivalue.name)
			}
			log.Debugf("Adding %s to map. Deprecated: %t", name, kubeapivalue.deprecated)
			KubeAPIs[name] = kubeapivalue
		}
	}
	return nil
}

// DiscoverResourceName provides a Resource Name based in its Group, Version and Kind
// This is necessary when you're listing all the existing resources in the cluster
// as you've to pass group/version/name (and not group/version/kind) to client.resource.List
func DiscoverResourceName(client *discovery.DiscoveryClient, group, version, kind string) string {
	var gv string
	if group != "" {
		gv = fmt.Sprintf("%s/%s", group, version)
	} else {
		gv = version
	}
	resources, err := client.ServerResourcesForGroupVersion(gv)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ""
		}
		if apierrors.IsForbidden(err) {
			log.Fatalf("Failed to list object %s attribute. Permission denied! Please check if you have the proper authorization", gv)
		}
		log.Fatalf("Failed communicating with k8s while discovering the object name for %s. \nError: %v", gv, err)
	}
	for i := range resources.APIResources {
		apires := &resources.APIResources[i]
		if apires.Kind == kind {
			return apires.Name
		}
	}
	return ""
}
