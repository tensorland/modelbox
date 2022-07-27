package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/diptanu/modelbox/server/config"
	"github.com/diptanu/modelbox/server/storage"
	"go.uber.org/zap"
)

type Agent struct {
	grpcServer *GrpcServer
	storage    storage.MetadataStorage
	ShutdownCh <-chan struct{}

	logger *zap.Logger
}

func NewAgent(config *config.ServerConfig, logger *zap.Logger) (*Agent, error) {
	lis, err := net.Listen("tcp", config.ListenAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to listen on interface: %v", err)
	}
	metadataStorage, err := storage.NewMetadataStorage(config, logger)
	if err != nil {
		return nil, err
	}
	fileStorageBuilder, err := storage.NewBlobStorageBuilder(config, logger)
	if err != nil {
		logger.Fatal(fmt.Sprintf("couldn't build basedire %v", err))
	}
	logger.Sugar().Infof("using storage backend: %v", metadataStorage.Backend())
	server := NewGrpcServer(metadataStorage, fileStorageBuilder, lis, logger)
	return &Agent{grpcServer: server, storage: metadataStorage, logger: logger}, nil
}

func (a *Agent) StartAndBlock() (int, error) {
	go a.grpcServer.Start()
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

	// Stop grpc server
	a.grpcServer.Stop()

	// stop the storage
	if err := a.storage.Close(); err != nil {
		a.logger.Error(fmt.Sprintf("error closing storage: %v", err))
	}
	return 0
}
