package swaggerstore

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

const (
	// SwaggerStore cannot assert the deleted version, so just return some string
	internalStatusVersion = "unknown"
)
