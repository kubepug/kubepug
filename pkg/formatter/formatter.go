package formatter

import (
	"fmt"

	"github.com/kubepug/kubepug/pkg/results"
)

// Formatter defines the behavior for a Formatter
type Formatter interface {
	Output(results results.Result) ([]byte, error)
}

// NewFormatter returns a new instance of formatter
func NewFormatter(t string) Formatter {
	f, err := NewFormatterWithError(t)
	if err != nil {
		f = newSTDOUTFormatter(false)
	}
	return f
}

// NewFormatterWithError returns a formatter or an error that can be returned by the
// formatter instance or in case the formatter is invalid
func NewFormatterWithError(t string) (Formatter, error) {
	switch t {
	case "stdout":
		return newSTDOUTFormatter(false), nil
	case "plain":
		return newSTDOUTFormatter(true), nil
	case "json":
		return newJSONFormatter(), nil
	case "yaml":
		return newYamlFormatter(), nil
	default:
		return nil, fmt.Errorf("invalid formatter selected: %s", t)
	}
}
