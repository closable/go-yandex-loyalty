package config

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/caarlos0/env/v10"
)

type config struct {
	ServerAddress  string `env:"RUN_ADDRESS"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DSN            string `env:"DATABASE_URI"`
}

var (
	FlagRunAddr     string
	FlagAccrualAddr string
	FlagDSN         string
	configEnv       = config{}
)

func ParseConfigEnv() {
	env.Parse(&configEnv)
	fmt.Println("111", configEnv)
}

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", "localhost:8090", "address and port to run server")
	flag.StringVar(&FlagAccrualAddr, "r", "localhost:8080", "accrual system address and port")
	//flag.StringVar(&FlagDSN, "d", "postgres://postgres:1303@localhost:5432/postgres", "access to DBMS")
	flag.StringVar(&FlagDSN, "d", "", "access to DBMS")
	flag.Parse()
}

func LoadConfig() *config {
	ParseConfigEnv()
	ParseFlags()
	var config = &config{}

	config.ServerAddress = firstValue(&configEnv.ServerAddress, &FlagRunAddr)
	config.AccrualAddress = firstValue(&configEnv.AccrualAddress, &FlagAccrualAddr)
	config.DSN = firstValue(&configEnv.DSN, &FlagDSN)

	acc, _ := url.Parse(config.AccrualAddress)
	if acc.Host == "" {
		config.AccrualAddress = fmt.Sprintf("http://%s", config.AccrualAddress)
	}

	return config
}

func firstValue(valEnv *string, valFlag *string) string {
	if len(*valEnv) > 0 {
		return *valEnv
	}
	return *valFlag
}
