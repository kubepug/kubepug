package fileinput

import (
	"context"
	"strings"

	"github.com/kubepug/kubepug/pkg/errors"
	"github.com/kubepug/kubepug/pkg/results"
	"github.com/kubepug/kubepug/pkg/store"
	"github.com/kubepug/kubepug/pkg/utils"
	"github.com/sirupsen/logrus"
)

// FileInput defines a struct that will be used when comparing APIs against a File Input file
type FileInput struct {
	FileItems FileItems
	Store     store.DefinitionStorer
	// We will have a IncludeGroup and a IgnoreGroup configs to tune false positives and false negatives
	// If there is an IncludeGroup, only the resources on this group will be parsed
	IncludePrefixGroup []string
	// If an API is inside the IgnoreGroup it will be bypassed
	IgnoreExactGroup []string
}

// NewFileInput returns the struct FileInput already populated
func NewFileInput(location string, storer store.DefinitionStorer) (fileInput *FileInput, err error) {
	fileInput = &FileInput{}
	fileitems, err := GetFileItems(location)
	if err != nil {
		return fileInput, err
	}

	fileInput.Store = storer
	fileInput.FileItems = fileitems
	// The groups below are: externaldns (not core), anything on x-k8s.io, internal flowcontrol and the autoscaling group that is actually a CRD (the real autoscaling is just autoscaling/version)
	fileInput.IgnoreExactGroup = []string{"externaldns.k8s.io", "x-k8s.io", "flowcontrol.apiserver.k8s.io", "autoscaling.k8s.io"}
	fileInput.IncludePrefixGroup = []string{".k8s.io"}

	return fileInput, nil
}

// GetDeprecations retrieves the map of FileItems and compares with Kubepug store
// returning the set of Deprecated results
func (f *FileInput) GetDeprecations() (deprecated, deleted []results.ResultItem, err error) {
	for key, item := range f.FileItems {
		gvk := strings.Split(key, "/")
		var group, version, kind string
		switch len(gvk) {
		// This is a CoreAPI, like v1/Namespace
		case 2:
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

		apiDef, err := f.Store.GetAPIDefinition(context.Background(), group, version, kind)
		if err != nil {
			if !errors.IsErrAPINotFound(err) {
				return deprecated, deleted, err
			}
		}

		if apiDef.DeletedVersion == "" && apiDef.DeprecationVersion == "" {
			continue
		}

		result := results.CreateItem(group, version, kind, item)
		result.Description = apiDef.Description

		if apiDef.Replacement != nil {
			result.Replacement = apiDef.Replacement
		}

		result.K8sVersion = apiDef.DeprecationVersion

		if apiDef.DeletedVersion != "" {
			result.K8sVersion = apiDef.DeletedVersion
			deleted = append(deleted, result)
			continue
		}
		deprecated = append(deprecated, result)
	}

	return deprecated, deleted, nil
}
