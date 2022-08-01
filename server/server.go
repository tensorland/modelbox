package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	pb "github.com/diptanu/modelbox/client-go/proto"
	"github.com/diptanu/modelbox/server/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
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
	model.SetFiles(storage.NewFileSetFromProto(model.Id, req.Files))
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
		storage.NewFileSetFromProto(req.Model, req.Files),
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
	checkpoint.SetFiles(storage.NewFileSetFromProto(checkpoint.Id, req.Files))
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
		apiFiles := make([]*pb.FileMetadata, len(p.Files))
		for i, b := range p.Files {
			apiFiles[i] = &pb.FileMetadata{
				Id:        b.Id,
				ParentId:  b.ParentId,
				Path:      b.Path,
				Checksum:  b.Checksum,
				FileType:  pb.FileType(b.Type),
				CreatedAt: timestamppb.New(time.Unix(b.CreatedAt, 0)),
				UpdatedAt: timestamppb.New(time.Unix(b.UpdatedAt, 0)),
			}
		}
		apiCheckpoints[i] = &pb.Checkpoint{
			Id:           p.Id,
			ExperimentId: req.ExperimentId,
			Epoch:        p.Epoch,
			Files:        apiFiles,
			Metrics:      p.Metrics,
			Metadata:     p.Meta,
			CreatedAt:    timestamppb.New(time.Unix(p.CreatedAt, 0)),
			UpdatedAt:    timestamppb.New(time.Unix(p.UpdtedAt, 0)),
		}
	}
	return &pb.ListCheckpointsResponse{Checkpoints: apiCheckpoints}, nil
}

func (s *GrpcServer) UploadFile(stream pb.ModelStore_UploadFileServer) error {
	req, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("unable to receive blob stream %v", err)
	}
	meta := req.GetMetadata()
	if meta == nil {
		return fmt.Errorf("the first message needs to be checkpoint metadata")
	}

	blobInfo := storage.NewFileMetadata(meta.ParentId, "", meta.Checksum, storage.FileTypeFromProto(meta.FileType), 0, 0)
	blobStorage := s.blobStorageBuilder.Build()
	if err := blobStorage.Open(blobInfo, storage.Write); err != nil {
		return err
	}
	defer blobStorage.Close()

	path, err := blobStorage.GetPath()
	if err != nil {
		return fmt.Errorf("unable to update blob info for checkpoint")
	}
	blobInfo.Path = path
	if err := s.metadataStorage.WriteFiles(stream.Context(), storage.FileSet{blobInfo}); err != nil {
		return fmt.Errorf("unable to create blob metadata: %v", err)
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
		n, err := blobStorage.Write(bytes)
		if err != nil {
			return err
		}
		totalBytes += uint64(n)
	}
	stream.SendAndClose(&pb.UploadFileResponse{FileId: blobInfo.Id})
	s.logger.Info(fmt.Sprintf("checkpoint with id: %v tot bytes: %v", blobInfo.Id, totalBytes))
	return nil
}

func (s *GrpcServer) DownloadFile(
	req *pb.DownloadFileRequest,
	stream pb.ModelStore_DownloadFileServer,
) error {
	blobInfo, err := s.metadataStorage.GetFile(stream.Context(), req.FileId)
	if err != nil {
		return fmt.Errorf("unable to retreive blob metadata: %v", err)
	}
	blobStorage := s.blobStorageBuilder.Build()
	if err := blobStorage.Open(blobInfo, storage.Read); err != nil {
		return fmt.Errorf("unable to create blob storage intf: %v", err)
	}
	defer blobStorage.Close()
	blobMeta := pb.DownloadFileResponse{
		StreamFrame: &pb.DownloadFileResponse_Metadata{
			Metadata: &pb.FileMetadata{
				Id:        blobInfo.Id,
				ParentId:  blobInfo.ParentId,
				Checksum:  blobInfo.Checksum,
				Path:      blobInfo.Path,
				CreatedAt: timestamppb.New(time.Unix(blobInfo.CreatedAt, 0)),
				UpdatedAt: timestamppb.New(time.Unix(blobInfo.UpdatedAt, 0)),
			},
		},
	}
	if err := stream.Send(&blobMeta); err != nil {
		return fmt.Errorf("unable to send blob metadata: %v", err)
	}
	buf := make([]byte, 1024)
	totalBytes := 0
	for {
		n, err := blobStorage.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading blob for id %v: %v", req.FileId, err)
		}
		blobChunk := pb.DownloadFileResponse{
			StreamFrame: &pb.DownloadFileResponse_Chunks{
				Chunks: buf[:n],
			},
		}
		if err := stream.Send(&blobChunk); err != nil {
			return fmt.Errorf("unable to stream chunks for id: %v, err: %v", req.FileId, err)
		}
		totalBytes += n
		if err == io.EOF {
			break
		}
	}
	s.logger.Info(
		fmt.Sprintf("checkpoint chunks sent for id: %v tot bytes: %v", req.FileId, totalBytes),
	)
	return nil
}

func (s *GrpcServer) UpdateMetadata(ctx context.Context, req *pb.UpdateMetadataRequest) (*pb.UpdateMetadataResponse, error) {
	metadataList := []*storage.Metadata{}
	for _, meta := range req.Metadata {
		for k, v := range meta.Payload.AsMap() {
			metadataList = append(metadataList, storage.NewMetadata(meta.ParentId, k, v))
		}
	}
	if err := s.metadataStorage.UpdateMetadata(ctx, metadataList); err != nil {
		return nil, err
	}
	updatedAt := timestamppb.New(time.Now())
	resp := &pb.UpdateMetadataResponse{
		UpdatedAt: updatedAt,
	}
	return resp, nil
}

func (s *GrpcServer) ListMetadata(ctx context.Context, req *pb.ListMetadataRequest) (*pb.ListMetadataResponse, error) {
	metadataList, err := s.metadataStorage.ListMetadata(ctx, req.ParentId)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	for _, meta := range metadataList {
		m[meta.Key] = meta.Value
	}
	payload, err := structpb.NewStruct(m)
	if err != nil {
		return nil, fmt.Errorf("unable to create structspb: %v", err)
	}
	return &pb.ListMetadataResponse{Payload: payload}, nil
}

func (s *GrpcServer) GetCheckpoint(ctx context.Context, req *pb.GetCheckpointRequest) (*pb.GetCheckpointResponse, error) {
	id := storage.GetCheckpointID(req.ExperimentId, req.Epoch)
	chk, err := s.metadataStorage.GetCheckpoint(ctx, id)
	if err != nil {
		return nil, err
	}
	return &pb.GetCheckpointResponse{
		Checkpoint: &pb.Checkpoint{
			Id:           chk.Id,
			Epoch:        chk.Epoch,
			ExperimentId: chk.ExperimentId,
			Metrics:      chk.Metrics,
			Metadata:     chk.Meta,
			CreatedAt:    &timestamppb.Timestamp{},
			UpdatedAt:    &timestamppb.Timestamp{},
		},
	}, nil
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
