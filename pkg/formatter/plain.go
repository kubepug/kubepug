package formatter

import (
	"fmt"

	"github.com/rikatz/kubepug/pkg/kubepug"
)

type plain struct{}

func newPlainFormatter() Formatter {
	return &plain{}
}

func (f *plain) Output(results kubepug.Result) ([]byte, error) {
	s := fmt.Sprintf("RESULTS:\nDeprecated APIs:\n\n")
	for _, api := range results.DeprecatedAPIs {
		s = fmt.Sprintf("%s%s found in %s/%s\n", s, api.Kind, api.Group, api.Version)
		if api.Description != "" {
			s = fmt.Sprintf("%sDescription: %s\n", s, api.Description)
		}
		for _, i := range api.Items {
			if i.Namespace != "" {
				s = fmt.Sprintf("%s%s: %s namespace: %s\n", s, i.Kind, i.Name, i.Namespace)
			} else {
				s = fmt.Sprintf("%s%s: %s \n", s, i.Kind, i.Name)
			}
		}
		s = fmt.Sprintf("%s\n", s)
	}
	return []byte(s), nil
}
