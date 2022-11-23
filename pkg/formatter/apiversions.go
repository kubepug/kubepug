package formatter

import (
	"fmt"
	"strings"

	"github.com/rikatz/kubepug/pkg/results"
)

type apiversions struct{}

func newAPIVersionsFormatter() Formatter {
	return &apiversions{}
}

func (f *apiversions) Output(data results.Result) ([]byte, error) {
	s := strings.Join(data.APIVersions, "\n")
	s = fmt.Sprintf("%s\n", s)

	return []byte(s), nil
}
