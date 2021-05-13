package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/xenitab/mqtt-log-stdout/pkg/config"
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
	cfg, err := newConfigClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to generate config: %q\n", err)
		os.Exit(1)
	}

	err = start(cfg)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func start(cfg config.Client) error {
	stopChan := newStopChannel()
	defer signal.Stop(stopChan)

	statusClient := newStatusClient(cfg)
	messageClient := newMessageClient()
	metricsServer := newMetricsServer(cfg, statusClient)
	mqttClient := newMqttClient(cfg, statusClient, messageClient)

	go metricsServer.Start()

	var result error
	err := mqttClient.Start()
	if err != nil {
		statusClient.Print("Received error starting mqtt client", err)
		result = multierror.Append(result, err)
	}

	stoppedBy := func() string {
		select {
		case sig := <-stopChan:
			return fmt.Sprintf("os.Signal (%s)", sig)
		case <-mqttClient.Done():
			return "mqtt client"
		case <-metricsServer.Done():
			return "metrics server"
		}
	}()

	statusClient.Print(fmt.Sprintf("Application stopping, initiated by: %s", stoppedBy), nil)

	err = mqttClient.Stop()
	if err != nil {
		statusClient.Print("Received error stopping mqtt client", err)
		result = multierror.Append(result, err)
	}

	err = metricsServer.Stop()
	if err != nil {
		statusClient.Print("Received error stopping metrics server", err)
		result = multierror.Append(result, err)
	}

	return result
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
