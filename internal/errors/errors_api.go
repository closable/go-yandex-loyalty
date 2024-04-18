package errorsapi

import (
	"errors"
	"fmt"
)

var (
	ErrorExecCommit   = errors.New("error during commit")
	ErrorExecQuery    = errors.New("error during executing query")
	ErrorScanQuery    = errors.New("error during scan data query")
	ErrorPrepareQuery = errors.New("error during prepare query")
	ErrorBeginTx      = errors.New("error during start transaction")
	ErrorRegInfo      = errors.New("part of register information is empty")
	ErrorConflict     = errors.New("informaion conflict")
)

type APIHandlerError struct {
	Err     error
	Message string
	//Status  int
}

func (he *APIHandlerError) Error() string {
	return fmt.Sprintf("%v %v", he.Err, he.Message)
}

// func (he *APIHandlerError) Code() int {
// 	return he.Status
// }

func NewAPIError(err error, msg string, status int) error {
	return &APIHandlerError{
		Err:     err,
		Message: msg,
		//Status:  status,
	}
}
