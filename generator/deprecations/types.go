package deprecations

type GroupVersionKind struct {
	Group   string `json:"group,omitempty"`
	Version string `json:"version,omitempty"`
	Kind    string `json:"kind,omitempty"`
}

type Version struct {
	VersionMajor int `json:"version_major,omitempty"`
	VersionMinor int `json:"version_minor,omitempty"`
}

type APIDeprecation struct {
	GroupVersionKind
	Description       string           `json:"description,omitempty"`
	IntroducedVersion Version          `json:"introduced_version,omitempty"`
	DeprecatedVersion Version          `json:"deprecated_version,omitempty"`
	RemovedVersion    Version          `json:"removed_version,omitempty"`
	Replacement       GroupVersionKind `json:"replacement,omitempty"`
}
