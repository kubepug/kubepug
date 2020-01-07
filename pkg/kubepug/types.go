package kubepug

// DeprecatedAPI defines an API to be searched into Kubernetes with its group, kind and version
type DeprecatedAPI struct {
	description string
	group       string
	kind        string
	version     string
	name        string
}
