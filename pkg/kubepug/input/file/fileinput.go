package fileinput

import (
	"strings"

	"github.com/rikatz/kubepug/pkg/parser"
	"github.com/rikatz/kubepug/pkg/results"
	"github.com/rikatz/kubepug/pkg/utils"
	"github.com/sirupsen/logrus"
)

// FileInput defines a struct that will be used when comparing APIs against a File Input file
type FileInput struct {
	FileItems FileItems
	Database  parser.APIGroups
	// We will have a IncludeGroup and a IgnoreGroup configs to tune false positives and false negatives
	// If there is an IncludeGroup, only the resources on this group will be parsed
	IncludePrefixGroup []string
	// If an API is inside the IgnoreGroup it will be bypassed
	IgnoreExactGroup []string
}

// NewFileInput returns the struct FileInput already populated
func NewFileInput(location string, k8sapi parser.APIGroups) (fileInput *FileInput, err error) {
	fileInput = &FileInput{}
	fileitems, err := GetFileItems(location)
	if err != nil {
		return fileInput, err
	}

	fileInput.Database = k8sapi
	fileInput.FileItems = fileitems
	// The groups below are: externaldns (not core), anything on x-k8s.io, internal flowcontrol and the autoscaling group that is actually a CRD (the real autoscaling is just autoscaling/version)
	fileInput.IgnoreExactGroup = []string{"externaldns.k8s.io", "x-k8s.io", "flowcontrol.apiserver.k8s.io", "autoscaling.k8s.io"}
	fileInput.IncludePrefixGroup = []string{".k8s.io"}

	return fileInput, nil
}

var listItems = func(items []results.Item) parser.ListerFunc {
	return func(group, version, resource, kind string) (results.ResultItem, error) {
		return results.CreateItem(group, version, kind, items), nil
	}
}

// GetDeprecated retrieves the map of FileItems and compares with Kubernetes swagger.json
// returning the set of Deprecated results
func (f *FileInput) GetDeprecations() (deprecated, deleted []results.ResultItem, err error) {
	for key, item := range f.FileItems {
		gvk := strings.Split(key, "/")
		var group, version, kind string
		switch len(gvk) {
		// This is a CoreAPI, like v1/Namespace
		case 2:
			group = parser.CoreAPI
			version = gvk[0]
			kind = gvk[1]
		case 3:
			group = gvk[0]
			version = gvk[1]
			kind = gvk[2]
		default:
			logrus.Info("unknown API type, skipping")
			continue
		}

		if !utils.ShouldParse(group, f.IgnoreExactGroup, f.IncludePrefixGroup) {
			continue
		}

		result, isdeleted, err := f.Database.CheckForItem(group, version, kind, kind, listItems(item))
		if err != nil {
			return deprecated, deleted, err
		}

		if result != nil {
			if isdeleted {
				deleted = append(deleted, *result)
				continue
			}
			deprecated = append(deprecated, *result)
		}
	}

	return deprecated, deleted, nil
}
