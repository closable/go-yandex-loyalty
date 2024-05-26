// Пакет предназначен для конфигурирования приложения, используя флани командной строки
// или системные переменные
package config

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/caarlos0/env/v10"
)

type config struct {
	// Адрес сервера приложения
	ServerAddress string `env:"RUN_ADDRESS"`
	// Адрес cервиса accrual
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	// DSN для подключения л PostgreSQL
	DSN string `env:"DATABASE_URI"`
}

var (
	FlagRunAddr     string
	FlagAccrualAddr string
	FlagDSN         string
	configEnv       = config{}
)

// Функция предназначена для обработки параметров, становленных в системных переменных
func ParseConfigEnv() {
	env.Parse(&configEnv)
}

// Функция  предназначена для рабора флагов командной строки
func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", "localhost:8090", "address and port to run server")
	flag.StringVar(&FlagAccrualAddr, "r", "localhost8080", "accrual system address and port")
	flag.StringVar(&FlagDSN, "d", "postgres://postgres:1303@localhost:5432/postgres", "access to DBMS")
	//flag.StringVar(&FlagDSN, "d", "", "access to DBMS")
	flag.Parse()
}

// Функция предназначена для обработки переменных среды окржения и установления рабочих параметров
// в зависимости от переданных параметров или принятых по умолчанию
func LoadConfig() *config {
	ParseConfigEnv()
	ParseFlags()
	var config = &config{}

	config.ServerAddress = FirstValue(&configEnv.ServerAddress, &FlagRunAddr)
	config.AccrualAddress = FirstValue(&configEnv.AccrualAddress, &FlagAccrualAddr)
	config.DSN = FirstValue(&configEnv.DSN, &FlagDSN)

	acc, _ := url.Parse(config.AccrualAddress)
	if acc.Host == "" {
		config.AccrualAddress = fmt.Sprintf("http://%s", config.AccrualAddress)
	}

	return config
}

// Функция триггер, обрабатывает входящие значения в порядке указанном в ТЗ
func FirstValue(valEnv *string, valFlag *string) string {
	if len(*valEnv) > 0 {
		return *valEnv
	}
	return *valFlag
}
