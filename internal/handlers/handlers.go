// Пакет для работы handlers
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	errorsapi "github.com/closable/go-yandex-loyalty/internal/errors"
	"github.com/closable/go-yandex-loyalty/internal/utils"
	"github.com/closable/go-yandex-loyalty/models"
	"go.uber.org/zap"
)

// Интерфейс для реализации функционала
type Sourcer interface {
	// Валидация данных пользователя
	ValidateRegisterInfo(login, pass string) error
	// Добавление пользователя
	AddUser(login, pass string) error
	// утентификация пользователя
	Login(login, pass string) (int, error)
	// Перечеь заказов пользователя
	GetOrders(userID int) ([]models.OrdersDB, error)
	// Баланс
	Balance(userID int) (float32, float32, error)
	// Добавление заказа
	AddOrder(userID int, orderNumber, accStatus string, accrual float32) error
	// Добавление списания доступных баллов/рублей
	AddWithdraw(userID int, orderNumber string, sum float32) error
	// Перечент всех списаний
	GetWithdrawals(userID int) ([]models.WithdrawGetDB, error)
	// Подготовка первичного состояния системы хранения данных
	PrepareDB() error
}

type (
	// Структура АПИ
	APIHandler struct {
		db         Sourcer
		sugar      zap.SugaredLogger
		accAddress string
	}
	// Запрос регистрации
	RegisterRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	// Номер заказа
	Orders struct {
		Number   string  `json:"number"`
		Status   string  `json:"status"`
		Accrual  float32 `json:"accrual"`
		UploadAt string  `json:"upload_at"`
	}
	// Запрос списания
	WithdrawGet struct {
		Order string  `json:"order"`
		Sum   float32 `json:"sum"`
	}
	// Единица списания баллов
	Withdraw struct {
		Order       string  `json:"order"`
		Sum         float32 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}
)

// Подготовка СУБД и создание экземпляра хранения
func New(src Sourcer, sugar zap.SugaredLogger, accAddress string) (*APIHandler, error) {
	// prepare db
	err := src.PrepareDB()
	if err != nil {
		sugar.Infoln("can't create DB set")
		return &APIHandler{
			db:         src,
			sugar:      sugar,
			accAddress: accAddress,
		}, errors.New("SQL Server problem")
	}

	return &APIHandler{
		db:         src,
		sugar:      sugar,
		accAddress: accAddress,
	}, nil
}

// Регистрация нового пользователя
//
//	@Summary		Register
//	@Description	Register new user
//	@ID RegisterNewUser
//	@Accept		json
//	@Produce		json
//	@Param request body RegisterRequest true "Requst user data"
//	@Success		200		{string}	string			"ok"
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/api/user/register [post]
func (ah *APIHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := &RegisterRequest{}
	if err = json.Unmarshal(body, req); err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		w.WriteHeader(http.StatusInternalServerError) // think about
		return
	}

	if err := ah.db.ValidateRegisterInfo(req.Login, req.Password); err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)

		if errors.Is(errorsapi.ErrorConflict, err) {
			w.WriteHeader(http.StatusConflict)
		} else if errors.Is(errorsapi.ErrorRegInfo, err) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	err = ah.db.AddUser(req.Login, req.Password)
	if err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID, status := LoginAction(w, ah, req.Login, req.Password)
	if status != 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("login error status %d", status))
		w.WriteHeader(status)
		return
	}

	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("register success %s userID %d", req.Login, userID))
	w.WriteHeader(http.StatusOK)
}

// Сервисная функция для перепроверки токена аутентификация
// используется для дополнительной проверки, если вдруг middleware
// по какой-либо причине пропустит корректную обработку
func LoginAction(w http.ResponseWriter, ah *APIHandler, login, pass string) (int, int) {
	userID, err := ah.db.Login(login, pass)
	if err != nil {
		if errors.Is(err, errorsapi.ErrorRegInfo) {
			return 0, http.StatusBadRequest
		}
		return 0, http.StatusInternalServerError
	}

	token, err := utils.BuildJWTString(userID)
	if err != nil {
		return 0, http.StatusInternalServerError
	}

	w.Header().Add("Authorization", token)
	cookie := http.Cookie{
		Name:    "Authorization",
		Expires: time.Now().Add(utils.TokenEXP),
		Value:   token,
	}
	http.SetCookie(w, &cookie)
	return userID, 0
}

// @Summary		Login
// @Description	Login
// @ID Login
// @Accept		json
// @Produce		json
// @Param request body RegisterRequest true "Requst user data"
// @Success		200		{string}	string			"ok"
// @Failure		400		{string}	string	"Bad request"
// @Failure		500		{string}	string	"Internal server error"
// @Router			/api/user/login [post]
func (ah *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "err body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := &RegisterRequest{}
	if err = json.Unmarshal(body, req); err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "err body")
		w.WriteHeader(http.StatusInternalServerError) // think about
		return
	}
	userID, status := LoginAction(w, ah, req.Login, req.Password)
	if status != 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("login error status %d", status))
		w.WriteHeader(status)
		return
	}

	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description userID", userID)

	w.WriteHeader(http.StatusOK)
}
