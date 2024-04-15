package main

import (
	"net/http"
	"os"

	"github.com/closable/go-yandex-loyalty/internal/config"
	"github.com/closable/go-yandex-loyalty/internal/db"
	"github.com/closable/go-yandex-loyalty/internal/handlers"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	cfg := config.LoadConfig()
	logger := handlers.NewLogger()
	sugar := *logger.Sugar()

	//var src handlers.Sourcer
	var err error

	src, err := db.NewDB(cfg.DSN) // cfg.DSN)
	if err != nil {
		sugar.Infoln(err)
		os.Exit(1)
	}

	handler, err := handlers.New(src, sugar, cfg.AccrualAddress)
	if err != nil {
		sugar.Infoln(err)
		src.DB.Close()
		os.Exit(1)
	}

	sugar.Infoln("Setup DBMS successfuly ->", cfg.DSN)
	sugar.Infoln("Accrual system address ->", cfg.AccrualAddress)
	sugar.Infoln("Running server on ->", cfg.ServerAddress)

	return http.ListenAndServe(cfg.ServerAddress, handler.InitRouter())
}
