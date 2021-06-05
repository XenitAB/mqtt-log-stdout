package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/xenitab/mqtt-log-stdout/pkg/config"
	h "github.com/xenitab/mqtt-log-stdout/pkg/helper"
	"github.com/xenitab/mqtt-log-stdout/pkg/message"
	"github.com/xenitab/mqtt-log-stdout/pkg/metrics"
	"github.com/xenitab/mqtt-log-stdout/pkg/mqtt"
	"github.com/xenitab/mqtt-log-stdout/pkg/status"
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
	cfg, err := newConfigClient(Version, Revision, Created)
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
	errGroup, ctx, cancel := h.NewErrGroupAndContext()
	defer cancel()

	stopChan := h.NewStopChannel()
	defer signal.Stop(stopChan)

	statusClient := newStatusClient(cfg)
	messageClient := newMessageClient()
	metricsServer := newMetricsServer(cfg, statusClient)
	mqttClient := newMqttClient(cfg, statusClient, messageClient)

	h.StartService(ctx, errGroup, metricsServer)
	h.StartService(ctx, errGroup, mqttClient)

	stoppedBy := h.WaitForStop(stopChan, ctx)
	statusClient.Print(fmt.Sprintf("Application stopping, initiated by: %s", stoppedBy), nil)

	cancel()

	timeoutCtx, timeoutCancel := h.NewShutdownTimeoutContext()
	defer timeoutCancel()

	h.StopService(timeoutCtx, errGroup, mqttClient)
	h.StopService(timeoutCtx, errGroup, metricsServer)

	return h.WaitForErrGroup(errGroup)
}

func newConfigClient(version, revision, created string) (config.Client, error) {
	opts := config.Options{
		Version:  version,
		Revision: revision,
		Created:  created,
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
