package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/closable/go-yandex-loyalty/internal/config"
	"github.com/closable/go-yandex-loyalty/internal/db"
	"github.com/closable/go-yandex-loyalty/internal/utils"
)

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

var dsn string

func initDSN() {
	if len(dsn) > 0 {
		return
	}
	dsn = config.LoadConfig().DSN
}

func TestAPIHandler_AddOrder(t *testing.T) {
	if len(dsn) == 0 {
		initDSN()
	}
	src, _ := db.NewDB(dsn)
	logger := NewLogger()
	sugar := *logger.Sugar()
	ah, _ := New(src, sugar)
	type wants struct {
		body       string
		statusCode int
		method     string
		url        string
		authAction bool
		step       int
	}

	rnd := rand.Intn(1000)
	rndNext := rand.Intn(1000)
	orderTest := utils.SillyGenerateOrderNumberLuhna(8)
	tests := []struct {
		name  string
		wants wants
	}{
		{
			name: "Create new user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, rnd),
				method:     "POST",
				url:        "/api/user/register",
				statusCode: http.StatusOK,
				authAction: false,
				step:       1,
			},
		},
		{
			name: "Login new user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, rnd),
				method:     "POST",
				url:        "/api/user/login",
				statusCode: http.StatusOK,
				authAction: true,
				step:       2,
			},
		},
		{
			name: "Add order to new user",
			wants: wants{
				body:       orderTest, // utils.SillyGenerateOrderNumberLuhna(10), //"12345678903",
				method:     "POST",
				url:        "/api/user/orders",
				statusCode: http.StatusOK,
				authAction: false,
				step:       3,
			},
		},
		{
			name: "Add invalid order to new user",
			wants: wants{
				body:       "123",
				method:     "POST",
				url:        "/api/user/orders",
				statusCode: http.StatusUnprocessableEntity,
				authAction: false,
				step:       3,
			},
		},
		{
			name: "Login not registred user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, 111),
				method:     "POST",
				url:        "/api/user/login",
				statusCode: http.StatusUnauthorized,
				authAction: true,
				step:       2,
			},
		},
		{
			name: "Create new user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rndNext, rndNext),
				method:     "POST",
				url:        "/api/user/register",
				statusCode: http.StatusOK,
				authAction: false,
				step:       1,
			},
		},
		{
			name: "Login next user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rndNext, rndNext),
				method:     "POST",
				url:        "/api/user/login",
				statusCode: http.StatusOK,
				authAction: true,
				step:       2,
			},
		},
		{
			name: "Add order to next user (conflict)",
			wants: wants{
				body:       orderTest, // utils.SillyGenerateOrderNumberLuhna(10), //"12345678903",
				method:     "POST",
				url:        "/api/user/orders",
				statusCode: http.StatusConflict,
				authAction: false,
				step:       3,
			},
		},
		{
			name: "Add order to next user",
			wants: wants{
				body:       utils.SillyGenerateOrderNumberLuhna(10), //"12345678903",
				method:     "POST",
				url:        "/api/user/orders",
				statusCode: http.StatusOK,
				authAction: false,
				step:       3,
			},
		},
	}

	var userID int
	for _, tt := range tests {
		bodyReader := strings.NewReader(tt.wants.body)
		r := httptest.NewRequest(tt.wants.method, tt.wants.url, bodyReader)
		w := httptest.NewRecorder()

		switch tt.wants.step {
		case 1:
			ah.Register(w, r)
		case 2:
			ah.Login(w, r)
		case 3:
			if userID > 0 && tt.wants.step >= 2 {
				values := url.Values{}
				values.Add("userID", fmt.Sprintf("%d", userID))
				r.PostForm = values
			}
			ah.AddOrder(w, r)

		}
		// login && add user id into form
		if tt.wants.authAction && tt.wants.step == 2 && w.Code == http.StatusOK {
			headerAuth := w.Header().Get("Authorization")
			// set user ID from header
			userID = utils.GetUserID(headerAuth)
		}

		if tt.wants.statusCode != w.Code {
			t.Errorf("different status code wants- %v actual- %v", tt.wants.statusCode, w.Code)
		}
	}
}

