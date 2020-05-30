package kubepug

import (
	fileinput "github.com/rikatz/kubepug/pkg/kubepug/input/file"
	"github.com/rikatz/kubepug/pkg/parser"
	"github.com/rikatz/kubepug/pkg/results"
)

// FileInput defines a struct that will be used when comparing APIs against a File Input file
type FileInput struct {
	K8sapi    parser.KubernetesAPIs
	FileItems fileinput.FileItems
}

// NewFileInput returns the struct FileInput already populated
func NewFileInput(location string, K8sapi parser.KubernetesAPIs) (FileInput FileInput) {

	FileInput.K8sapi = K8sapi
	FileInput.FileItems = fileinput.GetFileItems(location)

	return FileInput
}

// ListDeprecated lists the deprecated objects from a FileInput file
func (i FileInput) ListDeprecated() (deprecatedapis []results.DeprecatedAPI) {
	deprecatedapis = fileinput.GetDeprecated(i.FileItems, i.K8sapi)
	return deprecatedapis
}

// ListDeleted lists the non-existend objects in some K8s version from a FileInput file
func (i FileInput) ListDeleted() (deletedapis []results.DeletedAPI) {
	deletedapis = fileinput.GetDeleted(i.FileItems, i.K8sapi)
	return deletedapis
}
