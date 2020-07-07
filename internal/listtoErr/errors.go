package listtoErr

import "fmt"

// ListtoError is the type for error handling within Listto.
type ListtoError struct {
	callingMethod string
	message string
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

// ConvertError from generic error interface to ListtoError.
func ConvertError(err error) *ListtoError {
	return &ListtoError{
		message: err.Error(),
	}
}

// InvalidEnvvar returns an error when an envvar is not as expected.
func InvalidEnvvar(envvar string) *ListtoError {
	return &ListtoError{
		message: fmt.Sprintf("envvar was invalid: %s", envvar),
	}
}
