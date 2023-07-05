package store

import "context"

type DefinitionStorer interface {
	// GetAPIDefinition returns the description, if the API is deprecated or an error.
	// The error may be of type ErrAPINotFound, which means the API is deleted
	GetAPIDefinition(ctx context.Context, group, version, kind string) (string, bool, error)
}
