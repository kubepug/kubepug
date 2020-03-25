package kubepug

// KubeAPI represents a Kubernetes API defined in swagger.json
type KubeAPI struct {
	description string
	group       string
	kind        string
	version     string
	name        string
	deprecated  bool
}

// KubernetesAPIs is a map of KubeAPI objects
type KubernetesAPIs map[string]KubeAPI

// DeprecatedAPI definition of an API
type DeprecatedAPI struct {
	Description string           `json,yaml:"description,omitempty"`
	Group       string           `json,yaml:"group,omitempty"`
	Kind        string           `json,yaml:"kind,omitempty"`
	Version     string           `json,yaml:"version,omitempty"`
	Name        string           `json,yaml:"name,omitempty"`
	Deprecated  bool             `json,yaml:"deprecated,omitempty"`
	Items       []DeprecatedItem `json,yaml:"deprecated_items,omitempty"`
}

// DeprecatedItem definition of the Items inside a deprecated API
type DeprecatedItem struct {
	Kind      string `json,yaml:"kind,omitempty"`
	Name      string `json,yaml:"name,omitempty"`
	Namespace string `json,yaml:"namespace,omitempty"`
}

// DeletedAPI definition of an API
type DeletedAPI struct {
	Group   string           `json,yaml:"group,omitempty"`
	Kind    string           `json,yaml:"kind,omitempty"`
	Version string           `json,yaml:"version,omitempty"`
	Name    string           `json,yaml:"name,omitempty"`
	Deleted bool             `json,yaml:"deleted,omitempty"`
	Items   []DeprecatedItem `json,yaml:"deleted_items,omitempty"`
}

// Result to show final user
type Result struct {
	DeprecatedAPIs []DeprecatedAPI `json,yaml:"deprecated_apis,omitempty"`
	DeletedAPIs    []DeletedAPI    `json,yaml:"deleted_apis,omitempty"`
}
