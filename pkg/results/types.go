package results

import "github.com/kubepug/kubepug/pkg/apis/v1alpha1"

const (
	namespacedObject = "OBJECT"
	clusterObject    = "GLOBAL"
)

// Item definition of the Items inside a deprecated API
type Item struct {
	Scope      string `json:"scope,omitempty" yaml:"scope,omitempty"`
	ObjectName string `json:"objectname,omitempty" yaml:"objectname,omitempty"`
	Namespace  string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Location   string `json:"location,omitempty" yaml:"location,omitempty"`
}

type ResultItem struct {
	Group       string                     `json:"group,omitempty" yaml:"group,omitempty"`
	Kind        string                     `json:"kind,omitempty" yaml:"kind,omitempty"`
	Version     string                     `json:"version,omitempty" yaml:"version,omitempty"`
	Replacement *v1alpha1.GroupVersionKind `json:"replacement,omitempty" yaml:"replacement,omitempty"`
	// K8sVersion defines which k8s version this API was flagged
	K8sVersion  string `json:"k8sversion,omitempty" yaml:"k8sversion,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Items       []Item `json:"deleted_items,omitempty" yaml:"deleted_items,omitempty"`
}

// Result to show final user
type Result struct {
	DeprecatedAPIs []ResultItem `json:"deprecated_apis" yaml:"deprecated_apis"`
	DeletedAPIs    []ResultItem `json:"deleted_apis" yaml:"deleted_apis"`
}
