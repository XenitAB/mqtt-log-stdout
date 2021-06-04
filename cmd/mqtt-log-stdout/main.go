package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xenitab/mqtt-log-stdout/pkg/config"
	"github.com/xenitab/mqtt-log-stdout/pkg/message"
	"github.com/xenitab/mqtt-log-stdout/pkg/metrics"
	"github.com/xenitab/mqtt-log-stdout/pkg/mqtt"
	"github.com/xenitab/mqtt-log-stdout/pkg/status"
	"golang.org/x/sync/errgroup"
)

var (
	// Version is set at build time to print the released version using --version
	Version = "v0.0.0-dev"
	// Revision is set at build time to print the release git commit sha using --version
	Revision = ""
	// Created is set at build time to print the timestamp for when it was built using --version
	Created = ""
)

func main() {
	cfg, err := newConfigClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to generate config: %q\n", err)
		os.Exit(1)
	}

	err = run(cfg)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func run(cfg config.Client) error {
	g, ctx, cancel := newErrGroup()
	defer cancel()

	stopChan := newStopChannel()
	defer signal.Stop(stopChan)

	statusClient := newStatusClient(cfg)
	messageClient := newMessageClient()
	metricsServer := newMetricsServer(cfg, statusClient)
	mqttClient := newMqttClient(cfg, statusClient, messageClient)

	start(ctx, g, metricsServer)
	start(ctx, g, mqttClient)

	stoppedBy := waitForStop(stopChan, ctx)
	statusClient.Print(fmt.Sprintf("Application stopping, initiated by: %s", stoppedBy), nil)

	cancel()

	timeoutCtx, timeoutCancel := newTimeoutContext()
	defer timeoutCancel()

	stop(timeoutCtx, g, mqttClient)
	stop(timeoutCtx, g, metricsServer)

	return waitForErrGroup(g)
}

func newErrGroup() (*errgroup.Group, context.Context, context.CancelFunc) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)
	return g, ctx, cancel
}

func waitForErrGroup(g *errgroup.Group) error {
	err := g.Wait()
	if err != nil {
		return fmt.Errorf("error groups error: %w", err)
	}

	return nil
}

func newTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
}

func waitForStop(stopChan chan os.Signal, ctx context.Context) string {
	select {
	case sig := <-stopChan:
		return fmt.Sprintf("os.Signal (%s)", sig)
	case <-ctx.Done():
		return "context"
	}
}

type starter interface {
	Start(ctx context.Context) error
}

type stopper interface {
	Stop(ctx context.Context) error
}

func start(ctx context.Context, g *errgroup.Group, s starter) {
	g.Go(func() error {
		return s.Start(ctx)
	})
}

func stop(ctx context.Context, g *errgroup.Group, s stopper) {
	g.Go(func() error {
		return s.Stop(ctx)
	})
}

func newConfigClient() (config.Client, error) {
	opts := config.Options{
		Version:  Version,
		Revision: Revision,
		Created:  Created,
	}

	return config.NewClient(opts)
}

func newStatusClient(cfg config.Client) status.Client {
	opts := status.Options{
		ClientID: cfg.ClientID,
	}

	return status.NewClient(opts)
}

func newMessageClient() message.Client {
	opts := message.Options{}

	return message.NewClient(opts)
}

func newMetricsServer(cfg config.Client, statusClient status.Client) *metrics.Server {
	opts := metrics.Options{
		Address:      cfg.MetricsAddress,
		Port:         cfg.MetricsPort,
		StatusClient: statusClient,
	}

	return metrics.NewServer(opts)
}

func newMqttClient(cfg config.Client, statusClient status.Client, messageClient message.Client) *mqtt.Client {
	opts := mqtt.Options{
		BrokerAddresses: cfg.BrokerAddresses,
		Topic:           cfg.Topic,
		QoS:             cfg.QoS,
		ClientID:        cfg.ClientID,
		Username:        cfg.Username,
		Password:        cfg.Password,
		CleanSession:    cfg.CleanSession,
		KeepAlive:       cfg.KeepAlive,
		ConnectTimeout:  cfg.ConnectTimeout,
		StatusClient:    statusClient,
		MessageClient:   messageClient,
	}

	return mqtt.NewClient(opts)
}

func newStopChannel() chan os.Signal {
	stopChan := make(chan os.Signal, 2)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)
	return stopChan
}
