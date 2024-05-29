// Пакет консолидации описаний ошибок приложения
package errorsapi

import (
	"errors"
	"fmt"
)

var (
	// Ошибка выполнения сохранения данных
	ErrorExecCommit = errors.New("error during commit")
	// Ошибка выполнения SQL запроса
	ErrorExecQuery = errors.New("error during executing query")
	// Ошибка выборки элементов запроса
	ErrorScanQuery = errors.New("error during scan data query")
	// Ошибка подготовки запроса
	ErrorPrepareQuery = errors.New("error during prepare query")
	// Ошибка транзакции
	ErrorBeginTx = errors.New("error during start transaction")
	// Ошибка не заполненной или частично заполннной информацц
	ErrorRegInfo = errors.New("part of register information is empty")
	// Ошибка (конфликт) дулирующая информаця
	ErrorConflict = errors.New("informaion conflict")
	// Ошибка, информация уже существует
	ErrorInfoFound = errors.New("informaion already present (it's not error)")
)

type APIHandlerError struct {
	Err     error
	Message string
	//Status  int
}

// Deprecated: Информация удалена из дальнейшей поддержки
func (he *APIHandlerError) Error() string {
	return fmt.Sprintf("%v %v", he.Err, he.Message)
}

// func (he *APIHandlerError) Code() int {
// 	return he.Status
// }

// Deprecated: Информация удалена из дальнейшей поддержки
func NewAPIError(err error, msg string, status int) error {
	return &APIHandlerError{
		Err:     err,
		Message: msg,
		//Status:  status,
	}
}
