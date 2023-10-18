package kubepug

import (
	"github.com/kubepug/kubepug/pkg/results"
)

// Deprecator implements an interface for reading some sort of Input and comparing against the
// map of Kubernetes APIs to check if there's some Deprecated or Deleted
type Deprecator interface {
	GetDeprecations() (deprecated []results.ResultItem, deleted []results.ResultItem, err error)
}

// GetDeprecations returns the results of the comparison between the Input and the APIs
func GetDeprecations(d Deprecator) (result results.Result, err error) {
	deprecated, deleted, err := d.GetDeprecations()
	if err != nil {
		return result, err
	}
	result.DeprecatedAPIs = deprecated
	result.DeletedAPIs = deleted

	return result, nil
}