func TestAPIHandler_Orders(t *testing.T) {
	if len(dsn) == 0 {
		initDSN()
	}
	src, _ := db.NewDB(dsn)
	logger := NewLogger()
	sugar := *logger.Sugar()
	ah, _ := New(src, sugar)
	type wants struct {
		body       string
		statusCode int
		method     string
		url        string
		authAction bool
		step       int
		ordersCnt  int
	}

	rnd := rand.Intn(1000)
	rndNext := rand.Intn(1000)
	tests := []struct {
		name  string
		wants wants
	}{
		{
			name: "Create new user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, rnd),
				method:     "POST",
				url:        "/api/user/register",
				statusCode: http.StatusOK,
				authAction: false,
				step:       1,
				ordersCnt:  0,
			},
		},
		{
			name: "Login new user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, rnd),
				method:     "POST",
				url:        "/api/user/login",
				statusCode: http.StatusOK,
				authAction: true,
				step:       2,
				ordersCnt:  0,
			},
		},
		{
			name: "Add order 1 to user",
			wants: wants{
				body:       utils.SillyGenerateOrderNumberLuhna(10), //"12345678903",
				method:     "POST",
				url:        "/api/user/orders",
				statusCode: http.StatusOK,
				authAction: false,
				step:       3,
				ordersCnt:  0,
			},
		},
		{
			name: "Add order 2 to user",
			wants: wants{
				body:       utils.SillyGenerateOrderNumberLuhna(10), //"12345678903",
				method:     "POST",
				url:        "/api/user/orders",
				statusCode: http.StatusOK,
				authAction: false,
				step:       3,
				ordersCnt:  0,
			},
		},

		{
			name: "Get user orders",
			wants: wants{
				body:       "",
				method:     "GET",
				url:        "/api/user/orders",
				statusCode: http.StatusOK,
				authAction: false,
				step:       4,
				ordersCnt:  2,
			},
		},

		{
			name: "Create new user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rndNext, rndNext),
				method:     "POST",
				url:        "/api/user/register",
				statusCode: http.StatusOK,
				authAction: false,
				step:       1,
				ordersCnt:  0,
			},
		},
		{
			name: "Login new user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rndNext, rndNext),
				method:     "POST",
				url:        "/api/user/login",
				statusCode: http.StatusOK,
				authAction: true,
				step:       2,
				ordersCnt:  0,
			},
		},

		{
			name: "Get user orders empty",
			wants: wants{
				body:       "",
				method:     "GET",
				url:        "/api/user/orders",
				statusCode: http.StatusNoContent,
				authAction: false,
				step:       4,
				ordersCnt:  0,
			},
		},
	}

	var userID int
	orders := []Orders{}

	for _, tt := range tests {

		bodyReader := strings.NewReader(tt.wants.body)
		fmt.Println(tt.wants.body, bodyReader)
		r := httptest.NewRequest(tt.wants.method, tt.wants.url, bodyReader)
		w := httptest.NewRecorder()

		switch tt.wants.step {
		case 1:
			ah.Register(w, r)
		case 2:
			ah.Login(w, r)
		case 3:
			values := url.Values{}
			values.Add("userID", fmt.Sprintf("%d", userID))
			r.PostForm = values
			ah.AddOrder(w, r)
		case 4:
			values := url.Values{}
			values.Add("userID", fmt.Sprintf("%d", userID))
			r.PostForm = values

			ah.Orders(w, r)
			if w.Code == http.StatusOK {
				body, _ := io.ReadAll(w.Body)
				//fmt.Println("orders", string(body))
				if err := json.Unmarshal(body, &orders); err != nil {
					t.Fatal("warning! error unmarshal body")
				}
			}

		}
		// login && add user id into form
		if tt.wants.authAction && tt.wants.step == 2 && w.Code == http.StatusOK {
			headerAuth := w.Header().Get("Authorization")
			// set user ID from header
			userID = utils.GetUserID(headerAuth)
		}

		if tt.wants.statusCode != w.Code {
			t.Errorf("different status code wants- %v actual- %v", tt.wants.statusCode, w.Code)
		}

		if tt.wants.ordersCnt > 0 && tt.wants.ordersCnt != len(orders) {
			t.Errorf("different number of recs wants -%v actual -%v", tt.wants.ordersCnt, len(orders))
		}
	}
}

