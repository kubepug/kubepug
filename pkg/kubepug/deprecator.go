package kubepug

import (
	"github.com/rikatz/kubepug/pkg/results"
)

// Deprecator implements an interface for reading some sort of Input and comparing against the
// map of Kubernetes APIs to check if there's some Deprecated or Deleted
type Deprecator interface {
	ListDeprecated() []results.DeprecatedAPI
	ListDeleted() []results.DeletedAPI
}

// GetDeprecations returns the results of the comparision between the Input and the APIs
func GetDeprecations(d Deprecator) (result results.Result) {
	result.DeprecatedAPIs = d.ListDeprecated()
	result.DeletedAPIs = d.ListDeleted()
	return result
}
