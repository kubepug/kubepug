package store

import (
	"context"

	api "github.com/kubepug/kubepug/pkg/apis/v1alpha1"
)

type DefinitionStorer interface {
	// GetAPIDefinition returns the description, if the API is deprecated or an error.
	// The error may be of type ErrAPINotFound, which means the API is deleted
	GetAPIDefinition(ctx context.Context, group, version, kind string) (api.APIVersionStatus, error)
}
