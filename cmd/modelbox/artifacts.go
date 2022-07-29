package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	id   string
	path string
)

var artifactsCmd = &cobra.Command{
	Use:   "artifacts",
	Short: "Client subcommand to track and manage artifacts",
}

var downloadCmd = &cobra.Command{
	Use: "download",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClientUi(ConfigPath)
		if err != nil {
			zap.L().Sugar().Panicf("unable to create client: %v", err)
		}
		if err := client.DownloadCheckpoint(id, path); err != nil {
			zap.L().Sugar().Panicf("unable to download artifact: %v", err)
		}
	},
}

func init() {
	clientCmd.AddCommand(artifactsCmd)
	artifactsCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVar(&id, "id", "", "id of the artifact to download")
	downloadCmd.Flags().StringVar(&path, "path", "", "path to download the artifact")
}
