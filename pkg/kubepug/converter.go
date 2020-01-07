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

func getDeprecatedValues(value map[string]interface{}) (DeprecatedAPI, bool) {
	var valid bool
	var description, group, version, kind string
	for k, v := range value {
		if valString, ok := v.(string); k == "description" && ok {
			if strings.Contains(strings.ToLower(valString), "deprecated") {
				description = valString
			}
		}
		// Just set something deprecated if it has x-kubernetes-group-version-kind also
		if k == "x-kubernetes-group-version-kind" {
			valid = true
			// GroupVersionKind is an array of one value only
			gvkMap := v.([]interface{})[0]
			group, version, kind = getGroupVersionKind(gvkMap.(map[string]interface{}))
		}
	}
	if description != "" && valid {
		return DeprecatedAPI{
			description: description,
			group:       group,
			kind:        kind,
			version:     version,
		}, true
	}

	return DeprecatedAPI{}, false
}

// DiscoverResourceName provides a Resource Name based in its Group, Version and Kind
func DiscoverResourceName(client *discovery.DiscoveryClient, group, version, kind string) string {
	gv := fmt.Sprintf("%s/%s", group, version)
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

// DeprecatedAPIs receives a file location and Kubernetes Config 
// and returns a map of deprecated APIs to the main program
func DeprecatedAPIs(config *rest.Config, swaggerfile string) (deprecatedApis map[string]DeprecatedAPI) {

	// Open our jsonFile
	jsonFile, err := os.Open(swaggerfile)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)
	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above

	json.Unmarshal(byteValue, &definitionsMap)

	deprecatedApis = make(map[string]DeprecatedAPI)
	definitions := definitionsMap["definitions"].(map[string]interface{})

	for key, value := range definitions {
		val := value.(map[string]interface{})
		if deprecatedAPI, deprecated := getDeprecatedValues(val); deprecated {
			disco, err := discovery.NewDiscoveryClientForConfig(config)
			if err != nil {
				panic(err)
			}
			resource := DiscoverResourceName(disco, deprecatedAPI.group, deprecatedAPI.version, deprecatedAPI.kind)
			deprecatedApis[key] = DeprecatedAPI{
				group:       deprecatedAPI.group,
				kind:        deprecatedAPI.kind,
				description: deprecatedAPI.description,
				name:        resource,
				version:     deprecatedAPI.version,
			}
		}
	}

	return deprecatedApis
}
