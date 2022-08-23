package main

import (
	"github.com/tensorland/modelbox/server"
	"github.com/tensorland/modelbox/server/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the modelbox server",
	Long: `Starts the modelbox server. All server configuration are provided by the
server config. 

modelbox server start --config-path /path/to/config
`,
	Run: func(cmd *cobra.Command, args []string) {
		logger, _ := zap.NewProduction()
		srvConfig, err := config.NewServerConfig(ConfigPath)
		if err != nil {
			logger.Sugar().Panicf("error creating server config from path: %v, err: %v", ConfigPath, err)
		}

		logger.Sugar().Info("starting modelbox server")
		agent, err := server.NewAgent(srvConfig, logger)
		if err != nil {
			logger.Sugar().Panicf("error creating modelbox server agent: %v", err)
		}
		agent.StartAndBlock()
		logger.Info("modelbox exiting")
	},
}

func init() {
	serverCmd.AddCommand(startCmd)
}
