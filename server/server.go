package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb "github.com/tensorland/modelbox/sdk-go/proto"
	"github.com/tensorland/modelbox/server/membership"
	"github.com/tensorland/modelbox/server/storage"
	"github.com/tensorland/modelbox/server/storage/artifacts"
	"github.com/tensorland/modelbox/server/storage/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcServer struct {
	grpcServer         *grpc.Server
	httpServer         *http.Server
	metadataStorage    storage.MetadataStorage
	experimentLogger   logging.ExperimentLogger
	blobStorageBuilder artifacts.BlobStorageBuilder
	clusterMebership   membership.ClusterMembership

	grpcLis net.Listener
	httpLis net.Listener
	logger  *zap.Logger
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
	)
	model.SetFiles(NewFileSetFromProto(req.Files))
	if _, err := s.metadataStorage.CreateModel(ctx, model, getMetadataOrDefault(req.Metadata)); err != nil {
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
		NewFileSetFromProto(req.Files),
		req.UniqueTags,
	)
	if _, err := s.metadataStorage.CreateModelVersion(ctx, modelVersion, getMetadataOrDefault(req.Metadata)); err != nil {
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
			Id:          m.Id,
			Name:        m.Name,
			Owner:       m.Owner,
			Namespace:   m.Namespace,
			Task:        m.Task,
			Description: m.Description,
			Files:       FileSetToProto(m.Files),
		}
	}
	return &pb.ListModelsResponse{Models: apiModels}, nil
}

// Creates a new experiment
func (s *GrpcServer) CreateExperiment(
	ctx context.Context,
	req *pb.CreateExperimentRequest,
) (*pb.CreateExperimentResponse, error) {
	fwk := MLFrameworkFromProto(req.Framework)
	experiment := storage.NewExperiment(
		req.Name,
		req.Owner,
		req.Namespace,
		req.ExternalId,
		fwk,
	)
	result, err := s.metadataStorage.CreateExperiment(ctx, experiment, getMetadataOrDefault(req.Metadata))
	if err != nil {
		return nil, err
	}
	return &pb.CreateExperimentResponse{
		ExperimentId:     result.ExperimentId,
		ExperimentExists: result.Exists,
		CreatedAt:        timestamppb.New(time.Unix(experiment.CreatedAt, 0)),
		UpdatedAt:        timestamppb.New(time.Unix(experiment.UpdatedAt, 0)),
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
			Framework: MLFrameworkToProto(e.Framework),
			CreatedAt: timestamppb.New(time.Unix(e.CreatedAt, 0)),
			UpdatedAt: timestamppb.New(time.Unix(e.UpdatedAt, 0)),
		}
	}
	return &pb.ListExperimentsResponse{Experiments: experimentResponses}, nil
}

func (s *GrpcServer) GetExperiment(ctx context.Context, req *pb.GetExperimentRequest) (*pb.GetExperimentResponse, error) {
	experiment, err := s.metadataStorage.GetExperiment(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetExperimentResponse{Experiment: &pb.Experiment{
		Id:         experiment.Id,
		Name:       experiment.Name,
		Namespace:  experiment.Namespace,
		Owner:      experiment.Owner,
		Framework:  pb.MLFramework(experiment.Framework),
		ExternalId: experiment.ExternalId,
		CreatedAt:  timestamppb.New(time.Unix(experiment.CreatedAt, 0)),
		UpdatedAt:  timestamppb.New(time.Unix(experiment.UpdatedAt, 0)),
	}}, nil
}

func (s *GrpcServer) CreateCheckpoint(
	ctx context.Context,
	req *pb.CreateCheckpointRequest,
) (*pb.CreateCheckpointResponse, error) {
	checkpoint := storage.NewCheckpoint(
		req.ExperimentId,
		req.Epoch,
		req.Metrics,
	)
	checkpoint.SetFiles(NewFileSetFromProto(req.Files))
	if _, err := s.metadataStorage.CreateCheckpoint(ctx, checkpoint, getMetadataOrDefault(req.Metadata)); err != nil {
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
				FileType:  FileTypeToProto(b.Type),
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
			CreatedAt:    timestamppb.New(time.Unix(p.CreatedAt, 0)),
			UpdatedAt:    timestamppb.New(time.Unix(p.UpdtedAt, 0)),
		}
	}
	return &pb.ListCheckpointsResponse{Checkpoints: apiCheckpoints}, nil
}

