package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	errors_api "github.com/closable/go-yandex-loyalty/errors"
	"github.com/closable/go-yandex-loyalty/internal/utils"
)

func (ah *ApiHandler) Orders(w http.ResponseWriter, r *http.Request) {

	//userID := 6
	userID, _ := strconv.Atoi(r.FormValue("userID"))

	orders, err := ah.db.GetOrders(userID)
	if err != nil {
		httpErr, ok := err.(*errors_api.ApiHandlerError)
		if ok {
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

func (ah *ApiHandler) Balance(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.FormValue("userID"))

	withdraw, err := ah.db.Balance(userID)
	if err != nil {
		httpErr, ok := err.(*errors_api.ApiHandlerError)
		if ok {
			w.WriteHeader(httpErr.Code())
		}
		return
	}

	resp, err := json.Marshal(withdraw)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}

func (ah *ApiHandler) AddOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	userID, _ := strconv.Atoi(r.FormValue("userID"))
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		fmt.Println("Err body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderNumber := string(body)
	if ok := utils.CheckOrderByLuna(orderNumber); !ok {
		//fmt.Println(ok, orderNumber)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = ah.db.AddOrder(userID, orderNumber)

	httpErr, ok := err.(*errors_api.ApiHandlerError)
	if err != nil {
		if ok {
			w.WriteHeader(httpErr.Code())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ah *ApiHandler) GetWithdraw(w http.ResponseWriter, r *http.Request) {

	userID, _ := strconv.Atoi(r.FormValue("userID"))

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		fmt.Println("Err body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := &WithdrawGet{}
	if err = json.Unmarshal(body, req); err != nil {
		fmt.Println("Err body")
		w.WriteHeader(http.StatusInternalServerError) // think about
		return
	}

	// if ok := utils.CheckOrderByLuna(req.Order); !ok {
	// 	w.WriteHeader(http.StatusUnprocessableEntity)
	// 	return
	// }

	withdraw, err := ah.db.Balance(userID)
	httpErr, ok := err.(*errors_api.ApiHandlerError)
	if err != nil {
		if ok {
			w.WriteHeader(httpErr.Code())
		}
		return
	}

	fmt.Println(withdraw.Current, withdraw.Withdrawn, req.Sum)

	// check balance
	if withdraw.Current-withdraw.Withdrawn < req.Sum {
		w.WriteHeader(http.StatusPaymentRequired)
	}

	err = ah.db.AddWithdraw(userID, req.Order, req.Sum)
	if err != nil {
		httpErr, ok := err.(*errors_api.ApiHandlerError)
		if ok {
			w.WriteHeader(httpErr.Code())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ah *ApiHandler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.FormValue("userID"))

	orders, err := ah.db.GetWithdrawals(userID)
	if err != nil {
		httpErr, ok := err.(*errors_api.ApiHandlerError)
		if ok {
			w.WriteHeader(httpErr.Code())
		}
		return
	}

	if len(orders) == 0 {
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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
