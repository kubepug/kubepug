package k8sinput

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/kubepug/kubepug/pkg/errors"
	"github.com/kubepug/kubepug/pkg/results"
	"github.com/kubepug/kubepug/pkg/store"
	"github.com/kubepug/kubepug/pkg/utils"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// K8sInput defines a struct that will be used when comparing APIs against a K8s Cluster
type K8sInput struct {
	K8sconfig *genericclioptions.ConfigFlags
	Store     store.DefinitionStorer

	Client          dynamic.Interface
	DiscoveryClient discovery.DiscoveryInterface

	// We will have a IncludeGroup and a IgnoreGroup configs to tune false positives and false negatives
	// If there is an IncludeGroup, only the resources on this group will be parsed
	IncludePrefixGroup []string
	// If an API is inside the IgnoreGroup it will be bypassed
	IgnoreExactGroup []string
}

var deprecatedAPIReplacements = map[string]schema.GroupVersionResource{
	"extensions/v1beta1/Ingress": {
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	},
}

var apisvcgvr = schema.GroupVersionResource{
	Group:    "apiregistration.k8s.io",
	Version:  "v1",
	Resource: "apiservices",
}

func (f *K8sInput) IgnoreAPIService() error {
	apisvcList, err := f.Client.Resource(apisvcgvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get apiservices: %w", err)
	}

	for _, d := range apisvcList.Items {
		_, foundSvc, errSvc := unstructured.NestedString(d.Object, "spec", "service", "name")
		// No services fields or group field found, move on!
		if errSvc != nil || !foundSvc {
			logrus.Infof("local service %s found, skipping", d.GetName())
			continue
		}

		group, foundGrp, err := unstructured.NestedString(d.Object, "spec", "group")
		// No services fields or group field found, move on!
		if err != nil || !foundGrp {
			logrus.Warningf("failed to parse the apiservice %s, this will be skipped and may generate inconsistency. err: %s", d.GetName(), err)
			continue
		}
		f.IgnoreExactGroup = append(f.IgnoreExactGroup, group)
	}
	return nil
}

// GetDeprecated retrieves the map of FileItems and compares with Kubepug store,
// returning the set of Deprecated results
func (f *K8sInput) GetDeprecations() (deprecated, deleted []results.ResultItem, err error) {
	if err := f.IgnoreAPIService(); err != nil {
		return deprecated, deleted, err
	}

	apiresources, err := f.DiscoveryClient.ServerPreferredResources()
	if err != nil {
		if !discovery.IsGroupDiscoveryFailedError(err) {
			return deprecated, deleted, err
		}
		logrus.Warningf("failed to discovery some apiresources, they will be skipped: %s", err)
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
		replacement, err := f.checkForReplacement(gv.Group, gv.Version, resources.APIResources[i].Kind)
		if err != nil {
			return nil, nil, err
		}

		// If there is a proper API replacement, we can just skip
		if replacement != "" {
			logrus.Infof("found replacement %s for %s/%s/%s, skipping", replacement, gv.Group, gv.Version, resources.APIResources[i].Kind)
			continue
		}

		apiResult, err := f.Store.GetAPIDefinition(context.Background(), gv.Group, gv.Version, resources.APIResources[i].Kind)
		if err != nil {
			if !errors.IsErrAPINotFound(err) {
				return deprecated, deleted, err
			}
		}

		if apiResult.DeprecationVersion == "" && apiResult.DeletedVersion == "" {
			continue
		}

		items, err := getResources(f.Client, gv.Group, gv.Version, resources.APIResources[i].Name)
		if err != nil {
			return deprecated, deleted, err
		}

		if len(items) == 0 {
			continue
		}

		result := results.CreateItem(gv.Group, gv.Version, resources.APIResources[i].Kind, items)
		result.Description = apiResult.Description
		if apiResult.Replacement != nil {
			result.Replacement = apiResult.Replacement
		}

		result.K8sVersion = apiResult.DeprecationVersion
		if apiResult.DeletedVersion != "" {
			result.K8sVersion = apiResult.DeletedVersion
			deleted = append(deleted, result)
			continue
		}
		deprecated = append(deprecated, result)
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

// Before checking for the API, we need to verify if it already have a proper replacement on the server.
// The example here is on Ingress API. It formerly existed on a different group (extensions/v1beta1) and
// was migrated to networking.k8s.io/v1, so at some moment a server would have two preferred resources, one for extensions
// and one for networking.k8s.io. In this case, even querying extensions/v1beta1 would return objects and generate a false positive
// so we need to verify if the replacement also exists.
func (f *K8sInput) checkForReplacement(group, version, kind string) (string, error) {
	keyForReplacement := fmt.Sprintf("%s/%s/%s", group, version, kind)
	if replacement, ok := deprecatedAPIReplacements[keyForReplacement]; ok {
		apiresources, err := f.DiscoveryClient.ServerResourcesForGroupVersion(replacement.GroupVersion().String())
		if err != nil {
			if !discovery.IsGroupDiscoveryFailedError(err) {
				return "", err
			}
			logrus.Warningf("failed to get resources for groupversion %s, this can generate a false positive", replacement.GroupVersion().String())
		}
		if apiresources != nil {
			for i := range apiresources.APIResources {
				if apiresources.APIResources[i].Name == replacement.Resource {
					return fmt.Sprintf("%s/%s", replacement.GroupVersion().String(), apiresources.APIResources[i].Kind), nil
				}
			}
		}
	}
	// No replacement was found
	return "", nil
}
