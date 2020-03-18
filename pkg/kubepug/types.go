package kubepug

import "github.com/fatih/color"

var gvColor = color.New(color.FgHiBlack).Add(color.Bold).SprintFunc()
var resourceColor = color.New(color.FgRed).Add(color.Bold).SprintFunc()
var globalColor = color.New(color.FgBlue).Add(color.Bold).SprintFunc()
var namespaceColor = color.New(color.FgCyan).Add(color.Bold).SprintFunc()
var errorColor = color.New(color.FgWhite).Add(color.BgRed).Add(color.Bold).SprintFunc()

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

// Result to show final user
type Result struct {
	DeprecatedAPIs []DeprecatedAPI `json,yaml:"deprecated_apis,omitempty"`
}
