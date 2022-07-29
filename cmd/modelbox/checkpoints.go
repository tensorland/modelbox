package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	experimentId   string
	checkpointPath string
	epoch          uint64
	upload         bool
)

var checkpointsCmd = &cobra.Command{
	Use:   "checkpoints",
	Short: "Client subcommand to create, download or checkpoints for an experiment",
}

var createCheckpointsCmd = &cobra.Command{
	Use: "create",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClientUi(ConfigPath)
		if err != nil {
			zap.L().Sugar().Panicf("unable to create client: %v", err)
		}
		if err := client.UploadCheckpoint(checkpointPath, experimentId, epoch, upload); err != nil {
			zap.L().Sugar().Panicf("unable to upload checkpoint: &v", err)
		}
	},
}

var listCheckpointsCmd = &cobra.Command{
	Use: "list",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClientUi(ConfigPath)
		if err != nil {
			zap.L().Sugar().Panicf("unable to create client: %v", err)
		}
		if err := client.ListCheckpoints(experimentId); err != nil {
			zap.L().Sugar().Panicf("unable to list checkpoint: &v", err)
		}
	},
}

func init() {
	clientCmd.AddCommand(checkpointsCmd)
	checkpointsCmd.AddCommand(createCheckpointsCmd)
	checkpointsCmd.AddCommand(listCheckpointsCmd)

	createCheckpointsCmd.Flags().StringVar(&experimentId, "experiment-id", "", "experiment id to which this checkpoint belongs.")
	createCheckpointsCmd.Flags().StringVar(&checkpointPath, "path", "", "path to the checkpoint")
	createCheckpointsCmd.Flags().Uint64Var(&epoch, "epoch", 0, "epoch of the checkpoint")
	createCheckpointsCmd.Flags().BoolVar(&upload, "upload", false, "upload the checkpoint")

	listCheckpointsCmd.Flags().StringVar(&experimentId, "experiment-id", "", "list checkpoints of this experiment id")
}
