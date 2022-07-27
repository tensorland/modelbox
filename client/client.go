package client

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/diptanu/modelbox/proto"
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

type BlobUploadResponse struct {
	Id       string
	Checksum string
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
		return "", nil
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

type ApiCreateCheckpoint struct {
	ExperimentId string
	Epoch        uint64
	Path         string
}

func (m *ModelBoxClient) CreateCheckpoint(chk *ApiCreateCheckpoint) (*proto.CreateCheckpointResponse, error) {
	checkpointRequest := proto.CreateCheckpointRequest{
		ExperimentId: chk.ExperimentId,
		Epoch:        chk.Epoch,
		Blobs: []*proto.BlobMetadata{{
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

func (m *ModelBoxClient) UploadBlob(path string, parentId string, t storage.BlobType) (*BlobUploadResponse, error) {
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
	req := &proto.UploadBlobRequest{
		Blob: &proto.UploadBlobRequest_Metadata{
			Metadata: &proto.BlobMetadata{
				ParentId: parentId,
				BlobType: storage.BlobTypeToProto(t),
				Checksum: checksum,
			},
		},
	}
	stream, err := m.client.UploadBlob(ctx)
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
		req := &proto.UploadBlobRequest{
			Blob: &proto.UploadBlobRequest_Chunks{Chunks: bytes[:n]},
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
	return &BlobUploadResponse{resp.BlobId, checksum}, nil
}

func (m *ModelBoxClient) DownloadBlob(id string, path string) (*CheckpointDownloadResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DEADLINE)
	defer cancel()
	req := &proto.DownloadBlobRequest{
		BlobId: id,
	}
	h := md5.New()
	stream, err := m.client.DownloadBlob(ctx, req)
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
	println(resp.GetMetadata().ParentId)
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
		return nil, fmt.Errorf("actual checksum %v, calculated checksum %v", checksum, serverChecksum)
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
