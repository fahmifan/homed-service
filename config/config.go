package config

import (
	"os"
)

// defaults
const (
	DefaultHTTPort    string = "8080"
	DefaultBoltDBName        = "homed.db"
)

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
