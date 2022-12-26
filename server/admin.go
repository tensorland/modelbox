package server

import (
	"context"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	pb "github.com/tensorland/modelbox/sdk-go/proto"
	"github.com/tensorland/modelbox/server/scheduler"
	"github.com/tensorland/modelbox/server/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type AdminServer struct {
	srvr      *grpc.Server
	lis       net.Listener
	scheudler *scheduler.ActionScheduler
	storage   storage.MetadataStorage
	logger    *zap.Logger
	pb.UnimplementedModelBoxAdminServer
}

func (a *AdminServer) RegisterAgent(ctx context.Context, req *pb.RegisterAgentRequest) (*pb.RegisterAgentResponse, error) {
	agent := storage.NewAgent(req.AgentName, req.NodeInfo.HostName, req.NodeInfo.IpAddr, req.NodeInfo.Arch, []string{"todo"})
	if err := a.storage.RegisterNode(ctx, agent); err != nil {
		return nil, err
	}
	return &pb.RegisterAgentResponse{
		NodeId: agent.AgentId(),
	}, nil
}

func (a *AdminServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if err := a.storage.Heartbeat(ctx, &storage.Heartbeat{AgentId: req.NodeId, Time: uint64(req.At.AsTime().Unix())}); err != nil {
		return nil, err
	}
	return &pb.HeartbeatResponse{}, nil
}

func (a *AdminServer) GetWork(ctx context.Context, req *pb.GetRunnableActionInstancesRequest) (*pb.GetRunnableActionInstancesResponse, error) {
	instances, err := a.storage.GetActionInstances(ctx, storage.StatusPending)
	if err != nil {
		return nil, err
	}
	runnableInstances := make([]*pb.RunnableAction, len(instances))
	for i, instance := range instances {
		runnableInstances[i] = &pb.RunnableAction{
			Id:       instance.Id,
			ActionId: instance.ActionId,
			Command:  "",
			Params:   map[string]*structpb.Value{},
		}
	}
	return &pb.GetRunnableActionInstancesResponse{Instances: runnableInstances}, nil
}

func (a *AdminServer) UpdateActionStatus(ctx context.Context, req *pb.UpdateActionStatusRequest) (*pb.UpdateActionStatusResponse, error) {
	update := storage.NewActionInstanceUpdate(req.ActionInstanceId, storage.ActionStatus(req.Status), storage.ActionOutcome(req.Outcome), req.OutcomeReason, req.UdpateTime)
	if _, err := a.scheudler.UpdateInstanceStatus(ctx, update); err != nil {
		return nil, err
	}
	return &pb.UpdateActionStatusResponse{}, nil
}

func NewAdminServer(logger *zap.Logger, lis net.Listener, storage storage.MetadataStorage, scheduler *scheduler.ActionScheduler) *AdminServer {
	srvr := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor))
	adminSrvr := &AdminServer{
		srvr:      srvr,
		lis:       lis,
		scheudler: scheduler,
		storage:   storage,
		logger:    logger,
	}
	pb.RegisterModelBoxAdminServer(srvr, adminSrvr)
	return adminSrvr
}

func (a *AdminServer) Start() {
	a.logger.Sugar().Infof("[admin-server] server listening on addr: %v", a.lis.Addr().String())
	if err := a.srvr.Serve(a.lis); err != nil {
		a.logger.Sugar().Fatalf("[admin-server] can't start admin server: %v", err)
	}
}

func (a *AdminServer) Stop() {
	a.srvr.Stop()
}
