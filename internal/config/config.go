package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr              string
	DatabaseURI          string
	AccrualSystemAddress string
}

func NewConfig() *Config {
	config := &Config{}

	flag.StringVar(&config.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&config.DatabaseURI, "d", "", "database URI")
	flag.StringVar(&config.AccrualSystemAddress, "r", ":8090", "accrual system address")
	flag.Parse()

	if envRunAddr, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		config.RunAddr = envRunAddr
	}

	if envDatabaseURI, ok := os.LookupEnv("DATABASE_URI"); ok {
		config.DatabaseURI = envDatabaseURI
	}

	if envAccrualSystemAddress, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		config.AccrualSystemAddress = envAccrualSystemAddress
	}

	return config
}
