package parser

// KubeAPI represents a Kubernetes API defined in swagger.json
type KubeAPI struct {
	Description string
	Group       string
	// kind, as for Kind: Pod
	Kind    string
	Version string
	// Name is the resource name in plural (pods) to be used by the resource lister for dynamic client
	Name       string
	Deprecated bool
}

// KubernetesAPIs is a map of KubeAPI objects
type KubernetesAPIs map[string]KubeAPI
