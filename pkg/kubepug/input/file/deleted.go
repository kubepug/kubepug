package fileinput

import (
	"strings"

	"github.com/rikatz/kubepug/pkg/parser"
	"github.com/rikatz/kubepug/pkg/results"
)

// GetDeleted takes a set of FileItems and checks if they still exists in the API
func GetDeleted(FileItems FileItems, KubeAPIs parser.KubernetesAPIs) (deleted []results.DeletedAPI) {

	for key, item := range FileItems {

		// Here we want to skip CRDs, so if there's some object with Group like "pug.rkatz.io" we will skip
		// Valid groups does not contain "." in the middle (like "apps/v1") or if so, they contain the reserved
		// "k8s.io" (like "scheduling.k8s.io")

		var group, version, kind string
		gvk := strings.Split(key, "/")
		if len(gvk) > 2 {
			group = gvk[0]
			version = gvk[1]
			kind = gvk[2]
			if strings.Contains(group, ".") && !strings.Contains("group", "k8s.io") {
				continue
			}
		} else {
			version = gvk[0]
			kind = gvk[1]

		}

		if _, ok := KubeAPIs[key]; !ok {

			api := results.DeletedAPI{
				Kind:    kind,
				Deleted: true,
				Group:   group,
				Version: version,
			}

			api.Items = item
			deleted = append(deleted, api)

		}
	}

	return deleted

}
