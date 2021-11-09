package formatter

import (
	yamlencoder "gopkg.in/yaml.v3"

	"github.com/rikatz/kubepug/pkg/results"
)

type yaml struct{}

func newYamlFormatter() Formatter {
	return &yaml{}
}

func (f *yaml) Output(data results.Result) ([]byte, error) {
	y, err := yamlencoder.Marshal(data)
	if err != nil {
		return nil, err
	}

	return y, nil
}