func (s *GrpcServer) UploadFile(stream pb.ModelStore_UploadFileServer) error {
	req, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("unable to receive file stream %v", err)
	}
	meta := req.GetMetadata()
	if meta == nil {
		return fmt.Errorf("the first message needs to be checkpoint metadata")
	}
	fileMetadata := artifacts.NewFileMetadata(meta.ParentId, "", meta.Checksum, FileTypeFromProto(meta.FileType), 0, 0)
	blobStorage, err := s.blobStorageBuilder.BuildWriter(fileMetadata)
	if err != nil {
		return fmt.Errorf("unable to build artifact storage client: %v", err)
	}
	defer blobStorage.Close()

	path, err := blobStorage.GetPath()
	if err != nil {
		return fmt.Errorf("unable to update file info for checkpoint")
	}
	fileMetadata.Path = path
	if err := s.metadataStorage.WriteFiles(stream.Context(), artifacts.FileSet{fileMetadata}); err != nil {
		// TODO This is not great, we should create a new error type and throw and check on the error type
		// or code.
		if strings.HasPrefix(err.Error(), "duplicate file") {
			stream.SendAndClose(&pb.UploadFileResponse{FileId: fileMetadata.Id})
			return nil
		}
		return status.Errorf(codes.Internal, "unable to create blob metadata: %v", err)
	}
	fileMetadata.Path = path
	var totalBytes uint64 = 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("unable to receive artifact chunks: %v", err)
		}
		bytes := req.GetChunks()
		n, err := blobStorage.Write(bytes)
		if err != nil {
			return fmt.Errorf("unable to write chunks to blob-store: %v", err)
		}
		totalBytes += uint64(n)
	}
	stream.SendAndClose(&pb.UploadFileResponse{FileId: fileMetadata.Id})
	s.logger.Info(fmt.Sprintf("checkpoint with id: %v tot bytes: %v", fileMetadata.Id, totalBytes))
	return nil
}

func (s *GrpcServer) DownloadFile(
	req *pb.DownloadFileRequest,
	stream pb.ModelStore_DownloadFileServer,
) error {
	fileMetadata, err := s.metadataStorage.GetFile(stream.Context(), req.FileId)
	if err != nil {
		return fmt.Errorf("unable to retreive blob metadata: %v", err)
	}
	blobStorage, err := s.blobStorageBuilder.BuildReader(fileMetadata)
	if err != nil {
		return fmt.Errorf("unable to build artifact storage client: %v", err)
	}
	defer blobStorage.Close()
	blobMeta := pb.DownloadFileResponse{
		StreamFrame: &pb.DownloadFileResponse_Metadata{
			Metadata: &pb.FileMetadata{
				Id:        fileMetadata.Id,
				ParentId:  fileMetadata.ParentId,
				Checksum:  fileMetadata.Checksum,
				FileType:  FileTypeToProto(fileMetadata.Type),
				Path:      fileMetadata.Path,
				CreatedAt: timestamppb.New(time.Unix(fileMetadata.CreatedAt, 0)),
				UpdatedAt: timestamppb.New(time.Unix(fileMetadata.UpdatedAt, 0)),
			},
		},
	}
	if err := stream.Send(&blobMeta); err != nil {
		return fmt.Errorf("unable to send file metadata: %v", err)
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

func (s *GrpcServer) TrackArtifacts(ctx context.Context, req *pb.TrackArtifactsRequest) (*pb.TrackArtifactsResponse, error) {
	files := NewFileSetFromProto(req.Files)
	if err := s.metadataStorage.WriteFiles(ctx, files); err != nil {
		if strings.HasPrefix(err.Error(), "unable to create blobs for model: Error 1062:") {
			goto respond
		}
		return nil, err
	}
respond:
	return &pb.TrackArtifactsResponse{
		NumFilesTracked: int32(len(files)),
		CreatedAt:       timestamppb.New(time.Now()),
	}, nil
}

func (s *GrpcServer) ListArtifacts(ctx context.Context, req *pb.ListArtifactsRequest) (*pb.ListArtifactsResponse, error) {
	files, err := s.metadataStorage.GetFiles(ctx, req.ParentId)
	if err != nil {
		return nil, err
	}
	return &pb.ListArtifactsResponse{
		Files: FileSetToProto(files),
	}, nil
}

func (s *GrpcServer) UpdateMetadata(ctx context.Context, req *pb.UpdateMetadataRequest) (*pb.UpdateMetadataResponse, error) {
	if err := s.metadataStorage.UpdateMetadata(ctx, req.ParentId, getMetadataOrDefault(req.Metadata)); err != nil {
		return nil, err
	}
	updatedAt := timestamppb.New(time.Now())
	resp := &pb.UpdateMetadataResponse{
		UpdatedAt: updatedAt,
	}
	return resp, nil
}

func (s *GrpcServer) ListMetadata(ctx context.Context, req *pb.ListMetadataRequest) (*pb.ListMetadataResponse, error) {
	metadata, err := s.metadataStorage.ListMetadata(ctx, req.ParentId)
	if err != nil {
		return nil, err
	}
	return &pb.ListMetadataResponse{Metadata: metadata}, nil
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
			CreatedAt:    &timestamppb.Timestamp{},
			UpdatedAt:    &timestamppb.Timestamp{},
		},
	}, nil
}

