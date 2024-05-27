package handlers

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"net/url"

	"github.com/closable/go-yandex-loyalty/internal/utils"
	"github.com/go-chi/chi/v5"
)

// Middleware для контроля за аутентифированными пользователями
func (ah *APIHandler) Authenticator(h http.Handler) http.Handler {
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

// Middleware для работы профилировщика pprof
func Profiler() http.Handler {
	r := chi.NewRouter()
	//r.Use(NoCache)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/pprof/", http.StatusMovedPermanently)
	})
	r.HandleFunc("/pprof", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/", http.StatusMovedPermanently)
	})

	r.HandleFunc("/pprof/*", pprof.Index)
	r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/pprof/profile", pprof.Profile)
	r.HandleFunc("/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/pprof/trace", pprof.Trace)
	r.Handle("/vars", expvar.Handler())

	r.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/pprof/mutex", pprof.Handler("mutex"))
	r.Handle("/pprof/heap", pprof.Handler("heap"))
	r.Handle("/pprof/block", pprof.Handler("block"))
	r.Handle("/pprof/allocs", pprof.Handler("allocs"))

	return r
}
