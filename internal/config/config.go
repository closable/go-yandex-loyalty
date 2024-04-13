package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

type config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	DSN           string `env:"DATABASE_DSN"`
}

var (
	FlagRunAddr string
	FlagDSN     string
	configEnv   = config{}
)

func ParseConfigEnv() {
	env.Parse(&configEnv)
}

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	//flag.StringVar(&FlagDSN, "d", "postgres://postgres:1303@localhost:5432/postgres", "access to DBMS")
	flag.StringVar(&FlagDSN, "d", "", "access to DBMS")

	flag.Parse()
}

func LoadConfig() *config {
	ParseConfigEnv()
	ParseFlags()

	var config = &config{}
	config.ServerAddress = firstValue(&configEnv.ServerAddress, &FlagRunAddr)
	config.DSN = firstValue(&configEnv.DSN, &FlagDSN)

	return config
}

func firstValue(valEnv *string, valFlag *string) string {
	if len(*valEnv) > 0 {
		return *valEnv
	}
	return *valFlag
}
