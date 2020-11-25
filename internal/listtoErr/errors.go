package listtoErr

import "fmt"

const (
	Internal     = "InternalError"
	InvalidVar   = "InvalidVariable"
	ListNotFound = "ListNotFound"
)

// ListtoError is the type for error handling within Listto.
type ListtoError struct {
	Code          string
	CallingMethod string
	Message       string
}

// Error returns a readable description of the error.
func (e *ListtoError) Error() string {
	return e.Message
}

// SetCallingMethodIfNil for the ListtoError.
func (e *ListtoError) SetCallingMethodIfNil(method string) {
	if e.CallingMethod != "" {
		e.CallingMethod = method
	}
}

// ConvertError from generic error interface to ListtoError.
func ConvertError(err error) *ListtoError {
	return &ListtoError{
		Code:    Internal,
		Message: err.Error(),
	}
}

// InvalidEnvvar returns an error when an envvar is not as expected.
func InvalidEnvvar(envvar string) *ListtoError {
	return &ListtoError{
		Code:    InvalidVar,
		Message: fmt.Sprintf("envvar was invalid: %s", envvar),
	}
}

// ListNotFoundError returns an error if a list couldn't be found.
func ListNotFoundError(list string) *ListtoError {
	return &ListtoError{
		Code:    ListNotFound,
		Message: fmt.Sprintf("could not find list: %s", list),
	}
}

// LogError prints the error in bot logs.
func (e *ListtoError) LogError() {
	fmt.Println(fmt.Sprintf("%s: %s", e.CallingMethod, e.Message))
}
