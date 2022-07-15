package k8sinput

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"strings"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/rikatz/kubepug/pkg/parser"
	"github.com/rikatz/kubepug/pkg/results"
)

// ignoreStruct is an empty map, the key is the API Group to be ignored. No value exists
type ignoreStruct map[string]struct{}

const (
	crdGroup          = "apiextensions.k8s.io"
	apiRegistration   = "apiregistration.k8s.io"
	v1beta1IngressApi = "extensions/v1beta1/Ingress"
)

// This function will receive an apiExtension (CRD) and populate it into the struct to be verified later
func (ignoreStruct ignoreStruct) populateCRDGroups(dynClient dynamic.Interface, version string) {
	log.Debugf("Populating CRDs array of version %s", version)

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
		log.Fatalf("Failed to connect to K8s cluster to List CRDs: %s", err)
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
	log.Debugf("Populating APIService array of version %s", version)
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
		log.Fatalf("Failed to connect to K8s cluster to List API Services: %s", err)
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

// GetDeleted walk through Kubernetes API and verifies which Resources doesn't exists anymore in swagger.json
func GetDeleted(kubeAPIs parser.KubernetesAPIs, config *genericclioptions.ConfigFlags) (deleted []results.DeletedAPI) {
	configRest, err := config.ToRESTConfig()
	if err != nil {
		log.Fatalf("Failed to create the K8s config parameters while listing Deleted objects: %s", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(configRest)
	if err != nil {
		log.Fatalf("Failed to create the K8s Discovery client: %s", err)
	}

	log.Debug("Getting all the Server Resources")
	resourcesList, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		if apierrors.IsForbidden(err) {
			log.Fatalf("Failed to list Server Resources. Permission denied! Please check if you have the proper authorization")
		}

		log.Fatalf("Failed communicating with k8s while discovering server resources. \nError: %v", err)
	}

	dynClient, err := dynamic.NewForConfig(configRest)
	if err != nil {
		log.Fatalf("Failed to create dynamic client. \nError: %v", err)
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

	log.Debugf("Walking through %d resource types", len(resourcesList))
	for _, resourceGroupVersion := range resourcesList {
		// We don't want CRDs or APIExtensions to be walked
		if _, ok := ignoreObjects[strings.Split(resourceGroupVersion.GroupVersion, "/")[0]]; ok {
			continue
		}

		for i := range resourceGroupVersion.APIResources {
			resource := &resourceGroupVersion.APIResources[i] // We don't want to check subObjects (like pods/status)
			if len(strings.Split(resource.Name, "/")) != 1 {
				continue
			}

			keyAPI := fmt.Sprintf("%s/%s", resourceGroupVersion.GroupVersion, resource.Kind)
			if _, ok := kubeAPIs[keyAPI]; !ok {

				gvr, list := getResources(dynClient, resourceGroupVersion.GroupVersion, resource.Name, resource.Kind)

				if keyAPI == v1beta1IngressApi {
					list.Items = fixDeletedIngressList(dynClient, list.Items)
				}

				if len(list.Items) > 0 {
					log.Debugf("Found %d deleted items in %s/%s", len(list.Items), gvr.Group, resource.Kind)
					d := results.DeletedAPI{
						Deleted: true,
						Name:    resource.Name,
						Group:   gvr.Group,
						Kind:    resource.Kind,
						Version: gvr.Version,
					}

					d.Items = results.ListObjects(list.Items)
					deleted = append(deleted, d)
				}
			}
		}
	}

	return deleted
}

func getResources(dynClient dynamic.Interface, groupVersion string, resourceName string, resourceKind string) (schema.GroupVersionResource, *unstructured.UnstructuredList) {

	gv, err := schema.ParseGroupVersion(groupVersion)
	if err != nil {
		log.Fatalf("Failed to Parse GroupVersion of Resource: %s", err)
	}

	gvr := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resourceName}
	list, err := dynClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if apierrors.IsNotFound(err) || apierrors.IsMethodNotSupported(err) {
		return gvr, list
	}

	if apierrors.IsForbidden(err) {
		log.Fatalf("Failed to list Server Resources of type %s/%s/%s. Permission denied! Please check if you have the proper authorization", gv.Group, gv.Version, resourceKind)
	}

	if err != nil {
		log.Fatalf("Failed to List objects of type %s/%s/%s. \nError: %v", gv.Group, gv.Version, resourceKind, err)
	}

	return gvr, list
}

// Removes the false positives deleted ingresses:
// The client library is returning the same results for "extensions/v1beta1/Ingress" and for "networking.k8s.io/v1/Ingress".
func fixDeletedIngressList(dynClient dynamic.Interface, v1beta1Ingresses []unstructured.Unstructured) []unstructured.Unstructured {

	_, stableIngresses := getResources(dynClient, "networking.k8s.io/v1", "ingresses", "Ingress")
	stableIngressesMap := make(map[string]bool)

	for _, item := range stableIngresses.Items {
		uid := spew.Sprint(item.Object["metadata"].(map[string]interface{})["uid"])
		stableIngressesMap[uid] = true
	}

	deletedItems := []unstructured.Unstructured{}
	for _, item := range v1beta1Ingresses {
		uid := spew.Sprint(item.Object["metadata"].(map[string]interface{})["uid"])
		// Only adds to the deleted list if not found in the stable ingresses list
		if !stableIngressesMap[uid] {
			deletedItems = append(deletedItems, item)
		}
	}

	return deletedItems
}
