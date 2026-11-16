package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BillingPort string
	BillingHost string
	BillingDSN  string
	OrderPort   string
	OrderDSN    string
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	billingDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("BILLING_DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	orderDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("ORDER_DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	return &Config{
		BillingPort: os.Getenv("BILLING_PORT"),
		BillingHost: os.Getenv("BILLING_HOST"),
		BillingDSN:  billingDSN,
		OrderPort:   os.Getenv("ORDER_PORT"),
		OrderDSN:    orderDSN,
	}, nil
}