func TestAPIHandler_Balance(t *testing.T) {
	if len(dsn) == 0 {
		initDSN()
	}
	src, _ := db.NewDB(dsn)
	logger := NewLogger()
	sugar := *logger.Sugar()
	ah, _ := New(src, sugar)
	type wants struct {
		body       string
		statusCode int
		method     string
		url        string
		authAction bool
		step       int
		balance    float64
	}

	rnd := rand.Intn(1000)
	rndNext := rand.Intn(1000)
	tests := []struct {
		name  string
		wants wants
	}{
		{
			name: "Create new user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, rnd),
				method:     "POST",
				url:        "/api/user/register",
				statusCode: http.StatusOK,
				authAction: false,
				step:       1,
				balance:    0,
			},
		},
		{
			name: "Login new user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, rnd),
				method:     "POST",
				url:        "/api/user/login",
				statusCode: http.StatusOK,
				authAction: true,
				step:       2,
				balance:    0,
			},
		},
		{
			name: "Add order 1 to user",
			wants: wants{
				body:       utils.SillyGenerateOrderNumberLuhna(10), //"12345678903",
				method:     "POST",
				url:        "/api/user/orders",
				statusCode: http.StatusOK,
				authAction: false,
				step:       3,
				balance:    0,
			},
		},
		{
			name: "Add order 2 to user",
			wants: wants{
				body:       utils.SillyGenerateOrderNumberLuhna(10), //"12345678903",
				method:     "POST",
				url:        "/api/user/orders",
				statusCode: http.StatusOK,
				authAction: false,
				step:       3,
				balance:    0,
			},
		},

		{
			name: "Get user balance",
			wants: wants{
				body:       "",
				method:     "GET",
				url:        "/api/user/balance",
				statusCode: http.StatusOK,
				authAction: false,
				step:       4,
				balance:    0,
			},
		},

		{
			name: "Create new user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rndNext, rndNext),
				method:     "POST",
				url:        "/api/user/register",
				statusCode: http.StatusOK,
				authAction: false,
				step:       1,
				balance:    0,
			},
		},
		{
			name: "Login new user",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test*%d"}`, rndNext, rndNext),
				method:     "POST",
				url:        "/api/user/login",
				statusCode: http.StatusUnauthorized,
				authAction: true,
				step:       2,
				balance:    0,
			},
		},

		{
			name: "Get user balance empty",
			wants: wants{
				body:       "",
				method:     "GET",
				url:        "/api/user/balance",
				statusCode: http.StatusUnauthorized,
				authAction: false,
				step:       4,
				balance:    0,
			},
		},
	}

	var userID int

	balance := &Balance{}
	for _, tt := range tests {

		bodyReader := strings.NewReader(tt.wants.body)
		fmt.Println(tt.wants.body, bodyReader)
		r := httptest.NewRequest(tt.wants.method, tt.wants.url, bodyReader)
		w := httptest.NewRecorder()

		switch tt.wants.step {
		case 1:
			ah.Register(w, r)
		case 2:
			ah.Login(w, r)
		case 3:
			values := url.Values{}
			values.Add("userID", fmt.Sprintf("%d", userID))
			r.PostForm = values
			ah.AddOrder(w, r)
		case 4:
			values := url.Values{}
			values.Add("userID", fmt.Sprintf("%d", userID))
			r.PostForm = values

			ah.Balance(w, r)
			if w.Code == http.StatusOK {
				body, _ := io.ReadAll(w.Body)
				//fmt.Println("orders", string(body))
				if err := json.Unmarshal(body, &balance); err != nil {
					t.Fatal("warning! error unmarshal body")
				}
			}

		}
		// login && add user id into form
		if tt.wants.authAction && tt.wants.step == 2 {
			if w.Code == http.StatusOK {
				headerAuth := w.Header().Get("Authorization")
				// set user ID from header
				userID = utils.GetUserID(headerAuth)
			} else {
				userID = 0
			}
		}

		if tt.wants.statusCode != w.Code {
			t.Errorf("different status code wants- %v actual- %v", tt.wants.statusCode, w.Code)
		}

	}
}

// func errTestAPIHandler_Withdrawals(t *testing.T) {
// 	if len(dsn) == 0 {
// 		initDSN()
// 	}
// 	src, _ := db.NewDB(dsn)
// 	logger := NewLogger()
// 	sugar := *logger.Sugar()
// 	ah := New(src, sugar)
// 	type wants struct {
// 		body       string
// 		statusCode int
// 		method     string
// 		url        string
// 		authAction bool
// 		step       int
// 		cnt        int
// 	}

