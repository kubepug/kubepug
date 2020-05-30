package fileinput

import (
	"github.com/rikatz/kubepug/pkg/parser"
	"github.com/rikatz/kubepug/pkg/results"
)

// GetDeprecated retrieves the map of FileItems and compares with Kubernetes swagger.json
// returning the set of Deprecated results
func GetDeprecated(FileItems FileItems, KubeAPIs parser.KubernetesAPIs) (deprecated []results.DeprecatedAPI) {

	for key, item := range FileItems {
		if kubeapi, ok := KubeAPIs[key]; ok {
			if kubeapi.Deprecated {
				api := results.DeprecatedAPI{
					Kind:        kubeapi.Kind,
					Deprecated:  kubeapi.Deprecated,
					Group:       kubeapi.Group,
					Version:     kubeapi.Version,
					Description: kubeapi.Description,
				}

				api.Items = item
				deprecated = append(deprecated, api)
			}
		}
	}

	return deprecated

}
