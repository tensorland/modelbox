package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	client "github.com/diptanu/modelbox/client-go"
	svrConfig "github.com/diptanu/modelbox/server/config"
	"github.com/diptanu/modelbox/server/storage"
	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap"
)

func WriteServerConfigToFile(path string) error {
	data, err := Asset("assets/modelbox_server.toml")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0600)
}

func CreateSchema(configPath string, schema string, logger *zap.Logger) error {
	config, err := svrConfig.NewServerConfig(configPath)
	if err != nil {
		return nil
	}
	storage, err := storage.NewMetadataStorage(config, logger)
	if err != nil {
		return fmt.Errorf("unable to create storage: %v", err)
	}
	return storage.CreateSchema(schema)
}

func WriteClientConfigToFile(path string) error {
	data, err := Asset("assets/modelbox_client.toml")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0600)
}

type ClientUi struct {
	client *client.ModelBoxClient
	table  *tablewriter.Table
}

func NewClientUi(configPath string) (*ClientUi, error) {
	config, err := client.NewClientConfig(configPath)
	if err != nil {
		return nil, err
	}
	client, err := client.NewModelBoxClient(config.ServerAddr)
	if err != nil {
		return nil, err
	}
	return &ClientUi{client, tablewriter.NewWriter(os.Stdout)}, nil
}

func (u *ClientUi) CreateExperiment(name, owner, namespace, framework string) error {
	experimentId, err := u.client.CreateExperiment(name, owner, namespace, framework)
	if err != nil {
		return err
	}
	u.table.SetHeader([]string{"Experiment ID"})
	u.table.Append([]string{experimentId})
	u.table.Render()
	return nil
}

func (u *ClientUi) ListExperiments(namespace string) error {
	resp, err := u.client.ListExperiments(namespace)
	if err != nil {
		return err
	}
	u.table.SetHeader([]string{"Experiment ID", "Namespace", "Owner", "Framework"})
	for _, exp := range resp.Experiments {
		u.table.Append([]string{exp.Id, exp.Namespace, exp.Owner, exp.Framework.String()})
	}
	u.table.Render()
	return nil
}

func (u *ClientUi) CreateModel(name, owner, namespace, task, description string) error {
	resp, err := u.client.CreateModel(name, owner, namespace, task, description, nil, nil)
	if err != nil {
		return err
	}
	u.table.SetHeader([]string{"Model ID"})
	u.table.Append([]string{resp.Id})
	u.table.Render()
	return nil
}

func (u *ClientUi) ListModels(namespace string) error {
	resp, err := u.client.ListModels(namespace)
	if err != nil {
		return err
	}
	u.table.SetHeader([]string{"Model ID", "Name", "Owner", "Namespace", "Task"})
	for _, model := range resp {
		u.table.Append([]string{model.Id, model.Name, model.Owner, model.Namespace, model.Task})
	}
	u.table.Render()
	return nil
}

func (u *ClientUi) UploadCheckpoint(path, experimentId string, epoch uint64, upload bool) error {
	resp, err := u.client.CreateCheckpoint(&client.ApiCreateCheckpoint{ExperimentId: experimentId, Path: path, Epoch: epoch})
	if err != nil {
		println("couldn't create checkpoint: ", err.Error())
		return nil
	}
	if !upload {
		return nil
	}
	uploadResp, err := u.client.UploadFile(path, resp.CheckpointId, storage.CheckpointFile)
	if err != nil {
		println("unable to upload checkpoint: ", err.Error())
		return nil
	}

	u.table.SetHeader([]string{"Checkpoint Id", "Checksum"})
	u.table.Append([]string{resp.CheckpointId, uploadResp.Checksum})
	u.table.Render()
	return nil
}

func (u *ClientUi) ListCheckpoints(experimentId string) error {
	resp, err := u.client.ListCheckpoints(experimentId)
	if err != nil {
		return err
	}
	if len(resp.Checkpoints) == 0 {
		println("no checkpoints are registered for experiment id ", experimentId)
		return nil
	}
	u.table.SetHeader([]string{"Checkpoint Id", "Experiment ID", "Epoch", "Id", "Path", "Metrics"})
	sort.Slice(resp.Checkpoints, func(i, j int) bool {
		return resp.Checkpoints[i].Epoch < resp.Checkpoints[j].Epoch
	})
	for _, chkpoint := range resp.Checkpoints {
		u.table.Append([]string{
			chkpoint.Id,
			chkpoint.ExperimentId,
			strconv.Itoa(int(chkpoint.Epoch)),
			chkpoint.Files[0].Id,
			chkpoint.Files[0].Path,
			u.metricsMapToString(chkpoint.Metrics),
		})
	}
	u.table.Render()
	return nil
}

func (u *ClientUi) DownloadCheckpoint(id string, path string) error {
	resp, err := u.client.DownloadBlob(id, path)
	if err != nil {
		return err
	}
	u.table.SetHeader([]string{"Checkpoint Id", "Path", "Checksum"})
	u.table.Append([]string{id, path, resp.Checksum})
	u.table.Render()

	return nil
}

func (u *ClientUi) metricsMapToString(metrics map[string]float32) string {
	if len(metrics) == 0 {
		return "-"
	}
	var metricsKv []string
	for metric, val := range metrics {
		metricsKv = append(metricsKv, fmt.Sprintf("%s: %f", metric, val))
	}
	return strings.Join(metricsKv, ",")
}
