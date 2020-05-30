package k8sinput

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"

	log "github.com/sirupsen/logrus"
)

// DiscoverResourceName provides a Resource Name based in its Group, Version and Kind
// This is necessary when you're listing all the existing resources in the cluster
// as you've to pass group/version/name (and not group/version/kind) to client.resource.List
func DiscoverResourceName(client *discovery.DiscoveryClient, group, version, kind string) string {
	var gv string
	if group != "" {
		gv = fmt.Sprintf("%s/%s", group, version)
	} else {
		gv = version
	}
	resources, err := client.ServerResourcesForGroupVersion(gv)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ""
		}
		if apierrors.IsForbidden(err) {
			log.Fatalf("Failed to list object %s attribute. Permission denied! Please check if you have the proper authorization", gv)
		}
		log.Fatalf("Failed communicating with k8s while discovering the object name for %s. Error: %v", gv, err)
	}
	for i := range resources.APIResources {
		apires := &resources.APIResources[i]
		if apires.Kind == kind {
			return apires.Name
		}
	}
	return ""
}
