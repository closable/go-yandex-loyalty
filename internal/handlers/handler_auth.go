package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	errors_api "github.com/closable/go-yandex-loyalty/internal/errors"
	"github.com/closable/go-yandex-loyalty/internal/utils"
	"github.com/closable/go-yandex-loyalty/models"
	"go.uber.org/zap"
)

func (ah *APIHandler) Orders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//userID := 6
	userID, _ := strconv.Atoi(r.FormValue("userID"))

	if userID == 0 {
		token, _ := r.Cookie("Authorization")
		headerAuth := r.Header.Get("Authorization")

		if len(token.String()) > 0 {
			userID = utils.GetUserID(token.Value)
			w.Header().Add("Authorization", token.Value)
			fmt.Printf("user get from existing cookies %d\n", userID)
		}

		if len(headerAuth) > 0 && userID == 0 {
			userID = utils.GetUserID(headerAuth)
			w.Header().Add("Authorization", headerAuth)
			fmt.Printf("user get from existing header %d\n", userID)
		}

		fmt.Printf("user id %d\n", userID)
	}

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
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("totals of orders - %d", len(orders)))
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

	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("totals of orders - %d", len(body)))
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
	w.Header().Set("Content-Type", "application/json")
	userID, _ := strconv.Atoi(r.FormValue("userID"))

	if userID == 0 {
		token, _ := r.Cookie("Authorization")
		headerAuth := r.Header.Get("Authorization")

		if len(token.String()) > 0 {
			userID = utils.GetUserID(token.Value)
			w.Header().Add("Authorization", token.Value)
		}

		if len(headerAuth) > 0 && userID == 0 {
			userID = utils.GetUserID(headerAuth)
			w.Header().Add("Authorization", headerAuth)
		}
		if userID != 0 {
			ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "Restore authorization balance !")
		}
	}

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
	fmt.Println("еруууунда", withdraw, err)
	resp, err := json.Marshal(withdraw)
	if err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("userID %d balance/withdraw - %f / %f", userID, withdraw.Current, withdraw.Withdrawn))
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
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("error order number %s", orderNumber))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	acc, accStatus := AccrualActions(orderNumber, &ah.sugar, ah.accAddress)
	if accStatus >= 400 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("the accrual system return wrong status %d", accStatus))
		w.WriteHeader(accStatus)
		return
	}

	status := "NEW"
	accrual := 0.0
	// if accrual return the result else default
	if accStatus < 204 {
		status = acc.Status
		accrual = acc.Accrual
	}

	err = ah.db.AddOrder(userID, orderNumber, status, accrual)
	if err != nil {
		httpErr, ok := err.(*errors_api.APIHandlerError)
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		if ok {
			w.WriteHeader(httpErr.Code())
		}
		return
	}
	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("added order %s", orderNumber))
	// w.WriteHeader(http.StatusOK)
	w.WriteHeader(http.StatusAccepted)

}

func (ah *APIHandler) GetWithdraw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, _ := strconv.Atoi(r.FormValue("userID"))

	if userID == 0 {
		token, _ := r.Cookie("Authorization")
		headerAuth := r.Header.Get("Authorization")

		if len(token.String()) > 0 {
			userID = utils.GetUserID(token.Value)
			w.Header().Add("Authorization", token.Value)
		}

		if len(headerAuth) > 0 && userID == 0 {
			userID = utils.GetUserID(headerAuth)
			w.Header().Add("Authorization", headerAuth)
		}
		if userID != 0 {
			ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "Restore authorization getwithdraw !")
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

	// withdraw, err := ah.db.Balance(userID)
	// if err != nil {
	// 	httpErr, ok := err.(*errors_api.APIHandlerError)
	// 	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
	// 	if ok {
	// 		w.WriteHeader(httpErr.Code())
	// 	}
	// 	return
	// }

	// fmt.Println(withdraw.Current, withdraw.Withdrawn, req.Sum)

	// check balance

	// if withdraw.Current-withdraw.Withdrawn < req.Sum {
	// 	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("balance %f/ withdraw %f", withdraw.Current-withdraw.Withdrawn, req.Sum))
	// 	w.WriteHeader(http.StatusPaymentRequired)
	// 	return
	// }

	err = ah.db.AddWithdraw(userID, req.Order, req.Sum)
	if err != nil {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", err)
		httpErr, _ := err.(*errors_api.APIHandlerError)
		w.WriteHeader(httpErr.Code())
		return
	}
	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("userID %d order %s withdrawn - %f", userID, req.Order, req.Sum))
	w.WriteHeader(http.StatusOK)
}

func (ah *APIHandler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, _ := strconv.Atoi(r.FormValue("userID"))

	if userID == 0 {
		token, _ := r.Cookie("Authorization")
		headerAuth := r.Header.Get("Authorization")

		if len(token.String()) > 0 {
			userID = utils.GetUserID(token.Value)
			w.Header().Add("Authorization", token.Value)
		}

		if len(headerAuth) > 0 && userID == 0 {
			userID = utils.GetUserID(headerAuth)
			w.Header().Add("Authorization", headerAuth)
		}
		if userID != 0 {
			ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", "Restore authorization wihdrawals !")
		}
	}

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
	fmt.Println("Что за фигня !!!!", userID, orders, err)
	if len(orders) == 0 {
		ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("no content userID %d", userID))
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
	ah.sugar.Infoln("uri", r.RequestURI, "method", r.Method, "description", fmt.Sprintf("userID %d withdrawals - %d", userID, len(body)))
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

func AccrualActions(orderNumber string, sugar *zap.SugaredLogger, accAddress string) (*models.AccrualGet, int) {

	client := &http.Client{}
	// check order into accrual
	accrual := &models.AccrualGet{}

	sugar.Infoln(fmt.Sprintf("getting info from accrual  %s/api/orders/%s", accAddress, orderNumber))
	accOrder, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/orders/%s", accAddress, orderNumber), nil)
	if err != nil {
		sugar.Infoln("accrual actions: getting order info into the system", err)
		return accrual, http.StatusInternalServerError
	}
	accOrder.Header.Set("Content-Type", "application/json")
	accResp, err := client.Do(accOrder)
	if err != err {
		sugar.Infoln(fmt.Sprintf("accrual actions: invalid %v", err))
		return accrual, http.StatusInternalServerError
	}
	if accResp.StatusCode > 202 {
		sugar.Infoln(fmt.Sprintf("accrual actions: return order %s status %d", orderNumber, accResp.StatusCode))
		return accrual, accResp.StatusCode
	}

	body, err := io.ReadAll(accResp.Body)
	if err != nil {
		sugar.Infoln("accrual actions: read body", err)
		return accrual, http.StatusInternalServerError
	}
	defer accResp.Body.Close()

	if err = json.Unmarshal(body, accrual); err != nil {
		sugar.Infoln("accrual actions: unpack body to json", err)
		return accrual, http.StatusInternalServerError
	}

	return accrual, accResp.StatusCode

}
