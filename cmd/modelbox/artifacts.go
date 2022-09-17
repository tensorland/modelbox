package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	parentId string
	path     string
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
		if err := client.DownloadCheckpoint(parentId, path); err != nil {
			zap.L().Sugar().Panicf("unable to download artifact: %v", err)
		}
	},
}

var uploadCmd = &cobra.Command{
	Use: "upload",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewClientUi(ConfigPath)
		if err != nil {
			zap.L().Sugar().Panicf("unable to create client: %v", err)
		}
		if err := client.UploadArtifact(path, parentId); err != nil {
			zap.L().Sugar().Panicf("unable to upload artifact: %v", err)
		}
	},
}

func init() {
	clientCmd.AddCommand(artifactsCmd)
	artifactsCmd.AddCommand(downloadCmd)
	artifactsCmd.AddCommand(uploadCmd)

	downloadCmd.Flags().StringVar(&parentId, "id", "", "id of the artifact to download")
	downloadCmd.Flags().StringVar(&path, "path", "", "path to download the artifact")

	uploadCmd.Flags().StringVar(&parentId, "parent-id", "", "parent id of the artifact to download")
	uploadCmd.Flags().StringVar(&path, "path", "", "path of the artifact to upload")
}
