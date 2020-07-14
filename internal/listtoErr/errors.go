package listtoErr

import "fmt"

const (
	Internal     = "InternalError"
	InvalidVar   = "InvalidVariable"
	ListNotFound = "ListNotFound"
)

// ListtoError is the type for error handling within Listto.
type ListtoError struct {
	code          string
	callingMethod string
	message       string
}

// Error returns a readable description of the error.
func (e *ListtoError) Error() string {
	return e.message
}

// SetCallingMethodIfNil for the ListtoError.
func (e *ListtoError) SetCallingMethodIfNil(method string) {
	if e.callingMethod != "" {
		e.callingMethod = method
	}
}

// CallingMethod returns the callingMethod of the ListtoError.
func (e *ListtoError) CallingMethod() string {
	return e.callingMethod
}

// Code returns the error code.
func (e *ListtoError) Code() string {
	return e.code
}

// ConvertError from generic error interface to ListtoError.
func ConvertError(err error) *ListtoError {
	return &ListtoError{
		code:    Internal,
		message: err.Error(),
	}
}

// InvalidEnvvar returns an error when an envvar is not as expected.
func InvalidEnvvar(envvar string) *ListtoError {
	return &ListtoError{
		code:    InvalidVar,
		message: fmt.Sprintf("envvar was invalid: %s", envvar),
	}
}

// ListNotFoundError returns an error if a list couldn't be found.
func ListNotFoundError(list string) *ListtoError {
	return &ListtoError{
		code:    ListNotFound,
		message: fmt.Sprintf("could not find list: %s", list),
	}
}

// LogError prints the error in bot logs.
func (e *ListtoError) LogError() {
	fmt.Println(fmt.Sprintf("%s: %s", e.callingMethod, e.message))
}
