package errors_api

import "fmt"

type ApiHandlerError struct {
	Err     error
	Message string
	Status  int
}

func (he *ApiHandlerError) Error() string {
	return fmt.Sprintf("%v %v", he.Err, he.Message)
}

func (he *ApiHandlerError) Code() int {
	return he.Status
}

func NewApiError(err error, msg string, status int) error {
	return &ApiHandlerError{
		Err:     err,
		Message: msg,
		Status:  status,
	}
}
