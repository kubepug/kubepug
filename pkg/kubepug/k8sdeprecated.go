package kubepug

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func listObjects(items []unstructured.Unstructured) {

	for _, d := range items {
		name := d.GetName()
		namespace := d.GetNamespace()
		if namespace != "" {
			fmt.Printf("\t\t-> %s %s %s %s\n", namespaceColor("Object:"), name, namespaceColor("namespace:"), namespace)
		} else {
			fmt.Printf("\t\t-> %s: %s \n", globalColor("GLOBAL"), name)
		}
	}

}

// ListDeprecated receives the Map of Deprecated API and List the existent Deprecated Objects in the Cluster
func (KubeAPIs KubernetesAPIs) ListDeprecated(config *rest.Config, showDescription bool) {

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	for _, dpa := range KubeAPIs {
		// We only want deprecated APIs :)
		if !dpa.deprecated {
			continue
		}
		gvr := schema.GroupVersionResource{Group: dpa.group, Version: dpa.version, Resource: dpa.name}
		list, err := client.Resource(gvr).List(metav1.ListOptions{})
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			panic(err)
		}
		if len(list.Items) > 0 {
			fmt.Printf("%s found in %s/%s\n", resourceColor(dpa.kind), gvColor(dpa.group), gvColor(dpa.version))
			if showDescription {
				fmt.Printf("\t ├─ %s\n", dpa.description)
			}

			listObjects(list.Items)
			fmt.Printf("\n")
		}
	}

}
