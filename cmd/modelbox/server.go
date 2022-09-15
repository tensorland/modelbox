package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var metadataSchemaPath string

var serverInitConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Creates a sample server config",
	Long: `Creates a sample server config By default the config
is created in the current directory. Help:
./modelbox server --init-config path/to/new/condig`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		err := WriteServerConfigToFile(path)
		if err == nil {
			fmt.Printf("config written to path: %v\n", path)
		} else {
			fmt.Printf("unable to write config: %v\n", err)
		}
	},
}

var createSchemaCmd = &cobra.Command{
	Use:   "create-schema --config-path /path/to/config --schema-path",
	Short: "Creates the database schema for modelbox.",
	Long: `Creates the database schema for modelbox. The database has to be
reachable by the client. 
./modebox server create-schema --config-path /path/to/config --schema-path /path/to/schema`,
	Run: func(cmd *cobra.Command, args []string) {
		logger, _ := zap.NewProduction()
		if err := CreateSchema(ConfigPath, metadataSchemaPath, logger); err != nil {
			os.Exit(1)
		}
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the ModelBox service",
	Long: `Start the modelbox server by specifying the server config file.
	
modelbox server --config-path ./path/to/config`,
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverInitConfigCmd)
	serverInitConfigCmd.Flags().String("path", "./modelbox_server.toml", "path to write the server config")

	serverCmd.AddCommand(createSchemaCmd)
	createSchemaCmd.Flags().StringVar(&metadataSchemaPath, "schema-dir", "", "path to metadata schema dir")
}
