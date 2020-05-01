package kubepug

import (
	"context"

	log "github.com/sirupsen/logrus"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func listObjects(items []unstructured.Unstructured) (deprecatedItems []DeprecatedItem) {
	for _, d := range items {
		name := d.GetName()
		namespace := d.GetNamespace()
		if namespace != "" {
			deprecatedItems = append(deprecatedItems, DeprecatedItem{Scope: "OBJECT", ObjectName: name, Namespace: namespace})
		} else {
			deprecatedItems = append(deprecatedItems, DeprecatedItem{Scope: "GLOBAL", ObjectName: name})
		}
	}
	return deprecatedItems
}

// ListDeprecated receives the Map of Deprecated API and List the existent Deprecated Objects in the Cluster
func (KubeAPIs KubernetesAPIs) ListDeprecated(config *rest.Config, showDescription bool) (deprecated []DeprecatedAPI) {

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create the K8s client while listing Deprecated objects")
	}

	for _, dpa := range KubeAPIs {
		// We only want deprecated APIs :)
		if !dpa.deprecated {
			continue
		}
		log.Debugf("Listing objects for %s/%s/%s", dpa.group, dpa.version, dpa.name)
		gvr := schema.GroupVersionResource{Group: dpa.group, Version: dpa.version, Resource: dpa.name}
		list, err := client.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
		if apierrors.IsNotFound(err) {
			continue
		}
		if apierrors.IsForbidden(err) {
			log.Fatalf("Failed to list objects in the cluster. Permission denied! Please check if you have the proper authorization")
		}
		if err != nil {
			log.Fatalf("Failed communicating with k8s while listing objects. \nError: %v", err)
		}
		if len(list.Items) > 0 {
			log.Infof("Found %d deprecated objects of type %s/%s/%s", len(list.Items), dpa.group, dpa.version, dpa.name)
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
