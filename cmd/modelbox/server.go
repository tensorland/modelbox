package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

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
	Use:   "crate-schema --config-path /path/to/config --schema-path",
	Short: "Creates the database schema for modelbox.",
	Long: `Creates the database schema for modelbox. The database has to be
reachable by the client. 
./modebox server create-schema --config-path /path/to/config --schema-path /path/to/schema`,
	Run: func(cmd *cobra.Command, args []string) {
		logger, _ := zap.NewProduction()
		configPath, _ := cmd.PersistentFlags().GetString("config-path")
		schemaPath, _ := cmd.PersistentFlags().GetString("schema-path")
		logger.Sugar().Infof("creating the schema: %v", schemaPath)
		CreateSchema(configPath, schemaPath, logger)
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
	createSchemaCmd.Flags().String("schema-path", "", "Path to the schema file")
}
