package errors

import "errors"

var ErrAPINotFound = errors.New("API definition not found")

// IsAPINotFoundError checks if the err is an error which
// indicates an item wasn't found on the database
func IsErrAPINotFound(e error) bool {
	return e == ErrAPINotFound
}
