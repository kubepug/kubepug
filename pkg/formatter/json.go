package formatter

import (
	jsonencoding "encoding/json"

	"github.com/rikatz/kubepug/pkg/kubepug"
)

type json struct{}

func newJSONFormatter() Formatter {
	return &json{}
}

func (f *json) Output(results kubepug.Result) ([]byte, error) {
	j, err := jsonencoding.Marshal(results)
	if err != nil {
		return nil, err
	}
	return j, nil
}
