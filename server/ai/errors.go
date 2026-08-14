package ai

import "errors"

type ErrorType string

const (
	ErrorTypeDisabled   ErrorType = "disabled"
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypePolicy     ErrorType = "policy"
	ErrorTypeQuota      ErrorType = "quota"
	ErrorTypeProvider   ErrorType = "provider"
	ErrorTypeTimeout    ErrorType = "timeout"
	ErrorTypeSchema     ErrorType = "schema"
)

type Error struct {
	Type    ErrorType
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return e.Message + ": " + e.Cause.Error()
}

func (e *Error) Unwrap() error { return e.Cause }

func ErrorKind(err error) ErrorType {
	var gatewayError *Error
	if errors.As(err, &gatewayError) {
		return gatewayError.Type
	}
	return ErrorTypeProvider
}
