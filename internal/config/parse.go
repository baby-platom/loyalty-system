package config

import (
	"flag"
	"os"
)

// Config includes variables parsed from flags
var Config struct {
	Address             string
	DatabaseURI         string
	AccrualSystemAdress string
	LogLevel            string
	AuthSecretKey       string
}

// ParseFlags parses flags and envs into the Config
func ParseFlags() {
	flag.StringVar(&Config.Address, "a", ":8080", "address and port to run server")
	flag.StringVar(&Config.DatabaseURI, "d", "", "database connection uri")
	flag.StringVar(&Config.AccrualSystemAdress, "r", "", "accrual system address")
	flag.StringVar(&Config.LogLevel, "l", "debug", "log level")
	flag.Parse()

	if envAddress := os.Getenv("RUN_ADDRESS"); envAddress != "" {
		Config.Address = envAddress
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		Config.DatabaseURI = envDatabaseURI
	}
	if envAccrualSystemAdress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAdress != "" {
		Config.AccrualSystemAdress = envAccrualSystemAdress
	}

	Config.AuthSecretKey = "unsecureSecretKey"
	if authSecretKey := os.Getenv("AUTH_SECRET_KEY"); authSecretKey != "" {
		Config.AuthSecretKey = authSecretKey
	}
}
