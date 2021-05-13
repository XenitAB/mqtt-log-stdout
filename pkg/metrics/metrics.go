package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

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
	done         chan struct{}
	doneMu       sync.Mutex
	stopping     bool
	stoppingMu   sync.Mutex
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
		done:         make(chan struct{}),
		stopping:     false,
		httpServer:   srv,
		statusClient: opts.StatusClient,
	}
}

// Done returns a channel that is closed if the application is stopped
func (server *Server) Done() <-chan struct{} {
	server.doneMu.Lock()
	if server.done == nil {
		server.done = make(chan struct{})
	}
	d := server.done
	server.doneMu.Unlock()
	return d
}

// Start starts the server
func (server *Server) Start() {
	server.statusClient.Print("Metrics server starting", nil)
	err := server.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		server.statusClient.Print("Metrics server failed to start or stop gracefully", err)
		_ = server.Stop()
	}
}

// Stop stops the server and calls StopWithContext()
func (server *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.StopWithContext(ctx)

	return err
}

// StopWithContext takes a context and stops the server
func (server *Server) StopWithContext(ctx context.Context) error {
	// Check if Metrics server already has been (or is being) stopped
	server.stoppingMu.Lock()
	if server.stopping {
		server.stoppingMu.Unlock()
		return nil
	}

	server.stopping = true
	server.stoppingMu.Unlock()

	// If server.done already has been closed, it would cause an error
	select {
	case <-server.done:
	default: // Channel is not closed, close it
		close(server.done)
	}

	// Shutdown the http server
	err := server.httpServer.Shutdown(ctx)
	if err != nil {
		server.statusClient.Print("Metrics server failed to stop gracefully", err)
		return err
	}

	server.statusClient.Print("Metrics server stopped", nil)

	return nil
}
