package handlers

import (
	"github.com/go-chi/chi/v5"
)

func (ah *APIHandler) InitRouter() chi.Router {
	router := chi.NewRouter()

	router.Post("/api/user/register", ah.Register)
	router.Post("/api/user/login", ah.Login)

	router.Group(func(r chi.Router) {
		r.Use(ah.Authenticator)
		r.Get("/api/user/orders", ah.Orders)
		r.Post("/api/user/orders", ah.AddOrder)
		r.Post("/api/user/balance/withdraw", ah.GetWithdraw)
		r.Get("/api/user/withdrawals", ah.Withdrawals)
		r.Get("/api/user/balance", ah.Balance)
	})

	return router
}
