package kubepug

import (
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type crdStruct map[string]struct{}

const crdGroup = "apiextensions.k8s.io"

func (crds crdStruct) populateCRDGroups(dynClient dynamic.Interface, version string) {
	crdgvr := schema.GroupVersionResource{
		Group:    crdGroup,
		Version:  version,
		Resource: "customresourcedefinitions",
	}

	crdList, err := dynClient.Resource(crdgvr).List(metav1.ListOptions{})
	if apierrors.IsNotFound(err) {
		return
	}
	if err != nil {
		panic(err)
	}

	// We'll create an empty map[crd] because that's easier than keep interating into an array/slice to find a value
	var empty struct{}

	for _, d := range crdList.Items {
		group, found, err := unstructured.NestedString(d.Object, "spec", "group")
		// No group fields found, move on!
		if err != nil || !found {
			continue
		}
		if _, ok := crds[group]; !ok {
			crds[group] = empty
		}
	}
}

// WalkObjects walk through Kubernetes API and verifies which Resources doesn't exists anymore in swagger.json
func (KubernetesAPIs KubernetesAPIs) WalkObjects(config *rest.Config) {

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

	var crds crdStruct = make(map[string]struct{})

	// Discovery CRDs versions to populate CRDs
	for _, resources := range resourcesList {
		if strings.Contains(resources.GroupVersion, crdGroup) {
			version := strings.Split(resources.GroupVersion, "/")[1]
			crds.populateCRDGroups(dynClient, version)
		}
	}

	for _, resourceGroupVersion := range resourcesList {

		// We dont want CRDs to be walked
		if _, ok := crds[strings.Split(resourceGroupVersion.GroupVersion, "/")[0]]; ok {
			continue
		}

		for i := range resourceGroupVersion.APIResources {
			resource := &resourceGroupVersion.APIResources[i]
			// We don't want to check subObjects (like pods/status)
			if len(strings.Split(resource.Name, "/")) != 1 {
				continue
			}

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
