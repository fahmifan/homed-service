package main

import (
	log "github.com/sirupsen/logrus"
	"gitlab.com/homed/homed-service/config"
	"gitlab.com/homed/homed-service/restapi"
)

func init() {
	if config.Env() == "production" {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	server := restapi.NewServer()
	server.Run()
}
