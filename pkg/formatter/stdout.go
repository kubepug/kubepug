package formatter

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rikatz/kubepug/pkg/results"
)

const (
	footer = "Kubepug validates the APIs using Kubernetes markers. To know what are the deprecated and deleted APIS it checks, please go to https://kubepug.xyz/status/ \n"
)

type stdout struct {
	plain bool
}

func newSTDOUTFormatter(plain bool) Formatter {
	return &stdout{
		plain: plain,
	}
}

var (
	gvColor        = color.New(color.FgHiBlack).Add(color.Bold).SprintFunc()
	resourceColor  = color.New(color.FgRed).Add(color.Bold).SprintFunc()
	globalColor    = color.New(color.FgBlue).Add(color.Bold).SprintFunc()
	namespaceColor = color.New(color.FgCyan).Add(color.Bold).SprintFunc()
	errorColor     = color.New(color.FgWhite).Add(color.BgRed).Add(color.Bold).SprintFunc()
	locationColor  = color.New(color.FgHiMagenta).Add(color.Bold).SprintFunc()
)

func (f *stdout) Output(data results.Result) ([]byte, error) {
	color.NoColor = f.plain

	var s string
	if len(data.DeprecatedAPIs) > 0 {
		s = fmt.Sprintf("%s:\n%s:\n", resourceColor("RESULTS"), resourceColor("Deprecated APIs"))

		for _, api := range data.DeprecatedAPIs {
			s = fmt.Sprintf("%s%s found in %s/%s\n", s, resourceColor(api.Kind), gvColor(api.Group), gvColor(api.Version))

			if api.K8sVersion != "" && api.K8sVersion != "unknown" {
				s = fmt.Sprintf("%s\t ├─ %s %s\n", s, namespaceColor("Deprecated at:"), api.K8sVersion)
			}

			if api.Replacement != nil {
				s = fmt.Sprintf("%s\t ├─ %s %s/%s/%s \n", s, namespaceColor("Replacement:"), api.Replacement.Group, api.Replacement.Version, api.Replacement.Kind)
			}

			if api.Description != "" {
				s = fmt.Sprintf("%s\t ├─ %s\n", s, strings.ReplaceAll(api.Description, "\n", ""))
			}

			items := stdoutListItems(api.Items)
			s = fmt.Sprintf("%s%s\n", s, items)
		}
	}

	if len(data.DeletedAPIs) > 0 {
		s = fmt.Sprintf("%s\n%s:\n", s, resourceColor("Deleted APIs"))
		s = fmt.Sprintf("%s\t %s\n", s, errorColor("APIs REMOVED FROM THE CURRENT VERSION AND SHOULD BE MIGRATED IMMEDIATELY!!"))

		for _, api := range data.DeletedAPIs {
			s = fmt.Sprintf("%s%s found in %s/%s\n", s, resourceColor(api.Kind), gvColor(api.Group), gvColor(api.Version))

			if api.K8sVersion != "" && api.K8sVersion != "unknown" {
				s = fmt.Sprintf("%s\t ├─ %s %s\n", s, namespaceColor("Deleted at:"), api.K8sVersion)
			}

			if api.Replacement != nil {
				s = fmt.Sprintf("%s\t ├─ %s %s/%s/%s \n", s, namespaceColor("Replacement:"), api.Replacement.Group, api.Replacement.Version, api.Replacement.Kind)
			}

			if api.Description != "" {
				s = fmt.Sprintf("%s\t ├─ %s\n", s, strings.ReplaceAll(api.Description, "\n", ""))
			}

			items := stdoutListItems(api.Items)
			s = fmt.Sprintf("%s%s\n", s, items)
		}
	}

	if len(data.DeletedAPIs) == 0 && len(data.DeprecatedAPIs) == 0 {
		s = "\nNo deprecated or deleted APIs found"
	}

	s = fmt.Sprintf("%s\n\n%s", s, footer)

	if f.plain {
		s = strings.ReplaceAll(s, "\t", "")
	}

	return []byte(s), nil
}

func stdoutListItems(items []results.Item) string {
	s := ""
	for _, i := range items {
		var fileLocation string
		if i.Location != "" {
			fileLocation = fmt.Sprintf("%s %s", locationColor("location:"), i.Location)
		}

		if i.Scope == "OBJECT" {
			if i.Namespace == "" {
				i.Namespace = metav1.NamespaceDefault
			}

			s = fmt.Sprintf("%s\t\t-> %s: %s %s %s %s\n", s, namespaceColor(i.Scope), i.ObjectName, namespaceColor("namespace:"), i.Namespace, fileLocation)
		} else {
			s = fmt.Sprintf("%s\t\t-> %s: %s %s\n", s, globalColor(i.Scope), i.ObjectName, fileLocation)
		}
	}

	return s
}
