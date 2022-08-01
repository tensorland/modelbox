package client

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/diptanu/modelbox/client-go/proto"
	"github.com/diptanu/modelbox/server/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func MLFrameworkProtoFromStr(framework string) proto.MLFramework {
	switch strings.ToLower(framework) {
	case "pytorch":
		return proto.MLFramework_PYTORCH
	case "keras":
		return proto.MLFramework_KERAS
	}
	return proto.MLFramework_UNKNOWN
}

const (
	DEADLINE = 10 * time.Second
)

type CheckpointDownloadResponse struct {
	Checksum       string
	ServerChecksum string
}

type FileUploadResponse struct {
	Id       string
	Checksum string
}

type CreateModelApiResponse struct {
	Id string
}

type ModelBoxClient struct {
	conn   *grpc.ClientConn
	client proto.ModelStoreClient
}

func NewModelBoxClient(addr string) (*ModelBoxClient, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	client := proto.NewModelStoreClient(conn)
	return &ModelBoxClient{conn: conn, client: client}, nil
}

func (m *ModelBoxClient) CreateExperiment(name string, owner string, namespace string, framework string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DEADLINE)
	defer cancel()
	req := &proto.CreateExperimentRequest{
		Name:      name,
		Owner:     owner,
		Namespace: namespace,
		Framework: MLFrameworkProtoFromStr(framework),
	}
	resp, err := m.client.CreateExperiment(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.ExperimentId, nil
}

func (m *ModelBoxClient) ListExperiments(namespace string) (*proto.ListExperimentsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DEADLINE)
	defer cancel()
	req := &proto.ListExperimentsRequest{Namespace: namespace}
	return m.client.ListExperiments(ctx, req)
}

func (m *ModelBoxClient) ListCheckpoints(experimentId string) (*proto.ListCheckpointsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DEADLINE)
	defer cancel()
	req := &proto.ListCheckpointsRequest{ExperimentId: experimentId}
	return m.client.ListCheckpoints(ctx, req)
}

func (m *ModelBoxClient) CreateModel(name, owner, namespace, task, description string, metadata map[string]string, files []*proto.FileMetadata) (*CreateModelApiResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DEADLINE)
	defer cancel()
	req := &proto.CreateModelRequest{
		Name:        name,
		Owner:       owner,
		Namespace:   namespace,
		Task:        task,
		Description: description,
		Metadata:    metadata,
		Files:       files,
	}

	resp, err := m.client.CreateModel(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unable to create model: %v", err)
	}
	return &CreateModelApiResponse{Id: resp.Id}, nil
}

func (m *ModelBoxClient) ListModels(namespace string) ([]*proto.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DEADLINE)
	defer cancel()
	req := &proto.ListModelsRequest{
		Namespace: namespace,
	}

	resp, err := m.client.ListModels(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Models, nil
}

type ApiCreateCheckpoint struct {
	ExperimentId string
	Epoch        uint64
	Path         string
}

func (m *ModelBoxClient) CreateCheckpoint(chk *ApiCreateCheckpoint) (*proto.CreateCheckpointResponse, error) {
	checkpointRequest := proto.CreateCheckpointRequest{
		ExperimentId: chk.ExperimentId,
		Epoch:        chk.Epoch,
		Files: []*proto.FileMetadata{{
			ParentId: chk.ExperimentId,
			Path:     chk.Path,
		}},
	}
	response, err := m.client.CreateCheckpoint(context.Background(), &checkpointRequest)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (m *ModelBoxClient) UploadFile(path string, parentId string, t storage.FileMIMEType) (*FileUploadResponse, error) {
	// This makes us read the file twice, this could be simplified
	// if we do bidirectional stream and send the
	// checkpoint at the end of the strem to the server to validate the file
	checksum, err := m.getChecksum(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ctx, cancel := context.WithTimeout(context.Background(), DEADLINE)
	defer cancel()
	req := &proto.UploadFileRequest{
		StreamFrame: &proto.UploadFileRequest_Metadata{
			Metadata: &proto.FileMetadata{
				ParentId: parentId,
				FileType: storage.FileTypeToProto(t),
				Checksum: checksum,
			},
		},
	}
	stream, err := m.client.UploadFile(ctx)
	if err != nil {
		return nil, err
	}
	if err := stream.Send(req); err != nil {
		return nil, err
	}

	bytes := make([]byte, 1024)
	for {
		n, e := f.Read(bytes)
		if e != nil && e != io.EOF {
			return nil, err
		}
		req := &proto.UploadFileRequest{
			StreamFrame: &proto.UploadFileRequest_Chunks{Chunks: bytes[:n]},
		}
		err = stream.Send(req)
		if err != nil {
			return nil, err
		}
		if e == io.EOF {
			break
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}
	return &FileUploadResponse{resp.FileId, checksum}, nil
}

func (m *ModelBoxClient) DownloadBlob(id string, path string) (*CheckpointDownloadResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DEADLINE)
	defer cancel()
	req := &proto.DownloadFileRequest{
		FileId: id,
	}
	h := md5.New()
	stream, err := m.client.DownloadFile(ctx, req)
	if err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	resp, err := stream.Recv()
	if err != nil {
		return nil, err
	}
	for {
		chunks, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		bytes := chunks.GetChunks()
		_, err = f.Write(bytes)
		h.Write(bytes)
		if err != nil {
			stream.CloseSend()
			return nil, err
		}
	}
	checksum := fmt.Sprintf("%x", h.Sum(nil))
	serverChecksum := resp.GetMetadata().GetChecksum()
	if checksum != serverChecksum {
		return nil, fmt.Errorf("actual checksum %v, calculated checksum %v", serverChecksum, checksum)
	}
	return &CheckpointDownloadResponse{Checksum: checksum, ServerChecksum: serverChecksum}, nil
}

func (m *ModelBoxClient) getChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("error reading file while calculating checksum: %v", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
