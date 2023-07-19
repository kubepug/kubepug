package v1alpha1

const (
	CoreAPI = "CORE"
)

type GroupVersionKind struct {
	Group   string `json:"group,omitempty"`
	Version string `json:"version,omitempty"`
	Kind    string `json:"kind,omitempty"`
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

// APIVersionStatus represents a result from a store query
type APIVersionStatus struct {
	// Description represents the description of the queried API
	Description string `json:"description,omitempty"`
	// DeprecationVersion represents when this API was marked as deprecated
	DeprecationVersion string `json:"deprecationVersion,omitempty"`
	// DeletedVersion represents when this API was marked as deleted
	DeletedVersion string `json:"deletedVersion,omitempty"`
	// IntroducedVersion represents when this API was introduced
	IntroducedVersion string `json:"introducedVersion,omitempty"`
	// Replacement represents what is the proper replacement of this API
	Replacement *GroupVersionKind `json:"replacement,omitempty"`
}
