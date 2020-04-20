package kubepug

import (
	"context"
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

type ignoreStruct map[string]struct{}

const crdGroup = "apiextensions.k8s.io"
const apiRegistration = "apiregistration.k8s.io"

// This function will receive an apiExtension (CRD) and populate it into the struct to be verified later
func (ignoreStruct ignoreStruct) populateCRDGroups(dynClient dynamic.Interface, version string) {
	crdgvr := schema.GroupVersionResource{
		Group:    crdGroup,
		Version:  version,
		Resource: "customresourcedefinitions",
	}

	crdList, err := dynClient.Resource(crdgvr).List(context.TODO(), metav1.ListOptions{})
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
		if _, ok := ignoreStruct[group]; !ok {
			ignoreStruct[group] = empty
		}
	}
}

// This function will receive an apiRegistration (APIService) and populate it into the struct
// to be verified later. It will consider only if the field "service" is not null
// representing an external Service
func (ignoreStruct ignoreStruct) populateAPIService(dynClient dynamic.Interface, version string) {
	apisvcgvr := schema.GroupVersionResource{
		Group:    apiRegistration,
		Version:  version,
		Resource: "apiservices",
	}

	apisvcList, err := dynClient.Resource(apisvcgvr).List(context.TODO(), metav1.ListOptions{})
	if apierrors.IsNotFound(err) {
		return
	}
	if err != nil {
		panic(err)
	}

	// We'll create an empty map[crd] because that's easier than keep interating into an array/slice to find a value
	var empty struct{}

	for _, d := range apisvcList.Items {
		_, foundSvc, errSvc := unstructured.NestedString(d.Object, "spec", "service", "name")
		group, foundGrp, errGrp := unstructured.NestedString(d.Object, "spec", "group")
		// No services fields or group field found, move on!
		if errSvc != nil || !foundSvc || errGrp != nil || !foundGrp {
			continue
		}

		if _, ok := ignoreStruct[group]; !ok {
			ignoreStruct[group] = empty
		}
	}
}

// WalkObjects walk through Kubernetes API and verifies which Resources doesn't exists anymore in swagger.json
func (KubernetesAPIs KubernetesAPIs) WalkObjects(config *rest.Config) []DeletedAPI {

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

	var ignoreObjects ignoreStruct = make(map[string]struct{})

	// Discovery CRDs versions to populate CRDs
	for _, resources := range resourcesList {
		if strings.Contains(resources.GroupVersion, crdGroup) {
			version := strings.Split(resources.GroupVersion, "/")[1]
			ignoreObjects.populateCRDGroups(dynClient, version)
		}
		if strings.Contains(resources.GroupVersion, apiRegistration) {
			version := strings.Split(resources.GroupVersion, "/")[1]
			ignoreObjects.populateAPIService(dynClient, version)
		}
	}

	deleted := []DeletedAPI{}
	for _, resourceGroupVersion := range resourcesList {

		// We dont want CRDs to be walked
		if _, ok := ignoreObjects[strings.Split(resourceGroupVersion.GroupVersion, "/")[0]]; ok {
			continue
		}

		for i := range resourceGroupVersion.APIResources {
			resource := &resourceGroupVersion.APIResources[i]
			// We don't want to check subObjects (like pods/status)
			if len(strings.Split(resource.Name, "/")) != 1 {
				continue
			}
			keyAPI := fmt.Sprintf("%s/%s\n", resourceGroupVersion.GroupVersion, resource.Name)
			if _, ok := KubernetesAPIs[keyAPI]; !ok {
				gv, err := schema.ParseGroupVersion(resourceGroupVersion.GroupVersion)
				if err != nil {
					panic(err)
				}
				gvr := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource.Name}
				list, err := dynClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
				if apierrors.IsNotFound(err) {
					continue
				}
				if apierrors.IsMethodNotSupported(err) {
					continue
				}
				if err != nil {
					panic(err)
				}
				if len(list.Items) > 0 {
					d := DeletedAPI{
						Deleted: true,
						Name:    resource.Name,
						Group:   gvr.GroupResource().String(),
						Kind:    resource.Kind,
						Version: gv.Version,
					}
					d.Items = listObjects(list.Items)
					deleted = append(deleted, d)
				}
			}
		}
	}
	return deleted
}
