package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/tensorland/modelbox/server/config"
	"github.com/tensorland/modelbox/server/membership"
	"github.com/tensorland/modelbox/server/scheduler"
	"github.com/tensorland/modelbox/server/storage"
	"github.com/tensorland/modelbox/server/storage/artifacts"
	"github.com/tensorland/modelbox/server/storage/logging"
	"go.uber.org/zap"
)

type Agent struct {
	grpcServer        *GrpcServer
	adminServer       *AdminServer
	promServer        *PromServer
	storage           storage.MetadataStorage
	scheduler         *scheduler.ActionScheduler
	clusterMembership membership.ClusterMembership
	ShutdownCh        <-chan struct{}

	logger *zap.Logger
}

func NewAgent(config *config.ServerConfig, logger *zap.Logger) (*Agent, error) {
	grpcLis, err := net.Listen("tcp", config.GrpcListenAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to listen grpc on interface: %v", err)
	}
	httpLis, err := net.Listen("tcp", config.HttpListenAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to listen grpc-web on interface: %v", err)
	}
	adminLis, err := net.Listen("tcp", config.AdminListenAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to listen to admin server on interface: %v", err)
	}
	pSrvr, err := NewPromServer(config, logger)
	if err != nil {
		return nil, err
	}
	metadataStorage, err := storage.NewMetadataStorage(config, logger)
	if err != nil {
		return nil, err
	}
	fileStorageBuilder, err := artifacts.NewBlobStorageBuilder(config, logger)
	if err != nil {
		logger.Fatal(fmt.Sprintf("couldn't build artifact storage driver: %v", err))
	}
	logger.Sugar().Infof("using metadata backend: %v, artifacts backend: %v ", metadataStorage.Backend(), fileStorageBuilder.Backend())
	experimentLogger, err := logging.NewExperimentLogger(config, logger)
	if err != nil {
		return nil, err
	}
	logger.Sugar().Infof("using metrics backend: %v", experimentLogger.Backend())

	clusterMembership, err := membership.NewClusterMembership(config, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create cluster membership: %v", err)
	}
	logger.Sugar().Infof("cluster membership backend: %v", clusterMembership.Backend())
	server := NewGrpcServer(metadataStorage, fileStorageBuilder, experimentLogger, grpcLis, httpLis, clusterMembership, logger)

	scheduler := scheduler.NewActionScheduler(metadataStorage, config.SchedulerTickDuration, logger)

	adminSrvr := NewAdminServer(logger, adminLis, metadataStorage, scheduler)
	return &Agent{
		grpcServer:        server,
		adminServer:       adminSrvr,
		storage:           metadataStorage,
		scheduler:         scheduler,
		clusterMembership: clusterMembership,
		logger:            logger,
		promServer:        pSrvr,
	}, nil
}

func (a *Agent) StartAndBlock() (int, error) {
	a.clusterMembership.Join()
	go a.promServer.Start()
	go a.grpcServer.Start()
	go a.adminServer.Start()
	a.scheduler.Start()
	return a.handleSignals(), nil
}

func (a *Agent) handleSignals() int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGPIPE)

	// Wait for a signal
WAIT:
	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case <-a.ShutdownCh:
		sig = os.Interrupt
	}

	// Skip any SIGPIPE signal and don't try to log it (See issues #1798, #3554)
	if sig == syscall.SIGPIPE {
		goto WAIT
	}
	// Check if this is a SIGHUP
	if sig == syscall.SIGHUP {
		// TODO Handle reloading config
		goto WAIT
	}

	// Stop the scheduler
	a.scheduler.Stop()

	// leave the cluster
	a.clusterMembership.Leave()

	a.promServer.Stop()

	// Stop grpc server
	a.grpcServer.Stop()

	// stop admin server
	a.adminServer.Stop()

	// stop the storage
	if err := a.storage.Close(); err != nil {
		a.logger.Error(fmt.Sprintf("error closing storage: %v", err))
	}
	return 0
}