// 	//rnd := rand.Intn(1000)
// 	rndNext := rand.Intn(1000)
// 	tests := []struct {
// 		name  string
// 		wants wants
// 	}{
// 		// {
// 		// 	name: "Create new user",
// 		// 	wants: wants{
// 		// 		body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, rnd),
// 		// 		method:     "POST",
// 		// 		url:        "/api/user/register",
// 		// 		statusCode: http.StatusOK,
// 		// 		authAction: false,
// 		// 		step:       1,
// 		// 		cnt:        0,
// 		// 	},
// 		// },
// 		{
// 			name: "Login new user",
// 			wants: wants{
// 				body:       `{"login": "kapa333", "password": "kapa333"}`, //fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, rnd),
// 				method:     "POST",
// 				url:        "/api/user/login",
// 				statusCode: http.StatusOK,
// 				authAction: true,
// 				step:       2,
// 				cnt:        0,
// 			},
// 		},
// 		// {
// 		// 	name: "Add order 1 to user",
// 		// 	wants: wants{
// 		// 		body:       utils.SillyGenerateOrderNumberLuhna(10), //"12345678903",
// 		// 		method:     "POST",
// 		// 		url:        "/api/user/orders",
// 		// 		statusCode: http.StatusOK,
// 		// 		authAction: false,
// 		// 		step:       3,
// 		// 		cnt:        0,
// 		// 	},
// 		// },
// 		{
// 			name: "Get user withdrawals",
// 			wants: wants{
// 				body:       "",
// 				method:     "GET",
// 				url:        "/api/user/withdrawals",
// 				statusCode: http.StatusOK,
// 				authAction: false,
// 				step:       4,
// 				cnt:        5,
// 			},
// 		},

// 		// {
// 		// 	name: "Create new user",
// 		// 	wants: wants{
// 		// 		body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rndNext, rndNext),
// 		// 		method:     "POST",
// 		// 		url:        "/api/user/register",
// 		// 		statusCode: http.StatusOK,
// 		// 		authAction: false,
// 		// 		step:       1,
// 		// 		cnt:        0,
// 		// 	},
// 		// },
// 		{
// 			name: "Login new user",
// 			wants: wants{
// 				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rndNext, rndNext),
// 				method:     "POST",
// 				url:        "/api/user/login",
// 				statusCode: http.StatusUnauthorized,
// 				authAction: true,
// 				step:       2,
// 				cnt:        0,
// 			},
// 		},

// 		{
// 			name: "Get user orders empty",
// 			wants: wants{
// 				body:       "",
// 				method:     "GET",
// 				url:        "/api/user/withdrawals",
// 				statusCode: http.StatusUnauthorized,
// 				authAction: false,
// 				step:       4,
// 				cnt:        0,
// 			},
// 		},
// 	}

// 	var userID int
// 	orders := []Withdraw{}

// 	for _, tt := range tests {

// 		bodyReader := strings.NewReader(tt.wants.body)
// 		fmt.Println(tt.wants.body, bodyReader)
// 		r := httptest.NewRequest(tt.wants.method, tt.wants.url, bodyReader)
// 		w := httptest.NewRecorder()

// 		switch tt.wants.step {
// 		case 1:
// 			ah.Register(w, r)
// 		case 2:
// 			ah.Login(w, r)
// 		case 3:
// 			values := url.Values{}
// 			values.Add("userID", fmt.Sprintf("%d", userID))
// 			r.PostForm = values
// 			ah.AddOrder(w, r)
// 		case 4:
// 			values := url.Values{}
// 			values.Add("userID", fmt.Sprintf("%d", userID))
// 			r.PostForm = values

// 			ah.Withdrawals(w, r)
// 			if w.Code == http.StatusOK {
// 				body, _ := io.ReadAll(w.Body)
// 				//fmt.Println("orders", string(body))
// 				if err := json.Unmarshal(body, &orders); err != nil {
// 					t.Fatal("warning! error unmarshal body")
// 				}
// 			}

// 		}
// 		// login && add user id into form
// 		if tt.wants.authAction && tt.wants.step == 2 {
// 			if w.Code == http.StatusOK {
// 				headerAuth := w.Header().Get("Authorization")
// 				// set user ID from header
// 				userID = utils.GetUserID(headerAuth)
// 			} else {
// 				userID = 0
// 			}
// 		}

// 		if tt.wants.statusCode != w.Code {
// 			t.Errorf("different status code wants- %v actual- %v", tt.wants.statusCode, w.Code)
// 		}

// 		if tt.wants.cnt > 0 && tt.wants.cnt != len(orders) {
// 			t.Errorf("different number of recs wants -%v actual -%v", tt.wants.cnt, len(orders))
// 		}
// 	}
// }
