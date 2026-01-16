// Package multierr contains simple implementation of multiple error handling.
// Inspired by https://github.com/hashicorp/go-multierror.
package multierr

import (
	"errors"
	"fmt"
	"strings"
)

// MultiError is container for multiple errors.
type MultiError []error

// Error returns an error message.
func (m MultiError) Error() string {
	errorMessages := make([]string, 0, len(m))

	for _, err := range m {
		// can contain nil errors
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}

	switch len(errorMessages) {
	case 0:
		return ""
	case 1:
		return errorMessages[0]
	default:
		return fmt.Sprintf("multiple errors returned:\n\t%s", strings.Join(errorMessages, ",\n\t"))
	}
}

// ErrorOrNil returns nil in case there are no errors inside.
func (m MultiError) ErrorOrNil() error {
	if m.Error() == "" {
		return nil
	}

	return m
}

// Is validates whenever any of included errors Is target error.
func (m MultiError) Is(target error) bool {
	if target == nil {
		return false
	}

	for _, err := range m {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}
