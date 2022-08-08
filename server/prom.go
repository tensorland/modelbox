package server

import (
	"net"
	"net/http"

	"github.com/diptanu/modelbox/server/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type PromServer struct {
	lis  net.Listener
	srvr http.Server

	logger *zap.Logger
}

func NewPromServer(config *config.ServerConfig, logger *zap.Logger) (*PromServer, error) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srvr := http.Server{
		Handler: mux,
	}
	logger.Sugar().Infof("prom: listening on addr: %v", config.PromAddr)
	lis, err := net.Listen("tcp", config.PromAddr)
	if err != nil {
		return nil, err
	}
	http.Handle("/metrics", promhttp.Handler())
	return &PromServer{lis: lis, srvr: srvr, logger: logger}, nil
}

func (p *PromServer) Start() error {
	return p.srvr.Serve(p.lis)
}

func (p *PromServer) Stop() {
	p.lis.Close()
}
