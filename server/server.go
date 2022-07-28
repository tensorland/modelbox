package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	pb "github.com/diptanu/modelbox/proto"
	"github.com/diptanu/modelbox/server/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcServer struct {
	grpcServer         *grpc.Server
	metadataStorage    storage.MetadataStorage
	blobStorageBuilder storage.BlobStorageBuilder

	lis    net.Listener
	logger *zap.Logger
	pb.UnimplementedModelStoreServer
}

// Create a new Model under a namespace. If no projects are specified, models
// are created under a default namespace.
func (s *GrpcServer) CreateModel(
	ctx context.Context,
	req *pb.CreateModelRequest,
) (*pb.CreateModelResponse, error) {
	model := storage.NewModel(
		req.Name,
		req.Owner,
		req.Namespace,
		req.Task,
		req.Description,
		req.Metadata,
	)
	model.SetBlobs(storage.NewBlobSetFromProto(model.Id, req.Blobs))
	if _, err := s.metadataStorage.CreateModel(ctx, model); err != nil {
		return nil, err
	}
	return &pb.CreateModelResponse{Id: model.Id}, nil
}

func (s *GrpcServer) CreateModelVersion(
	ctx context.Context,
	req *pb.CreateModelVersionRequest,
) (*pb.CreateModelVersionResponse, error) {
	modelVersion := storage.NewModelVersion(
		req.Name,
		req.Model,
		req.Version,
		req.Description,
		storage.MLFramework(req.Framework),
		req.Metadata,
		storage.NewBlobSetFromProto(req.Model, req.Blobs),
		req.UniqueTags,
	)
	if _, err := s.metadataStorage.CreateModelVersion(ctx, modelVersion); err != nil {
		return nil, fmt.Errorf("unable to create model version: %v", err)
	}
	return &pb.CreateModelVersionResponse{ModelVersion: modelVersion.Id}, nil
}

func (s *GrpcServer) ListModelVersions(
	_ context.Context,
	req *pb.ListModelVersionsRequest,
) (*pb.ListModelVersionsResponse, error) {
	return nil, nil
}

// List Models in a given namespace
func (s *GrpcServer) ListModels(ctx context.Context,
	req *pb.ListModelsRequest,
) (*pb.ListModelsResponse, error) {
	models, err := s.metadataStorage.ListModels(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	apiModels := make([]*pb.Model, len(models))
	for i, m := range models {
		apiModels[i] = &pb.Model{
			Id:        m.Id,
			Name:      m.Name,
			Owner:     m.Owner,
			Namespace: m.Namespace,
			Task:      m.Task,
		}
	}
	return &pb.ListModelsResponse{Models: apiModels}, nil
}

// Creates a new experiment
func (s *GrpcServer) CreateExperiment(
	ctx context.Context,
	req *pb.CreateExperimentRequest,
) (*pb.CreateExperimentResponse, error) {
	fwk := storage.MLFrameworkFromProto(req.Framework)
	experiment := storage.NewExperiment(
		req.Name,
		req.Owner,
		req.Namespace,
		req.ExternalId,
		fwk,
		req.Metadata,
	)
	result, err := s.metadataStorage.CreateExperiment(ctx, experiment)
	if err != nil {
		return nil, err
	}
	return &pb.CreateExperimentResponse{
		ExperimentId:     result.ExperimentId,
		ExperimentExists: result.Exists,
	}, nil
}

// List Experiments
func (s *GrpcServer) ListExperiments(
	ctx context.Context,
	req *pb.ListExperimentsRequest,
) (*pb.ListExperimentsResponse, error) {
	experiments, err := s.metadataStorage.ListExperiments(ctx, req.Namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get experiments from storage: %v", err)
	}
	experimentResponses := make([]*pb.Experiment, len(experiments))
	for i, e := range experiments {
		experimentResponses[i] = &pb.Experiment{
			Id:        e.Id,
			Name:      e.Name,
			Owner:     e.Owner,
			Namespace: e.Namespace,
			Framework: storage.MLFrameworkToProto(e.Framework),
			CreatedAt: timestamppb.New(time.Unix(e.CreatedAt, 0)),
			UpdatedAt: timestamppb.New(time.Unix(e.UpdatedAt, 0)),
		}
	}
	return &pb.ListExperimentsResponse{Experiments: experimentResponses}, nil
}

func (s *GrpcServer) CreateCheckpoint(
	ctx context.Context,
	req *pb.CreateCheckpointRequest,
) (*pb.CreateCheckpointResponse, error) {
	checkpoint := storage.NewCheckpoint(
		req.ExperimentId,
		req.Epoch,
		req.Metadata,
		req.Metrics,
	)
	checkpoint.SetBlobs(storage.NewBlobSetFromProto(checkpoint.Id, req.Blobs))
	if _, err := s.metadataStorage.CreateCheckpoint(ctx, checkpoint); err != nil {
		return nil, fmt.Errorf("unable to create checkpoint: %v", err)
	}
	return &pb.CreateCheckpointResponse{CheckpointId: checkpoint.Id}, nil
}

// Lists all the checkpoints for an experiment
func (s *GrpcServer) ListCheckpoints(
	ctx context.Context,
	req *pb.ListCheckpointsRequest,
) (*pb.ListCheckpointsResponse, error) {
	checkpoints, err := s.metadataStorage.ListCheckpoints(ctx, req.ExperimentId)
	if err != nil {
		return nil, fmt.Errorf("unable to get checkpoints from storage: %v", err)
	}
	apiCheckpoints := make([]*pb.Checkpoint, len(checkpoints))
	for i, p := range checkpoints {
		apiBlobs := make([]*pb.BlobMetadata, len(p.Blobs))
		for i, b := range p.Blobs {
			apiBlobs[i] = &pb.BlobMetadata{
				ParentId:  b.ParentId,
				Path:      b.Path,
				Checksum:  b.Checksum,
				BlobType:  pb.BlobType(b.Type),
				CreatedAt: timestamppb.New(time.Unix(b.CreatedAt, 0)),
				UpdatedAt: timestamppb.New(time.Unix(b.UpdatedAt, 0)),
			}
		}
		apiCheckpoints[i] = &pb.Checkpoint{
			Id:           p.Id,
			ExperimentId: req.ExperimentId,
			Epoch:        p.Epoch,
			Blobs:        apiBlobs,
			Metrics:      p.Metrics,
			Metadata:     p.Meta,
			CreatedAt:    timestamppb.New(time.Unix(p.CreatedAt, 0)),
			UpdatedAt:    timestamppb.New(time.Unix(p.UpdtedAt, 0)),
		}
	}
	return &pb.ListCheckpointsResponse{Checkpoints: apiCheckpoints}, nil
}

func (s *GrpcServer) UploadBlob(stream pb.ModelStore_UploadBlobServer) error {
	req, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("unable to receive blob stream %v", err)
	}
	meta := req.GetMetadata()
	if meta == nil {
		return fmt.Errorf("the first message needs to be checkpoint metadata")
	}

	blobInfo := &storage.BlobInfo{
		ParentId: meta.ParentId,
		Type:     storage.BlobTypeFromProto(meta.GetBlobType()),
		Checksum: meta.Checksum,
	}
	blobInfo.CreateId()

	blob := s.blobStorageBuilder.Build()
	if err := blob.Open(blobInfo.Id, storage.Write); err != nil {
		return err
	}
	defer blob.Close()

	path, err := blob.GetPath()
	if err != nil {
		return fmt.Errorf("unable to update blob info for checkpoint")
	}
	if err := s.metadataStorage.UpdateBlobPath(stream.Context(), path, blobInfo.ParentId, blobInfo.Type); err != nil {
		return fmt.Errorf("unable to update blob path: %v", err)
	}
	var totalBytes uint64 = 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		bytes := req.GetChunks()
		n, err := blob.Write(bytes)
		if err != nil {
			return err
		}
		totalBytes += uint64(n)
	}
	stream.SendAndClose(&pb.UploadBlobResponse{BlobId: blobInfo.Id})
	s.logger.Info(fmt.Sprintf("checkpoint with id: %v tot bytes: %v", blobInfo.Id, totalBytes))
	return nil
}

