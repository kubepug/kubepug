package kubepug

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func getKubeAPIValues(value map[string]interface{}, disco *discovery.DiscoveryClient) (KubeAPI, bool) {
	var valid, deprecated bool
	var description, group, version, kind, resourceName string

	gvk, valid, err := unstructured.NestedSlice(value, "x-kubernetes-group-version-kind")

	if !valid || err != nil {
		return KubeAPI{}, false
	}

	gvkMap := gvk[0]
	group, version, kind = getGroupVersionKind(gvkMap.(map[string]interface{}))

	if resourceName = DiscoverResourceName(disco, group, version, kind); resourceName == "" {
		// If no ResourceName is found in the API Server this Resource does not exists and should
		// be ignored
		valid = false
	}

	description, found, err := unstructured.NestedString(value, "description")

	if !found || err != nil {
		return KubeAPI{}, false
	}

	if strings.Contains(strings.ToLower(description), "deprecated") {
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
	fmt.Println("6.1")
	// Open our jsonFile
	jsonFile, err := os.Open(swaggerfile)
	// if we os.Open returns an error then handle it
	if err != nil {
		return err
	}
	fmt.Println("6.2")
	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = jsonFile.Close()
	if err != nil {
		return err
	}
	fmt.Println("6.3")
	err = json.Unmarshal(byteValue, &definitionsMap)
	if err != nil {
		fmt.Println("Error parsing the JSON, file might me invalid")
		return err
	}
	fmt.Println("6.4")
	definitions := definitionsMap["definitions"].(map[string]interface{})

	disco, err := discovery.NewDiscoveryClientForConfig(config)

	if err != nil {
		panic(err)
	}

	for i, value := range definitions {
		val := value.(map[string]interface{})
		fmt.Printf("6.4.%s\n", i)
		if kubeapivalue, valid := getKubeAPIValues(val, disco); valid {
			var name string
			if kubeapivalue.group != "" {
				name = fmt.Sprintf("%s/%s/%s", kubeapivalue.group, kubeapivalue.version, kubeapivalue.name)
			} else {
				name = fmt.Sprintf("%s/%s", kubeapivalue.version, kubeapivalue.name)
			}
			KubeAPIs[name] = kubeapivalue
		}
	}
	fmt.Println("6.5")
	return nil
}

// DiscoverResourceName provides a Resource Name based in its Group, Version and Kind
func DiscoverResourceName(client *discovery.DiscoveryClient, group, version, kind string) string {
	var gv string
	if group != "" {
		gv = fmt.Sprintf("%s/%s", group, version)
	} else {
		gv = version
	}
	resources, err := client.ServerResourcesForGroupVersion(gv)
	if err != nil {
		return ""
	}
	for i := range resources.APIResources {
		apires := &resources.APIResources[i]
		if apires.Kind == kind {
			return apires.Name
		}
	}
	return ""
}
