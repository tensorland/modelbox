package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	ConfigPath string
)

var rootCmd = &cobra.Command{
	Use:   "modelbox",
	Short: "cli for the modelbox service",
	Long: `modelbox cli is used to interact with the modelbox server.
The cli is run in either server or client mode. Please refer to the server or client
cli help messages.

modelbox server --help
modelbox client --help
	`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config-path", "", "path to server/client config")
	rootCmd.MarkFlagRequired("config-path")
}
