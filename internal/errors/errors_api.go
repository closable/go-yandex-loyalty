package errorsApi

import "fmt"

type APIHandlerError struct {
	Err     error
	Message string
	Status  int
}

func (he *APIHandlerError) Error() string {
	return fmt.Sprintf("%v %v", he.Err, he.Message)
}

func (he *APIHandlerError) Code() int {
	return he.Status
}

func NewAPIError(err error, msg string, status int) error {
	return &APIHandlerError{
		Err:     err,
		Message: msg,
		Status:  status,
	}
}
