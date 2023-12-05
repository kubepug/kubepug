package formatter

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubepug/kubepug/pkg/results"
)

const (
	footer = "Kubepug validates the APIs using Kubernetes markers. To know what are the deprecated and deleted APIS it checks, please go to https://kubepug.xyz/status/"
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

	s := sliceBuilder{}
	if len(data.DeprecatedAPIs) > 0 {
		s.add(resourceColor("RESULTS"), ":\n", resourceColor("Deprecated APIs"), ":\n")

		for _, api := range data.DeprecatedAPIs {
			s.add(resourceColor(api.Kind), " found in ", gvColor(api.Group), "/", gvColor(api.Version), "\n")

			if api.K8sVersion != "" && api.K8sVersion != "unknown" {
				s.add("\t ├─ ", namespaceColor("Deprecated at:"), " ", api.K8sVersion, "\n")
			}

			if api.Replacement != nil {
				s.add("\t ├─ ", namespaceColor("Replacement:"), " ", api.Replacement.Group, "/", api.Replacement.Version, "/", api.Replacement.Kind, "\n")
			}

			if api.Description != "" {
				s.add("\t ├─ ", strings.ReplaceAll(api.Description, "\n", ""), "\n")
			}

			s.addItems(api.Items)
		}
	}

	if len(data.DeletedAPIs) > 0 {
		s.add("\n", resourceColor("Deleted APIs"), ":\n")
		s.add("\t ", errorColor("APIs REMOVED FROM THE CURRENT VERSION AND SHOULD BE MIGRATED IMMEDIATELY!!"), "\n")

		for _, api := range data.DeletedAPIs {
			s.add(resourceColor(api.Kind), " found in ", gvColor(api.Group), "/", gvColor(api.Version), "\n")

			if api.K8sVersion != "" && api.K8sVersion != "unknown" {
				s.add("\t ├─ ", namespaceColor("Deleted at:"), " ", api.K8sVersion, "\n")
			}

			if api.Replacement != nil {
				s.add("\t ├─ ", namespaceColor("Replacement:"), " ", api.Replacement.Group, "/", api.Replacement.Version, "/", api.Replacement.Kind, "\n")
			}

			if api.Description != "" {
				s.add("\t ├─ ", strings.ReplaceAll(api.Description, "\n", ""), "\n")
			}

			s.addItems(api.Items)
		}
	}

	if len(data.DeletedAPIs) == 0 && len(data.DeprecatedAPIs) == 0 {
		s.add("\nNo deprecated or deleted APIs found")
	}

	s.add("\n\n", footer, "\n")

	out := s.String()
	if f.plain {
		out = strings.ReplaceAll(out, "\t", "")
	}

	return []byte(out), nil
}

// sliceBuilder is a String Builder that accepts any number of strings at once for ergonomics.
type sliceBuilder struct {
	strings.Builder
}

func (b *sliceBuilder) add(strs ...string) {
	for _, s := range strs {
		// strings.Builder.WriteString() error is always nil and can be ignored.
		b.WriteString(s)
	}
}

func (b *sliceBuilder) addItems(items []results.Item) {
	for _, i := range items {
		var fileLocation string
		if i.Location != "" {
			fileLocation = fmt.Sprintf("%s %s", locationColor("location:"), i.Location)
		}

		if i.Scope == "OBJECT" {
			if i.Namespace == "" {
				i.Namespace = metav1.NamespaceDefault
			}

			b.add("\t\t-> ", namespaceColor(i.Scope), ": ", i.ObjectName, " ", namespaceColor("namespace:"), " ", i.Namespace, " ", fileLocation, "\n")
		} else {
			b.add("\t\t-> ", globalColor(i.Scope), ": ", i.ObjectName, " ", fileLocation, "\n")
		}
	}
	b.add("\n")
}
