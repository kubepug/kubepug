package formatter

import (
	"github.com/rikatz/kubepug/pkg/results"
)

// Formatter defines the behavior for a Formatter
type Formatter interface {
	Output(results results.Result) ([]byte, error)
}

// NewFormatter returns a new instance of formatter
func NewFormatter(t string) Formatter {
	switch t {
	case "apiversions":
		return newAPIVersionsFormatter()
	case "stdout":
		return newSTDOUTFormatter()
	case "plain":
		return newPlainFormatter()
	case "json":
		return newJSONFormatter()
	case "yaml":
		return newYamlFormatter()
	default:
		return newSTDOUTFormatter()
	}
}
