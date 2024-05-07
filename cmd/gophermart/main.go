package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/closable/go-yandex-loyalty/internal/backgrounds"
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

	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				// start sync orders
				sugar.Infoln("Execute background process sync orders with accruals", t)
				orders, err := src.NotProcessedOrders()
				if err != nil {
					fmt.Println(orders, err)
				}
				if len(orders) > 0 {
					backgrounds.SyncAccruals(src, cfg.AccrualAddress, &sugar, orders...)
				}
			}
		}
	}()

	sugar.Infoln("Setup DBMS successfuly ->", cfg.DSN)
	sugar.Infoln("Accrual system address ->", cfg.AccrualAddress)
	sugar.Infoln("Running server on ->", cfg.ServerAddress)

	return http.ListenAndServe(cfg.ServerAddress, handler.InitRouter())
}