func (s *GrpcServer) LogMetrics(ctx context.Context, req *pb.LogMetricsRequest) (*pb.LogMetricsResponse, error) {
	switch v := req.Value.Value.(type) {
	case *pb.MetricsValue_FVal:
		if err := s.experimentLogger.LogFloats(ctx, req.ParentId, req.Key, storage.ToFloatLogFromProto(req.GetValue())); err != nil {
			return nil, fmt.Errorf("unable to log metric: %v", err)
		}
	case nil:
	default:
		return nil, fmt.Errorf("unable to write metric for %v", v)
	}
	return &pb.LogMetricsResponse{}, nil
}

func (s *GrpcServer) GetMetrics(ctx context.Context, req *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	metrics, err := s.experimentLogger.GetFloatLogs(ctx, req.ParentId)
	if err != nil {
		return nil, err
	}
	protoMetrics := make([]*pb.Metrics, 0)
	for key := range metrics {
		values := metrics[key]
		metricValues := make([]*pb.MetricsValue, len(values))
		for i, v := range values {
			metricValues[i] = &pb.MetricsValue{
				Step:          v.Step,
				WallclockTime: v.WallClock,
				Value:         &pb.MetricsValue_FVal{FVal: v.Value},
			}
		}
		m := &pb.Metrics{
			Key:    key,
			Values: metricValues,
		}
		protoMetrics = append(protoMetrics, m)
	}
	return &pb.GetMetricsResponse{Metrics: protoMetrics}, nil
}

func (s *GrpcServer) LogEvent(ctx context.Context, req *pb.LogEventRequest) (*pb.LogEventResponse, error) {
	if req.Event == nil {
		return nil, fmt.Errorf("event can't be nil")
	}
	if req.Event.Source == nil {
		return nil, fmt.Errorf("event source can't be nil")
	}
	event := storage.NewEvent(req.ParentId, req.Event.Source.Name, req.Event.Name, req.Event.WallclockTime.AsTime(), getMetadataOrDefault(req.Event.Metadata))
	return &pb.LogEventResponse{CreatedAt: timestamppb.Now()}, s.metadataStorage.LogEvent(ctx, req.ParentId, event)
}

func (s *GrpcServer) ListEvents(ctx context.Context, req *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	events, err := s.metadataStorage.ListEvents(ctx, req.ParentId)
	if err != nil {
		return nil, err
	}
	apiEvents := make([]*pb.Event, len(events))
	for i, event := range events {
		apiEvents[i] = &pb.Event{
			Name:          event.Name,
			Source:        &pb.EventSource{Name: event.Source},
			WallclockTime: timestamppb.New(time.Unix(int64(event.SourceWallclock), 0)),
			Metadata:      &pb.Metadata{Metadata: event.Metadata},
		}
	}
	return &pb.ListEventsResponse{Events: apiEvents}, nil
}

func (s *GrpcServer) CreateActions(ctx context.Context, req *pb.CreateActionRequest) (*pb.CreateActionResponse, error) {
	if err := s.metadataStorage.CreateAction(ctx, storage.NewAction(req.Name, req.Arch, req.ObjectId, req.Trigger.Predicate, req.Params)); err != nil {
		return nil, err
	}
	return &pb.CreateActionResponse{CreatedAt: timestamppb.New(time.Now())}, nil
}

func (s *GrpcServer) ListActions(ctx context.Context, req *pb.ListActionsRequest) (*pb.ListActionsResponse, error) {
	actions, err := s.metadataStorage.ListActions(ctx, req.ObjectId)
	if err != nil {
		return nil, err
	}
	actionResp := make([]*pb.Action, len(actions))
	for i, action := range actions {
		actionResp[i] = &pb.Action{
			Id:   action.Id,
			Name: action.Name,
		}
	}
	return &pb.ListActionsResponse{Actions: actionResp}, nil
}

