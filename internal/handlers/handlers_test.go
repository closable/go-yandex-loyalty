package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/closable/go-yandex-loyalty/internal/db"
)

func TestApiHandler_Register(t *testing.T) {
	DSN := "postgres://postgres:1303@localhost:5432"
	src, _ := db.NewDB(DSN)
	ah := New(src)
	type wants struct {
		body       string
		statusCode int
	}

	rnd := rand.Intn(1000)
	tests := []struct {
		name  string
		wants wants
	}{
		{
			name: "Register OK",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, rnd),
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Register Conflict 409",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, rnd),
				statusCode: http.StatusConflict,
			},
		},
		{
			name: "Register Body 400",
			wants: wants{
				body:       `{"login": "test", "password": ""}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Register 500",
			wants: wants{
				body:       `{"login": "test", "password": "dddd}`,
				statusCode: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		fmt.Println(tt)
		bodyReader := strings.NewReader(tt.wants.body)
		r := httptest.NewRequest("POST", "/api/user/register", bodyReader)
		w := httptest.NewRecorder()

		ah.Register(w, r)

		if tt.wants.statusCode != w.Code {
			t.Errorf("different status code wants- %v actual- %v", tt.wants.statusCode, w.Code)
		}
	}
}

func TestApiHandler_Login(t *testing.T) {
	DSN := "postgres://postgres:1303@localhost:5432"
	src, _ := db.NewDB(DSN)
	ah := New(src)
	type wants struct {
		body       string
		statusCode int
	}

	rnd := rand.Intn(1000)
	tests := []struct {
		name  string
		wants wants
	}{
		{
			name: "Login OK",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, rnd),
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Login unauthorized  401",
			wants: wants{
				body:       fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, -1, -1),
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Login Body 400",
			wants: wants{
				body:       `{"login": "test", "password": ""}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Login 500",
			wants: wants{
				body:       `{"login": "test", "password": "dddd}`,
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	// create user before login
	bodyReader := strings.NewReader(fmt.Sprintf(`{"login": "test%d", "password": "test%d"}`, rnd, rnd))
	r := httptest.NewRequest("POST", "/api/user/login", bodyReader)
	w := httptest.NewRecorder()
	ah.Register(w, r)

	for _, tt := range tests {
		fmt.Println(tt)
		bodyReader := strings.NewReader(tt.wants.body)
		r := httptest.NewRequest("POST", "/api/user/login", bodyReader)
		w := httptest.NewRecorder()

		ah.Login(w, r)

		if tt.wants.statusCode != w.Code {
			t.Errorf("different status code wants- %v actual- %v", tt.wants.statusCode, w.Code)
		}
	}
}
