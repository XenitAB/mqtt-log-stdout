package mqtt

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// metricsConnectionState shows the connection state of the MQTT client
	metricsConnectionState = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "mqtt_client_connection_state",
		Help: "Connection state of the MQTT client",
	})

	// metricsTotalMessages shows the total number of messages since start
	metricsTotalMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "mqtt_client_total_messages",
		Help: "Total number of messages handled by the MQTT client",
	})

	// metricsCurrentReconnectAttempts shows the current number of reconnect attempts
	metricsCurrentReconnectAttempts = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "mqtt_client_current_reconnect_attempts",
		Help: "Current number of reconnect attempts by the MQTT client",
	})

	// metricsTotalReconnectAttempts shows the total number of reconnect attempts
	metricsTotalReconnectAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "mqtt_client_total_reconnect_attempts",
		Help: "Total number of reconnect attempts by the MQTT client",
	})
)
