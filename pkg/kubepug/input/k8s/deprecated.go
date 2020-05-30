package k8sinput

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/rikatz/kubepug/pkg/parser"
	"github.com/rikatz/kubepug/pkg/results"
	log "github.com/sirupsen/logrus"
)

// GetDeprecated receives the Map of Deprecated API and List the existent Deprecated Objects in the Cluster
func GetDeprecated(KubeAPIs parser.KubernetesAPIs, config *genericclioptions.ConfigFlags) (deprecated []results.DeprecatedAPI) {

	var resourceName string

	configRest, err := config.ToRESTConfig()
	if err != nil {
		log.Fatalf("Failed to create the K8s config parameters while listing Deprecated objects")
	}

	client, err := dynamic.NewForConfig(configRest)
	if err != nil {
		log.Fatalf("Failed to create the K8s client while listing Deprecated objects")
	}

	// Feed the KubeAPIs with the resourceName as this is used to the K8s Resource lister
	disco, err := discovery.NewDiscoveryClientForConfig(configRest)

	if err != nil {
		log.Fatalf("Failed to create the K8s Discovery client")
	}

	for _, dpa := range KubeAPIs {
		// We only want deprecated APIs :)
		if !dpa.Deprecated {
			continue
		}
		group, version, kind := dpa.Group, dpa.Version, dpa.Kind

		if resourceName = DiscoverResourceName(disco, group, version, kind); resourceName == "" {
			// If no ResourceName is found in the API Server this Resource does not exists and should
			// be ignored
			log.Debugf("Skipping the resource %s/%s/%s because it doesn't exists in the APIServer", group, version, kind)
			continue
		}

		log.Debugf("Listing objects for %s/%s/%s", group, version, resourceName)
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resourceName}
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
