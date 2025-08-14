package errors

import "fmt"

// AppError is an application-specific error type
type AppError struct {
	Code    string
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// creates a new AppError
func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// wraps an error with a code and message
func Wrap(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// Error code constants
const (
	CodeInternal   = "INTERNAL_ERROR"
	CodeNotFound   = "NOT_FOUND"
	CodeInvalidArg = "INVALID_ARGUMENT"
	CodeExternal   = "EXTERNAL_ERROR"
	CodeConflict   = "CONFLICT"         // Resource already exists (UNIQUE violation)
	CodeDependency = "DEPENDENCY_ERROR" // Foreign key constraint violation
)
