package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xenitab/mqtt-log-stdout/pkg/status"
)

// Options takes the input configuration for the metrics server
type Options struct {
	Address      string
	Port         int
	StatusClient status.Client
}

// Server contains the metrics server struct
type Server struct {
	httpServer   *http.Server
	statusClient status.Client
}

// NewServer returns a metrics server
func NewServer(opts Options) *Server {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())
	listenAddress := net.JoinHostPort(opts.Address, fmt.Sprintf("%d", opts.Port))

	srv := &http.Server{
		Addr:    listenAddress,
		Handler: router,
	}

	return &Server{
		httpServer:   srv,
		statusClient: opts.StatusClient,
	}
}

// Start starts the server
func (server *Server) Start(ctx context.Context) error {
	server.statusClient.Print("Metrics server starting", nil)
	err := server.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		server.statusClient.Print("Metrics server failed to start or stop gracefully", err)
		return err
	}
	return nil
}

// StopWithContext takes a context and stops the server
func (server *Server) Stop(ctx context.Context) error {
	// Shutdown the http server
	err := server.httpServer.Shutdown(ctx)
	if err != nil {
		server.statusClient.Print("Metrics server failed to stop gracefully", err)
		return err
	}

	server.statusClient.Print("Metrics server stopped", nil)

	return nil
}
