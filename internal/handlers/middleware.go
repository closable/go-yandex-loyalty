package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/closable/go-yandex-loyalty/internal/utils"
)

func (ah *ApiHandler) Authenticator(h http.Handler) http.Handler {
	auth := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		var userID int
		token, _ := r.Cookie("Authorization")
		headerAuth := r.Header.Get("Authorization")
		//fmt.Printf("-1 %s 2- %s 3- %s ", token, errCookie, headerAuth)

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

		values := url.Values{}
		values.Add("userID", fmt.Sprintf("%d", userID))
		r.PostForm = values

		if userID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			h.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(auth)
}
