package kubepug

import (
	k8sinput "github.com/rikatz/kubepug/pkg/kubepug/input/k8s"

	"github.com/rikatz/kubepug/pkg/parser"
	"github.com/rikatz/kubepug/pkg/results"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// K8sInput defines a struct that will be used when comparing APIs against a K8s Cluster
type K8sInput struct {
	K8sconfig *genericclioptions.ConfigFlags
	K8sapi    parser.KubernetesAPIs
	Apiwalk   bool
	Monitor   bool
}

// ListDeprecated lists the deprecated objects from a Kubernetes cluster
func (i K8sInput) ListDeprecated() (deprecatedapis []results.DeprecatedAPI) {
	deprecatedapis = k8sinput.GetDeprecated(i.K8sapi, i.K8sconfig)
	return deprecatedapis

}

// ListDeleted lists the non-existend objects in some K8s version from a Kubernetes cluster
func (i K8sInput) ListDeleted() (deletedapis []results.DeletedAPI) {
	if i.Apiwalk {
		deletedapis = k8sinput.GetDeleted(i.K8sapi, i.K8sconfig)
	}
	return deletedapis
}
