package k8sinput

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/rikatz/kubepug/pkg/parser"
	"github.com/rikatz/kubepug/pkg/results"
	"github.com/rikatz/kubepug/pkg/utils"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// K8sInput defines a struct that will be used when comparing APIs against a K8s Cluster
type K8sInput struct {
	K8sconfig *genericclioptions.ConfigFlags
	Database  parser.APIGroups
	APIWalk   bool

	Client          dynamic.Interface
	DiscoveryClient discovery.DiscoveryInterface

	// We will have a IncludeGroup and a IgnoreGroup configs to tune false positives and false negatives
	// If there is an IncludeGroup, only the resources on this group will be parsed
	IncludePrefixGroup []string
	// If an API is inside the IgnoreGroup it will be bypassed
	IgnoreExactGroup []string
}

var listItems = func(client dynamic.Interface, apigroup string) parser.ListerFunc {
	return func(group, version, resource, kind string) (results.ResultItem, error) {
		result := results.ResultItem{}
		items, err := getResources(client, group, version, resource)
		if err != nil {
			return result, err
		}
		return results.CreateItem(group, version, kind, items), nil
	}
}

// GetDeprecated retrieves the map of FileItems and compares with Kubernetes swagger.json
// returning the set of Deprecated results
func (f *K8sInput) GetDeprecations() (deprecated, deleted []results.ResultItem, err error) {
	apiresources, err := f.DiscoveryClient.ServerPreferredResources()
	if err != nil {
		return deprecated, deleted, err
	}

	for _, reslist := range apiresources {
		deprecatedRes, deletedRes, err := f.getResourceDeprecation(reslist)
		if err != nil {
			return deprecated, deleted, err
		}
		deprecated = append(deprecated, deprecatedRes...)
		deleted = append(deleted, deletedRes...)
	}

	return deprecated, deleted, nil
}

func (f *K8sInput) getResourceDeprecation(resources *metav1.APIResourceList) (deprecated, deleted []results.ResultItem, err error) {
	if resources == nil {
		return nil, nil, nil
	}
	gv, err := schema.ParseGroupVersion(resources.GroupVersion)
	if err != nil {
		logrus.Warningf("couldn't parse group %s, skipping", resources.GroupVersion)
		return nil, nil, nil
	}

	if !utils.ShouldParse(gv.Group, f.IgnoreExactGroup, f.IncludePrefixGroup) {
		logrus.Info("Ignoring group", gv.Group)
		return nil, nil, nil
	}

	for i := range resources.APIResources {
		// We use internally the Core API string to get our groups
		group := gv.Group
		if group == "" {
			group = parser.CoreAPI
		}

		result, isdeleted, err := f.Database.CheckForItem(group, gv.Version, resources.APIResources[i].Kind, resources.APIResources[i].Name, listItems(f.Client, gv.Group))
		if err != nil {
			return deprecated, deleted, err
		}

		if result != nil {
			if isdeleted {
				deleted = append(deleted, *result)
				continue
			}
			deprecated = append(deprecated, *result)
		}
	}
	return deprecated, deleted, nil
}

func getResources(dynClient dynamic.Interface, group, version, resource string) ([]results.Item, error) {
	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	list, err := dynClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if apierrors.IsNotFound(err) || apierrors.IsMethodNotSupported(err) {
		return make([]results.Item, 0), nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to List objects of type %s/%s/%s. \nError: %v", group, version, resource, err)
	}

	items := results.ListObjects(list.Items)

	return items, nil
}
