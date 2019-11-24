package cmd

import (
	"fmt"
	"os"

	"gitlab.com/homed/homed-service/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/homed/homed-service/restapi"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run homed server",
	Run: func(cmd *cobra.Command, args []string) {
		server := restapi.NewServer()
		server.Run()
	},
}

var rootCmd = &cobra.Command{Use: "homed"}

func init() {
	if config.Env() == "production" {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(log.InfoLevel)
	}
}

// Execute command
func Execute() {
	rootCmd.AddCommand(serverCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
