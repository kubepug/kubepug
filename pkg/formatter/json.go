package formatter

import (
	jsonencoding "encoding/json"

	"github.com/rikatz/kubepug/pkg/results"
)

type json struct{}

func newJSONFormatter() Formatter {
	return &json{}
}

func (f *json) Output(data results.Result) ([]byte, error) {
	j, err := jsonencoding.Marshal(data)
	if err != nil {
		return nil, err
	}

	return j, nil
}
