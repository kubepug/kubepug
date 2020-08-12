package kubepug

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rikatz/kubepug/pkg/results"
)

// Deprecator implements an interface for reading some sort of Input and comparing against the
// map of Kubernetes APIs to check if there's some Deprecated or Deleted
type Deprecator interface {
	ListDeprecated() []results.DeprecatedAPI
	ListDeleted() []results.DeletedAPI
}

func MeasureDeprecations(r results.Result, c prometheus.CounterVec) {
	for _, k := range r.DeprecatedAPIs {
		for _, item := range k.Items {
			c.WithLabelValues(k.Group, k.Kind, k.Version, k.Name, item.Scope, item.ObjectName, item.Namespace).Inc()
		}
	}
}

func MeasureDeletions(r results.Result, c prometheus.CounterVec) {
	for _, k := range r.DeletedAPIs {
		for _, item := range k.Items {
			c.WithLabelValues(k.Group, k.Kind, k.Version, k.Name, item.Scope, item.ObjectName, item.Namespace).Inc()
		}
	}
}

// GetDeprecations returns the results of the comparision between the Input and the APIs
func GetDeprecations(d Deprecator) (result results.Result) {
	// TODO(igaskin):  wrapp this in a ticker to emitt prom metrics
	result.DeprecatedAPIs = d.ListDeprecated()
	result.DeletedAPIs = d.ListDeleted()
	return result
}