func (s *GrpcServer) GetClusterMembers(ctx context.Context, req *pb.GetClusterMembersRequest) (*pb.GetClusterMembersResponse, error) {
	members, err := s.clusterMebership.GetMembers()
	if err != nil {
		return nil, err
	}
	clusterMembers := []*pb.ClusterMember{}
	for _, member := range members {
		clusterMembers = append(clusterMembers, &pb.ClusterMember{
			Id:       member.Id,
			HostName: member.HostName,
			HttpAddr: member.HTTPAddr,
			RpcAddr:  member.RPCAddr,
		})
	}
	return &pb.GetClusterMembersResponse{
		Members: clusterMembers,
	}, nil
}

func (s *GrpcServer) WatchNamespace(
	req *pb.WatchNamespaceRequest,
	stream pb.ModelStore_WatchNamespaceServer,
) error {
	since := time.Now()
	var pushTicker <-chan time.Time
	pushTicker = time.After(0)
	for {
		select {
		case <-pushTicker:
			changes, err := s.metadataStorage.ListChanges(stream.Context(), req.Namespace, since)
			since = time.Now()
			if err != nil {
				return fmt.Errorf("unable to list changes: %v", err)
			}
			for _, change := range changes {
				val, err := structpb.NewValue(change.Payload)
				if err != nil {
					return fmt.Errorf("unable to create proto value: %v", err)
				}
				resp := &pb.WatchNamespaceResponse{
					Event:   pb.ServerEvent_OBJECT_CREATED,
					Payload: val,
				}
				stream.Send(resp)
			}
			pushTicker = time.After(5 * time.Second)
		case <-stream.Context().Done():
			return nil
		}
	}
}

func NewGrpcServer(
	metadatStorage storage.MetadataStorage,
	blobStorageBuilder artifacts.BlobStorageBuilder,
	experimentLogger logging.ExperimentLogger,
	grpcLis net.Listener,
	httpLis net.Listener,
	clusterMebership membership.ClusterMembership,
	logger *zap.Logger,
) *GrpcServer {
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	modelBoxServer := &GrpcServer{
		metadataStorage:    metadatStorage,
		experimentLogger:   experimentLogger,
		grpcServer:         grpcServer,
		blobStorageBuilder: blobStorageBuilder,
		grpcLis:            grpcLis,
		httpLis:            httpLis,
		clusterMebership:   clusterMebership,
		logger:             logger,
	}
	pb.RegisterModelStoreServer(grpcServer, modelBoxServer)

	// setup grpc-web
	wrappedGrpc := grpcweb.WrapServer(grpcServer, grpcweb.WithOriginFunc(func(origin string) bool {
		// TODO Harden this and only allow certain origins in production.
		return true
	}))
	router := chi.NewRouter()
	router.Use(
		chiMiddleware.Logger,
		chiMiddleware.Recoverer,
		NewGrpcWebMiddleware(wrappedGrpc).Handler,
	)
	httpServer := &http.Server{Handler: router}
	modelBoxServer.httpServer = httpServer

	return modelBoxServer
}

func (s *GrpcServer) Start() {
	go s.startGrpcWeb()
	s.logger.Sugar().Infof("server listening on addr: %v", s.grpcLis.Addr().String())
	if err := s.grpcServer.Serve(s.grpcLis); err != nil {
		s.logger.Sugar().Fatalf("can't start grpc server: %v", err)
	}
}

func (s *GrpcServer) startGrpcWeb() {
	s.logger.Sugar().Infof("grpc-web listening on addr: %v", s.httpLis.Addr().String())
	if err := s.grpcServer.Serve(s.httpLis); err != nil {
		s.logger.Sugar().Fatalf("unable to start grpc-web: %v", err)
	}
}

func (s *GrpcServer) Stop() {
	s.httpServer.Close()
	s.grpcServer.Stop()
}

func (s *GrpcServer) getWebWrapper() *grpcweb.WrappedGrpcServer {
	wrappedGrpc := grpcweb.WrapServer(s.grpcServer, grpcweb.WithOriginFunc(func(origin string) bool {
		// TODO Harden this and only allow certain origins in production.
		return true
	}))
	return wrappedGrpc
}
