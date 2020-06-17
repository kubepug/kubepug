package k8sinput

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"

	log "github.com/sirupsen/logrus"
)

// ResourceStruct define a Group/Version/ResourceName to be used in the PreferredResource Map
type ResourceStruct struct {
	GroupVersion, ResourceName string
}

// PreferredResource is a map that with a given Kind returns a ResourceName and the preferred API
type PreferredResource map[string]ResourceStruct

// DiscoverResourceNameAndPreferredGV provides a Resource Name and the preferred Group/Version based in its Kind
// This is necessary when you're listing all the existing resources in the cluster
// as you've to pass group/version/name (and not group/version/kind) to client.resource.List
// and also to verify if the server supports newer version of some API and it's not deprecated
func DiscoverResourceNameAndPreferredGV(client *discovery.DiscoveryClient) PreferredResource {

	pr := make(PreferredResource)

	resourcelist, err := client.ServerPreferredResources()
	if err != nil {
		if apierrors.IsNotFound(err) {
			return pr
		}
		if apierrors.IsForbidden(err) {
			log.Fatalf("Failed to list objects for Name discovery. Permission denied! Please check if you have the proper authorization")
		}
		log.Fatalf("Failed communicating with k8s while discovering the object preferred name and gv. Error: %v", err)
	}
	for _, rl := range resourcelist {
		for _, resources := range rl.APIResources {
			item := ResourceStruct{
				GroupVersion: rl.GroupVersion,
				ResourceName: resources.Name,
			}
			pr[resources.Kind] = item

		}
	}
	return pr
}
