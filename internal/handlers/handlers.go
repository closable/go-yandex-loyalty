package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	errors_api "github.com/closable/go-yandex-loyalty/internal/errors"
	"github.com/closable/go-yandex-loyalty/internal/utils"
	"github.com/closable/go-yandex-loyalty/models"
	"go.uber.org/zap"
)

type Sourcer interface {
	ValidateRegisterInfo(login, pass string) error
	AddUser(login, pass string) error
	Login(login, pass string) (int, error)
	GetOrders(userID int) ([]models.OrdersDB, error)
	// Balance(userID int) (models.WithdrawDB, error)
	Balance(userID int) (float32, float32, error)
	AddOrder(userID int, orderNumber, accStatus string, accrual float32) error
	AddWithdraw(userID int, orderNumber string, sum float32) error
	GetWithdrawals(userID int) ([]models.WithdrawGetDB, error)
	PrepareDB() error
}

type (
	APIHandler struct {
		db         Sourcer
		sugar      zap.SugaredLogger
		accAddress string
	}
	RegisterRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	Orders struct {
		Number   string  `json:"number"`
		Status   string  `json:"status"`
		Accrual  float32 `json:"accrual"`
		UploadAt string  `json:"upload_at"`
	}
	WithdrawGet struct {
		Order string  `json:"order"`
		Sum   float32 `json:"sum"`
	}
	Withdraw struct {
		Order       string  `json:"order"`
		Sum         float32 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}
)

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
		fmt.Println("Err body")
		w.WriteHeader(http.StatusInternalServerError) // think about
		return
	}

	if err := ah.db.ValidateRegisterInfo(req.Login, req.Password); err != nil {
		httpErr, ok := err.(*errors_api.APIHandlerError)
		if ok {
			ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
			w.WriteHeader(httpErr.Code())
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	err = ah.db.AddUser(req.Login, req.Password)
	if err != nil {
		httpErr, ok := err.(*errors_api.APIHandlerError)
		if ok {
			ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
			w.WriteHeader(httpErr.Code())
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
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

func LoginAction(w http.ResponseWriter, ah *APIHandler, login, pass string) (int, int) {
	userID, err := ah.db.Login(login, pass)
	if err != nil {
		httpErr, _ := err.(*errors_api.APIHandlerError)
		return 0, httpErr.Code()
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
	// userID, err := ah.db.Login(req.Login, req.Password)
	// if err != nil {
	// 	httpErr, ok := err.(*errors_api.APIHandlerError)
	// 	if ok {
	// 		w.WriteHeader(httpErr.Code())
	// 	}
	// 	return
	// }

	// token, err := utils.BuildJWTString(userID)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// w.Header().Add("Authorization", token)
	// cookie := http.Cookie{
	// 	Name:    "Authorization",
	// 	Expires: time.Now().Add(utils.TokenEXP),
	// 	Value:   token,
	// }
	// http.SetCookie(w, &cookie)

	//fmt.Println(userID)
	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description userID", userID)

	w.WriteHeader(http.StatusOK)
}
