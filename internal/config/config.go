package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	RunAddr              string
	DatabaseURI          string
	AccrualSystemAddress string
	JWTSecret            string
	JWTExp               time.Duration
}

func NewConfig() *Config {
	config := &Config{}

	flag.StringVar(&config.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&config.DatabaseURI, "d", "", "database URI")
	flag.StringVar(&config.AccrualSystemAddress, "r", ":8090", "accrual system address")
	flag.StringVar(&config.JWTSecret, "j", "", "JWT secret")
	flag.DurationVar(&config.JWTExp, "e", time.Hour, "JWT expiration time")
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

	if envJWTSecret, ok := os.LookupEnv("JWT_SECRET"); ok {
		config.JWTSecret = envJWTSecret
	}

	if envJWTExp, ok := os.LookupEnv("JWT_EXP"); ok {
		config.JWTExp, _ = time.ParseDuration(envJWTExp)
	}

	return config
}
