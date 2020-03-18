package kubepug

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func listObjects(items []unstructured.Unstructured) []DeprecatedItem {
	deprecatedItems := []DeprecatedItem{}
	for _, d := range items {
		name := d.GetName()
		namespace := d.GetNamespace()
		if namespace != "" {
			deprecatedItems = append(deprecatedItems, DeprecatedItem{Kind: "OBJECT", Name: name, Namespace: namespace})
		} else {
			deprecatedItems = append(deprecatedItems, DeprecatedItem{Kind: "GLOBAL", Name: name})
		}
	}
	return deprecatedItems
}

// ListDeprecated receives the Map of Deprecated API and List the existent Deprecated Objects in the Cluster
func (KubeAPIs KubernetesAPIs) ListDeprecated(config *rest.Config, showDescription bool) []DeprecatedAPI {

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deprecated := []DeprecatedAPI{}

	for _, dpa := range KubeAPIs {
		// We only want deprecated APIs :)
		if !dpa.deprecated {
			continue
		}
		gvr := schema.GroupVersionResource{Group: dpa.group, Version: dpa.version, Resource: dpa.name}
		list, err := client.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			panic(err)
		}
		if len(list.Items) > 0 {
			api := DeprecatedAPI{
				Kind:       dpa.kind,
				Deprecated: dpa.deprecated,
				Group:      dpa.group,
				Name:       dpa.name,
				Version:    dpa.version,
			}
			if showDescription {
				api.Description = dpa.description
			}
			api.Items = listObjects(list.Items)
			deprecated = append(deprecated, api)
		}
	}

	return deprecated
}
