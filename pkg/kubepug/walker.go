package kubepug

import (
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func WalkObjects(config *rest.Config, KubernetesAPIs map[string]KubeAPI) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err)
	}

	resourcesList, err := discoveryClient.ServerResources()
	if err != nil {
		panic(err)
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	for _, resourceGroupVersion := range resourcesList {
		//fmt.Printf("%s\n\n", v.GroupVersion)
		for _, resource := range resourceGroupVersion.APIResources {

			// If this is a subObject (like pods/status) we will disconsider this
			if len(strings.Split(resource.Name, "/")) == 1 {

				keyAPI := fmt.Sprintf("%s/%s", resourceGroupVersion.GroupVersion, resource.Name)
				if _, ok := KubernetesAPIs[keyAPI]; !ok {
					gv, err := schema.ParseGroupVersion(resourceGroupVersion.GroupVersion)
					if err != nil {
						panic(err)
					}
					gvr := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource.Name}
					list, err := dynClient.Resource(gvr).List(metav1.ListOptions{})
					if apierrors.IsNotFound(err) {
						continue
					}
					if err != nil {
						panic(err)
					}
					if len(list.Items) > 0 {
						fmt.Printf("%s found in %s/%s\n", resourceColor(resource.Kind), gvColor(gv.Group), gvColor(gv.Version))
						fmt.Printf("\t ├─ %s\n", errorColor("API REMOVED FROM THE CURRENT VERSION AND SHOULD BE MIGRATED IMMEDIATELLY!!"))
						listObjects(list.Items)
						fmt.Printf("\n")

					}
				}
			}
		}
	}
}
