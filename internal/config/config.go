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
	config := &Config{
		RunAddr:              ":8080",
		DatabaseURI:          "",
		AccrualSystemAddress: "http://localhost:8090",
		JWTSecret:            "",
		JWTExp:               time.Hour,
	}

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

	flag.StringVar(&config.RunAddr, "a", config.RunAddr, "address and port to run server")
	flag.StringVar(&config.DatabaseURI, "d", config.DatabaseURI, "database URI")
	flag.StringVar(&config.AccrualSystemAddress, "r", config.AccrualSystemAddress, "accrual system address")
	flag.StringVar(&config.JWTSecret, "j", config.JWTSecret, "JWT secret")
	flag.DurationVar(&config.JWTExp, "e", config.JWTExp, "JWT expiration time")
	flag.Parse()

	return config
}
