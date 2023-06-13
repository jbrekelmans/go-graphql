package graphql

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Error is an error type used by *Client to feed back GraphQL-level errors.
type Error struct {
	// Err is the wrapped error.
	Err error

	// Errors are the response errors, if available.
	// Errors reflects the "errors" property of JSON objects in HTTP response bodies.
	// See https://spec.graphql.org/.
	Errors []ErrorItem

	Message string

	// Operation is the GraphQL query/mutation/operation for which the error occurred.
	Operation string
}

var _ error = (*Error)(nil)

// NewErrorf creates an *Error with printf-style error text.
// Like fmt.Errorf, NewErrorf respects %w format specifiers to
// create wrapped errors.
func NewErrorf(format string, args ...any) *Error {
	err := fmt.Errorf(format, args...)
	return &Error{
		Message: err.Error(),
		Err:     errors.Unwrap(err),
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Unwrap supports Golang 1.13+ error wrapping. See https://go.dev/blog/go1.13-errors
func (e *Error) Unwrap() error {
	return e.Err
}

// ErrorItem is a response error. See https://spec.graphql.org/.
type ErrorItem struct {
	// Error message.
	Message string

	// Raw entries as per the JSON value returned by the server.
	// This can be used to get extensions, path and locations entries.
	// See https://spec.graphql.org/.
	Raw map[string]json.RawMessage
}

// MarshalJSON implements the Marshaler interface.
func (e *ErrorItem) MarshalJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}
	return json.Marshal(e.Raw)
}

// UnmarshalJSON implements the Unmarshaler interface.
func (e *ErrorItem) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	err := json.Unmarshal(b, &e.Raw)
	if err != nil {
		return fmt.Errorf(`error in (*graphql.ErrorItem).UnmarshalJSON: %w`, err)
	}
	if msgRaw, ok := e.Raw["message"]; ok {
		// Ignore errors
		_ = json.Unmarshal(msgRaw, &e.Message)
	}
	return nil
}

func getOrCreateError(err error) *Error {
	structuredErr, ok := err.(*Error)
	if !ok {
		structuredErr = &Error{
			Err:     err,
			Message: err.Error(),
		}
	}
	return structuredErr
}

func setErrorOperation(err error, operation string) error {
	if err == nil {
		return nil
	}
	structuredErr := getOrCreateError(err)
	structuredErr.Operation = operation
	return structuredErr
}

func setErrorItems(err error, errors []ErrorItem) error {
	if err == nil {
		return nil
	}
	structuredErr := getOrCreateError(err)
	structuredErr.Errors = errors
	return structuredErr
}
