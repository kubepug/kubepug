package formatter

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rikatz/kubepug/pkg/results"
)

type stdout struct{}

func newSTDOUTFormatter() Formatter {
	return &stdout{}
}

var gvColor = color.New(color.FgHiBlack).Add(color.Bold).SprintFunc()
var resourceColor = color.New(color.FgRed).Add(color.Bold).SprintFunc()
var globalColor = color.New(color.FgBlue).Add(color.Bold).SprintFunc()
var namespaceColor = color.New(color.FgCyan).Add(color.Bold).SprintFunc()
var errorColor = color.New(color.FgWhite).Add(color.BgRed).Add(color.Bold).SprintFunc()

func (f *stdout) Output(results results.Result) ([]byte, error) {
	s := fmt.Sprintf("%s:\n%s:\n\n", resourceColor("RESULTS"), resourceColor("Deprecated APIs"))
	for _, api := range results.DeprecatedAPIs {
		s = fmt.Sprintf("%s%s found in %s/%s\n", s, resourceColor(api.Kind), gvColor(api.Group), gvColor(api.Version))
		if api.Description != "" {
			s = fmt.Sprintf("%s\t ├─ %s\n", s, api.Description)
		}
		items := stdoutListItems(api.Items)
		s = fmt.Sprintf("%s%s\n", s, items)
	}
	s = fmt.Sprintf("%s\n%s:\n\n", s, resourceColor("Deleted APIs"))
	for _, api := range results.DeletedAPIs {
		s = fmt.Sprintf("%s%s found in %s/%s\n", s, resourceColor(api.Kind), gvColor(api.Group), gvColor(api.Version))
		s = fmt.Sprintf("%s\t ├─ %s\n", s, errorColor("API REMOVED FROM THE CURRENT VERSION AND SHOULD BE MIGRATED IMMEDIATELY!!"))
		items := stdoutListItems(api.Items)
		s = fmt.Sprintf("%s%s\n", s, items)
	}
	return []byte(s), nil
}

func stdoutListItems(items []results.Item) string {
	s := fmt.Sprintf("")
	for _, i := range items {
		if i.Namespace != "" {
			s = fmt.Sprintf("%s\t\t-> %s: %s %s %s\n", s, namespaceColor(i.Scope), i.ObjectName, namespaceColor("namespace:"), i.Namespace)
		} else {
			s = fmt.Sprintf("%s\t\t-> %s: %s \n", s, globalColor(i.Scope), i.ObjectName)
		}
	}
	return s
}
