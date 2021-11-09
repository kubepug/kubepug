package formatter

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rikatz/kubepug/pkg/results"
)

type plain struct{}

func newPlainFormatter() Formatter {
	return &plain{}
}

func (f *plain) Output(data results.Result) ([]byte, error) {
	s := "RESULTS:\nDeprecated APIs:\n\n"

	for _, api := range data.DeprecatedAPIs {
		s = fmt.Sprintf("%s%s found in %s/%s\n", s, api.Kind, api.Group, api.Version)
		if api.Description != "" {
			s = fmt.Sprintf("%sDescription: %s\n", s, api.Description)
		}

		items := listItems(api.Items)
		s = fmt.Sprintf("%s%s\n", s, items)
	}

	s = fmt.Sprintf("%s\nDeleted APIs:\n\n", s)

	for _, api := range data.DeletedAPIs {
		s = fmt.Sprintf("%s%s found in %s/%s\n", s, api.Kind, api.Group, api.Version)
		items := listItems(api.Items)
		s = fmt.Sprintf("%s%s\n", s, items)
	}

	return []byte(s), nil
}

func listItems(items []results.Item) string {
	s := ""
	for _, i := range items {
		var fileLocation string
		if i.Location != "" {
			fileLocation = fmt.Sprintf("location: %s", i.Location)
		}

		if i.Scope == "OBJECT" {
			if i.Namespace == "" {
				i.Namespace = metav1.NamespaceDefault
			}

			s = fmt.Sprintf("%s%s: %s namespace: %s %s\n", s, i.Scope, i.ObjectName, i.Namespace, fileLocation)
		} else {
			s = fmt.Sprintf("%s%s: %s %s\n", s, i.Scope, i.ObjectName, fileLocation)
		}
	}

	return s
}
