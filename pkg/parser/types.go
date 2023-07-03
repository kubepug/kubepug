package parser

const (
	CoreAPI = "CORE"
)

// APIVersionStatus represents the status of a group/kind/version
type APIVersionStatus struct {
	Description string
	Deprecated  bool
}

// APIVersion is an APIVersion of a group/kind that will be queried on description
// and if it is deprecated
type APIVersion map[string]APIVersionStatus

// APIKind contains a Kind of API (like "Ingress") and may also be populated with
// the resource name
type APIKinds map[string]APIVersion

// APIGroups contains a map of groups of APIs that exists. Eg.: networking.k8s.io
type APIGroups map[string]APIKinds

type ManifestsFiles map[string]struct {
	Filename  string
	APIGroups APIGroups
}

// definitionsJson defines the definitions structure to be unmarshalled
type definitionsJSON struct {
	Definitions map[string]struct {
		Description      string `json:"description,omitempty"`
		GroupVersionKind []struct {
			Group   string `json:"group,omitempty"`
			Version string `json:"version,omitempty"`
			Kind    string `json:"kind,omitempty"`
		} `json:"x-kubernetes-group-version-kind,omitempty"`
	} `json:"definitions,omitempty"`
}
