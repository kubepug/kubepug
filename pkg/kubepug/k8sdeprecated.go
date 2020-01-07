package kubepug

import (
	"fmt"

	"github.com/fatih/color"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// ListDeprecated receives the Map of Deprecated API and List the existent Deprecated Objects in the Cluster
func ListDeprecated(config *rest.Config, deprecatedApis map[string]DeprecatedAPI, showDescription bool) {

	gvColor := color.New(color.FgHiBlack).Add(color.Bold).SprintFunc()
	resourceColor := color.New(color.FgRed).Add(color.Bold).SprintFunc()
	globalColor := color.New(color.FgBlue).Add(color.Bold).SprintFunc()
	namespaceColor := color.New(color.FgCyan).Add(color.Bold).SprintFunc()

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	for _, dpa := range deprecatedApis {
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
			for _, d := range list.Items {
				name := d.GetName()
				namespace := d.GetNamespace()
				if namespace != "" {
					fmt.Printf("\t\t-> %s %s %s %s\n", namespaceColor("Object:"), name, namespaceColor("namespace:"), namespace)
				} else {
					fmt.Printf("\t\t-> %s: %s \n", globalColor("GLOBAL"), name)
				}
			}
			fmt.Printf("\n")
		}
	}

}
