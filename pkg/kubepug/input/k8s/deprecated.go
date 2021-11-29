package k8sinput

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/rikatz/kubepug/pkg/parser"
	"github.com/rikatz/kubepug/pkg/results"
)

// GetDeprecated receives the Map of Deprecated API and List the existent Deprecated Objects in the Cluster
func GetDeprecated(kubeAPIs parser.KubernetesAPIs, config *genericclioptions.ConfigFlags) (deprecated []results.DeprecatedAPI) {
	var resourceName string

	configRest, err := config.ToRESTConfig()
	if err != nil {
		log.Fatalf("Failed to create the K8s config parameters while listing Deprecated objects: %s", err)
	}

	client, err := dynamic.NewForConfig(configRest)
	if err != nil {
		log.Fatalf("Failed to create the K8s client while listing Deprecated objects: %s", err)
	}

	// Feed the KubeAPIs with the resourceName as this is used to the K8s Resource lister
	disco, err := discovery.NewDiscoveryClientForConfig(configRest)
	if err != nil {
		log.Fatalf("Failed to create the K8s Discovery client: %s", err)
	}

	ResourceAndGV := DiscoverResourceNameAndPreferredGV(disco)

	for _, dpa := range kubeAPIs {
		// We only want deprecated APIs :)
		if !dpa.Deprecated {
			continue
		}

		group, version, kind := dpa.Group, dpa.Version, dpa.Kind
		var gvk string

		if group != "" {
			gvk = fmt.Sprintf("%s/%s/%s", group, version, kind)
		} else {
			gvk = fmt.Sprintf("%s/%s", version, kind)
		}

		if _, ok := ResourceAndGV[gvk]; !ok {
			log.Debugf("Skipping the resource %s because it doesn't exists in the APIServer", gvk)
			continue
		}

		prefResource := ResourceAndGV[gvk]

		if prefResource.ResourceName == "" || prefResource.GroupVersion == "" {
			log.Debugf("Skipping the resource %s because it doesn't exists in the APIServer", gvk)
			continue
		}

		gv, err := schema.ParseGroupVersion(prefResource.GroupVersion)
		if err != nil {
			log.Warnf("Failed to parse GroupVersion %s of resource %s existing in the API Server: %s", prefResource.GroupVersion, prefResource.ResourceName, err)
		}

		gvrPreferred := gv.WithResource(prefResource.ResourceName)

		log.Debugf("Listing objects for %s/%s/%s", group, version, prefResource.ResourceName)
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: prefResource.ResourceName}
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

		// Now let's see if there's a preferred API containing the same objects
		if gvr != gvrPreferred {
			log.Infof("Listing objects for Preferred %s/%s", prefResource.GroupVersion, prefResource.ResourceName)

			listPref, err := client.Resource(gvrPreferred).List(context.TODO(), metav1.ListOptions{})
			if apierrors.IsForbidden(err) {
				log.Fatalf("Failed to list objects in the cluster. Permission denied! Please check if you have the proper authorization")
			}

			if err != nil && !apierrors.IsNotFound(err) {
				log.Fatalf("Failed communicating with k8s while listing objects. \nError: %v", err)
			}
			// If len of the lists is the same we can "assume" they're the same list
			if len(list.Items) == len(listPref.Items) {
				log.Infof("%s/%s/%s contains the same length of %d items that preferred %s/%s with %d items, skipping", group, version, kind, len(list.Items), prefResource.GroupVersion, kind, len(listPref.Items))
				continue
			}
		}

		if len(list.Items) > 0 {
			log.Infof("Found %d deprecated objects of type %s/%s/%s", len(list.Items), group, version, resourceName)
			api := results.DeprecatedAPI{
				Kind:        kind,
				Deprecated:  dpa.Deprecated,
				Group:       group,
				Name:        resourceName,
				Version:     version,
				Description: dpa.Description,
			}

			api.Items = results.ListObjects(list.Items)
			deprecated = append(deprecated, api)
		}
	}

	return deprecated
}
