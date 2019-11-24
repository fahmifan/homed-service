package main

import (
	"gitlab.com/homed/homed-service/restapi"
)

func main() {
	server := restapi.NewServer()
	server.Run()
}
