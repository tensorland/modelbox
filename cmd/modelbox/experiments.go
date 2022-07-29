package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	name      string
	owner     string
	framework string
	namespace string
)

var experimentsCmd = &cobra.Command{
	Use:   "experiments",
	Short: "Client subcommand to create or list experiments",
}

var createExperimentsCmd = &cobra.Command{
	Use: "create",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClientUi(ConfigPath)
		if err != nil {
			zap.L().Sugar().Panicf("unable to create client from config: %v ", err)
		}
		if err := client.CreateExperiment(name, owner, namespace, framework); err != nil {
			zap.L().Sugar().Panicf("unable to create experiment: %v ", err)
		}
	},
}

var listExperimentsCmd = &cobra.Command{
	Use: "list",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClientUi(ConfigPath)
		if err != nil {
			zap.L().Sugar().Panicf("unable to create client: %v", err)
		}
		if err := client.ListExperiments(namespace); err != nil {
			zap.L().Sugar().Panicf("unable to list experiments: %v", err)
		}
	},
}

func init() {
	clientCmd.AddCommand(experimentsCmd)
	experimentsCmd.AddCommand(createExperimentsCmd)
	experimentsCmd.AddCommand(listExperimentsCmd)

	createExperimentsCmd.Flags().StringVar(&name, "name", "", "name of the experiment")
	createExperimentsCmd.Flags().StringVar(&owner, "owner", "", "owner of the experiment")
	createExperimentsCmd.Flags().StringVar(&namespace, "namespace", "", "namespace to which the experiment belongs")
	createExperimentsCmd.Flags().StringVar(&framework, "framework", "", "framework used to create the model")

	listExperimentsCmd.Flags().StringVar(&namespace, "namespace", "", "namespace whose experiments are going to be listed")
}
