package provider

import (
	"errors"
	"fmt"
)

// NotFoundError is returned when a secret cannot be located in the backend.
type NotFoundError struct {
	Path string
	Key  string
}

func (e *NotFoundError) Error() string {
	if e.Key == "" {
		return fmt.Sprintf("secret not found: path=%q", e.Path)
	}
	return fmt.Sprintf("secret not found: path=%q key=%q", e.Path, e.Key)
}

// ErrNotFound constructs a *NotFoundError for the given path/key pair.
func ErrNotFound(path, key string) *NotFoundError {
	return &NotFoundError{Path: path, Key: key}
}

// IsNotFound reports whether err (or any error in its chain) is a
// *NotFoundError.
func IsNotFound(err error) bool {
	var nfe *NotFoundError
	return errors.As(err, &nfe)
}
