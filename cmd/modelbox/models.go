package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	task        string
	description string
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Client subcommand to create or list models",
}

var createModelCmd = &cobra.Command{
	Use: "create",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClientUi(ConfigPath)
		if err != nil {
			zap.L().Sugar().Panicf("unable to create client from config: %v ", err)
		}
		if err := client.CreateModel(name, owner, namespace, task, description); err != nil {
			zap.L().Sugar().Panicf("unable to create model: %v ", err)
		}
	},
}

var listModelsCmd = &cobra.Command{
	Use: "list",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClientUi(ConfigPath)
		if err != nil {
			zap.L().Sugar().Panicf("unable to create client: %v", err)
		}
		if err := client.ListModels(namespace); err != nil {
			zap.L().Sugar().Panicf("unable to list experiments: %v", err)
		}
	},
}

func init() {
	clientCmd.AddCommand(modelsCmd)
	modelsCmd.AddCommand(createModelCmd)
	modelsCmd.AddCommand(listModelsCmd)

	createModelCmd.Flags().StringVar(&name, "name", "", "name of the model")
	createModelCmd.Flags().StringVar(&owner, "owner", "", "owner of the model")
	createModelCmd.Flags().StringVar(&namespace, "namespace", "", "namespace to which the model belongs")
	createModelCmd.Flags().StringVar(&task, "task", "", "task of the model")
	createModelCmd.Flags().StringVar(&description, "description", "", "description of the model")

	listModelsCmd.Flags().StringVar(&namespace, "namespace", "", "namespace whose models are going to be listed")
}
