package kubepug

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
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
		panic(err)
	}

	for k, v := range value {
		if valString, ok := v.(string); k == "description" && ok {
			if strings.Contains(strings.ToLower(valString), "deprecated") {
				deprecated = true
			}
			description = valString
		}

		// Just set something as a valid API if it has x-kubernetes-group-version-kind also
		if k == "x-kubernetes-group-version-kind" {
			valid = true
			// GroupVersionKind is an array of one value only
			gvkMap := v.([]interface{})[0]
			group, version, kind = getGroupVersionKind(gvkMap.(map[string]interface{}))

			resourceName = DiscoverResourceName(disco, group, version, kind)

			if resourceName = DiscoverResourceName(disco, group, version, kind); resourceName == "" {
				// If no ResourceName is found in the API Server this Resource does not exists and should
				// be ignored
				valid = false
			}
		}
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
func PopulateKubeAPIMap(config *rest.Config, swaggerfile string) (KubeAPIs map[string]KubeAPI) {

	KubeAPIs = make(map[string]KubeAPI)

	// Open our jsonFile
	jsonFile, err := os.Open(swaggerfile)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	jsonFile.Close()

	json.Unmarshal(byteValue, &definitionsMap)

	definitions := definitionsMap["definitions"].(map[string]interface{})

	for _, value := range definitions {
		val := value.(map[string]interface{})
		if kubeapivalue, valid := getKubeAPIValues(val, config); valid {
			var name string
			if kubeapivalue.group != "" {
				name = fmt.Sprintf("%s/%s/%s", kubeapivalue.group, kubeapivalue.version, kubeapivalue.name)
			} else {
				name = fmt.Sprintf("%s/%s", kubeapivalue.version, kubeapivalue.name)
			}
			//			fmt.Printf("%v\n", kubeapivalue)
			KubeAPIs[name] = kubeapivalue
		}
	}
	return KubeAPIs
}

// DiscoverResourceName provides a Resource Name based in its Group, Version and Kind
func DiscoverResourceName(client *discovery.DiscoveryClient, group, version, kind string) string {
	var gv string
	if group != "" {
		gv = fmt.Sprintf("%s/%s", group, version)
	} else {
		gv = fmt.Sprintf("%s", version)
	}
	resources, err := client.ServerResourcesForGroupVersion(gv)
	if err != nil {
		return ""
	}
	for _, apires := range resources.APIResources {
		if apires.Kind == kind {
			return apires.Name
		}
	}
	return ""
}
