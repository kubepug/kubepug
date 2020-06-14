package results

// DeprecatedAPI definition of an API
type DeprecatedAPI struct {
	Description string `json,yaml:"description,omitempty"`
	Group       string `json,yaml:"group,omitempty"`
	Kind        string `json,yaml:"kind,omitempty"`
	Version     string `json,yaml:"version,omitempty"`
	Name        string `json,yaml:"name,omitempty"`
	// TODO: What is this boolean for? All APIs here aren't already marked as Deprecated?
	Deprecated bool   `json,yaml:"deprecated,omitempty"`
	Items      []Item `json,yaml:"deprecated_items,omitempty"`
}

// Item definition of the Items inside a deprecated API
type Item struct {
	Scope      string `json,yaml:"scope,omitempty"`
	ObjectName string `json,yaml:"objectname,omitempty"`
	Namespace  string `json,yaml:"namespace,omitempty"`
	Location   string `json,yaml:"location,omitempty"`
}

// DeletedAPI definition of an API
type DeletedAPI struct {
	Group   string `json,yaml:"group,omitempty"`
	Kind    string `json,yaml:"kind,omitempty"`
	Version string `json,yaml:"version,omitempty"`
	Name    string `json,yaml:"name,omitempty"`
	// TODO: What is this boolean for? All APIs here aren't already marked as Deleted?
	Deleted bool   `json,yaml:"deleted,omitempty"`
	Items   []Item `json,yaml:"deleted_items,omitempty"`
}

// Result to show final user
type Result struct {
	DeprecatedAPIs []DeprecatedAPI `json,yaml:"deprecated_apis,omitempty"`
	DeletedAPIs    []DeletedAPI    `json,yaml:"deleted_apis,omitempty"`
}
