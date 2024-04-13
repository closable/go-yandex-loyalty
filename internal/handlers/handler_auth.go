package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	errors_api "github.com/closable/go-yandex-loyalty/internal/errors"
	"github.com/closable/go-yandex-loyalty/internal/utils"
)

func (ah *APIHandler) Orders(w http.ResponseWriter, r *http.Request) {

	//userID := 6
	userID, _ := strconv.Atoi(r.FormValue("userID"))
	if userID == 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "user unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders, err := ah.db.GetOrders(userID)
	if err != nil {
		httpErr, ok := err.(*errors_api.APIHandlerError)
		if ok {
			ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
			w.WriteHeader(httpErr.Code())
		}
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body := make([]Orders, 0)
	for _, v := range orders {
		row := makeOrderItem(v.OrderNumber, v.Status, v.Accrual, v.UploadAt)
		body = append(body, row)
	}

	resp, err := json.Marshal(body)
	if err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("tatals orders - %d", len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}

func makeOrderItem(ordNumb, status string, accrual float64, uploadAt string) Orders {
	var res = &Orders{
		Number:   ordNumb,
		Status:   status,
		Accrual:  accrual,
		UploadAt: uploadAt,
	}
	return *res
}

func (ah *APIHandler) Balance(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.FormValue("userID"))
	if userID == 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "user unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	withdraw, err := ah.db.Balance(userID)
	if err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		httpErr, ok := err.(*errors_api.APIHandlerError)
		if ok {
			w.WriteHeader(httpErr.Code())
		}
		return
	}

	resp, err := json.Marshal(withdraw)
	if err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("balance/withdraw - %f / %f", withdraw.Current, withdraw.Withdrawn))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}

func (ah *APIHandler) AddOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	userID, _ := strconv.Atoi(r.FormValue("userID"))

	// check cookies I'm totaly disagree becaues headers set into middleware
	if userID == 0 {
		token, _ := r.Cookie("Authorization")
		headerAuth := r.Header.Get("Authorization")

		if len(token.String()) > 0 {
			userID = utils.GetUserID(token.Value)
			w.Header().Add("Authorization", token.Value)
			//fmt.Printf("user get from existing cookies %d\n", userID)
		}

		if len(headerAuth) > 0 && userID == 0 {
			userID = utils.GetUserID(headerAuth)
			w.Header().Add("Authorization", headerAuth)
			//fmt.Printf("user get from existing header %d\n", userID)
		}
	}

	if userID == 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "user unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderNumber := string(body)
	if ok := utils.CheckOrderByLuna(orderNumber); !ok {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "error order number")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = ah.db.AddOrder(userID, orderNumber)
	if err != nil {
		httpErr, ok := err.(*errors_api.APIHandlerError)
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		if ok {
			w.WriteHeader(httpErr.Code())
		}
		return
	}
	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("added order %s", orderNumber))
	w.WriteHeader(http.StatusOK)
}

func (ah *APIHandler) GetWithdraw(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.FormValue("userID"))
	if userID == 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "user unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := &WithdrawGet{}
	if err = json.Unmarshal(body, req); err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		w.WriteHeader(http.StatusInternalServerError) // think about
		return
	}

	if ok := utils.CheckOrderByLuna(req.Order); !ok {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "error order number")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	withdraw, err := ah.db.Balance(userID)
	if err != nil {
		httpErr, ok := err.(*errors_api.APIHandlerError)
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		if ok {
			w.WriteHeader(httpErr.Code())
		}
		return
	}

	// fmt.Println(withdraw.Current, withdraw.Withdrawn, req.Sum)

	// check balance
	if withdraw.Current-withdraw.Withdrawn < req.Sum {
		w.WriteHeader(http.StatusPaymentRequired)
	}

	err = ah.db.AddWithdraw(userID, req.Order, req.Sum)
	if err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		httpErr, ok := err.(*errors_api.APIHandlerError)
		if ok {
			w.WriteHeader(httpErr.Code())
		}
		return
	}
	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("withdrawn - %f", withdraw.Withdrawn))
	w.WriteHeader(http.StatusOK)
}

func (ah *APIHandler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.FormValue("userID"))
	if userID == 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "user unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders, err := ah.db.GetWithdrawals(userID)
	if err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		httpErr, ok := err.(*errors_api.APIHandlerError)
		if ok {
			w.WriteHeader(httpErr.Code())
		}
		return
	}

	if len(orders) == 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "no content")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body := make([]Withdraw, 0)
	for _, v := range orders {
		row := makeWithdrawItem(v.Order, v.Sum, v.ProcessedAt)
		body = append(body, row)
	}

	resp, err := json.Marshal(body)
	if err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("withdrawals - %d", len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}

func makeWithdrawItem(ordNumb string, sum float64, processedAt string) Withdraw {
	var res = &Withdraw{
		Order:       ordNumb,
		Sum:         sum,
		ProcessedAt: processedAt,
	}
	return *res
}
