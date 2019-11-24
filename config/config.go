package config

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// defaults
const (
	DefaultHTTPort    string = "8080"
	DefaultBoltDBName        = "homed.db"
	DefaultEnv               = "development"
)

func init() {
	err := godotenv.Load()
	if err == nil {
		log.Info("loading .env file")
	}
}

// Env :nodoc:
func Env() string {
	if val, ok := os.LookupEnv("ENV"); ok {
		return val
	}

	return DefaultEnv
}

// Port :nodoc:
func Port() string {
	if val, ok := os.LookupEnv("PORT"); ok {
		return val
	}

	return DefaultHTTPort
}

// BoltDBName :nodoc:
func BoltDBName() string {
	if val, ok := os.LookupEnv("BOLT_DB_NAME"); ok {
		return val
	}
	return DefaultBoltDBName
}
