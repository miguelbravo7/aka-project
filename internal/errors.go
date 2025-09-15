package internal

import "fmt"

// Error represents a custom error with a code and message.
type Error struct {
	Code    string
	Message string
	Err     error
}

// Error returns the string representation of the error.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%s, message=%s, err=%v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%s, message=%s", e.Code, e.Message)
}

// Wrap wraps an existing error with a new error.
func Wrap(err error, newErr *Error) *Error {
	newErr.Err = err
	return newErr
}

// NewError creates a new Error.
func NewError(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Error codes
const (
	ErrorCodeUnknown      = "unknown"
	ErrorCodeNotFound     = "not_found"
	ErrorCodeInternal     = "internal"
	ErrorCodeUnauthorized = "unauthorized"
)
