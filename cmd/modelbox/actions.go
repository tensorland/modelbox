package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var modelId string

var actionsCommand = &cobra.Command{
	Use:   "actions",
	Short: "client subcommand to create or list actions",
}

var createAction = &cobra.Command{
	Use: "create",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := NewClientUi(ConfigPath)
		if err != nil {
			zap.L().Sugar().Panicf("unable to create client: %v", err)
		}
	},
}

var listActions = &cobra.Command{
	Use: "list",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := NewClientUi(ConfigPath)
		if err != nil {
			zap.L().Sugar().Panicf("unable to create client: %v", err)
		}
	},
}

func init() {
	clientCmd.AddCommand(actionsCommand)
	actionsCommand.AddCommand(listActions)
	actionsCommand.AddCommand(createAction)
	createAction.Flags().StringVar(&experimentId, "experiment-id", "", "experiment id to which this action is attached to")
	createAction.Flags().StringVar(&modelId, "model-id", "", "model id to which this action is attached to")

	listActions.Flags().StringVar(&experimentId, "experiment-id", "", "experiment id to which this action is attached to")
	listActions.Flags().StringVar(&modelId, "model-id", "", "model id to which this action is attached to")
}