func (s *GrpcServer) DownloadBlob(
	req *pb.DownloadBlobRequest,
	stream pb.ModelStore_DownloadBlobServer,
) error {
	blob := s.blobStorageBuilder.Build()
	if err := blob.Open(req.BlobId, storage.Read); err != nil {
		return err
	}
	defer blob.Close()
	blobMeta := pb.DownloadBlobResponse{
		Blob: &pb.DownloadBlobResponse_Metadata{
			Metadata: &pb.BlobMetadata{},
		},
	}
	if err := stream.Send(&blobMeta); err != nil {
		return fmt.Errorf("unable to send blob metadata: %v", err)
	}
	buf := make([]byte, 1024)
	totalBytes := 0
	for {
		n, err := blob.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading blob for id %v: %v", req.BlobId, err)
		}
		blobChunk := pb.DownloadBlobResponse{
			Blob: &pb.DownloadBlobResponse_Chunks{
				Chunks: buf[:n],
			},
		}
		if err := stream.Send(&blobChunk); err != nil {
			return fmt.Errorf("unable to stream chunks for id: %v, err: %v", req.BlobId, err)
		}
		totalBytes += n
		if err == io.EOF {
			break
		}
	}
	s.logger.Info(
		fmt.Sprintf("checkpoint chunks sent for id: %v tot bytes: %v", req.BlobId, totalBytes),
	)
	return nil
}

func NewGrpcServer(
	metadatStorage storage.MetadataStorage,
	blobStorageBuilder storage.BlobStorageBuilder,
	lis net.Listener,
	logger *zap.Logger,
) *GrpcServer {
	grpcServer := grpc.NewServer()
	modelBoxServer := &GrpcServer{
		metadataStorage:    metadatStorage,
		grpcServer:         grpcServer,
		blobStorageBuilder: blobStorageBuilder,
		lis:                lis,
		logger:             logger,
	}
	pb.RegisterModelStoreServer(grpcServer, modelBoxServer)
	return modelBoxServer
}

func (s *GrpcServer) Start() {
	s.logger.Sugar().Infof("server listening on addr: %v", s.lis.Addr().String())
	if err := s.grpcServer.Serve(s.lis); err != nil {
		s.logger.Fatal(fmt.Sprintf("can't start grpc server: %v", err))
	}
}

func (s *GrpcServer) Stop() {
	s.grpcServer.Stop()
}
