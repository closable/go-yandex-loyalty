package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/closable/go-yandex-loyalty/internal/db"
	"github.com/closable/go-yandex-loyalty/internal/handlers"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	ServerAddress := "localhost:8080"
	DSN := "postgres://postgres:1303@localhost:5432"
	//var src handlers.Sourcer
	var err error

	src, err := db.NewDB(DSN) // cfg.DSN)
	if err != nil {
		os.Exit(1)
	}

	handler := handlers.New(src)
	fmt.Printf("Store DBMS setup successfuly -> %s\n", DSN)
	fmt.Printf("Running server on -> %s\n", ServerAddress)

	return http.ListenAndServe(ServerAddress, handler.InitRouter())
}
