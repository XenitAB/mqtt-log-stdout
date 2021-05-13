package mqtt

import (
	"context"
	"fmt"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/xenitab/mqtt-log-stdout/pkg/message"
	"github.com/xenitab/mqtt-log-stdout/pkg/status"
)

// Options takes the input configuration for the mqtt client
type Options struct {
	BrokerAddresses []string
	Topic           string
	QoS             int
	ClientID        string
	Username        string
	Password        string
	CleanSession    bool
	KeepAlive       time.Duration
	ConnectTimeout  time.Duration
	StatusClient    status.Client
	MessageClient   message.Client
}

// Client contains the mqtt client struct
type Client struct {
	topic          string
	qos            int
	done           chan struct{}
	doneMu         sync.Mutex
	stopping       bool
	stoppingMu     sync.Mutex
	connected      bool
	reconnectCount int
	reconnectMu    sync.Mutex
	statusClient   status.Client
	messageClient  message.Client
	mqttClient     pahomqtt.Client
}

// NewClient returns a mqtt client
func NewClient(opts Options) *Client {
	client := &Client{
		topic:          opts.Topic,
		qos:            opts.QoS,
		done:           make(chan struct{}),
		connected:      false,
		reconnectCount: 0,
		stopping:       false,
		statusClient:   opts.StatusClient,
		messageClient:  opts.MessageClient,
	}

	connOpts := pahomqtt.NewClientOptions().SetClientID(opts.ClientID).SetCleanSession(opts.CleanSession).SetKeepAlive(opts.KeepAlive).SetConnectTimeout(opts.ConnectTimeout)
	for _, broker := range opts.BrokerAddresses {
		opts.StatusClient.Print(fmt.Sprintf("Adding mqtt broker: %s", broker), nil)
		connOpts.AddBroker(broker)
	}

	if opts.Username != "" {
		connOpts.SetUsername(opts.Username)
		if opts.Password != "" {
			connOpts.SetPassword(opts.Password)
		}
	}

	connOpts.OnConnect = client.onConnectHandler
	connOpts.OnConnectionLost = client.connectionLostHandler
	connOpts.OnReconnecting = client.reconnectHandler

	mqttClient := pahomqtt.NewClient(connOpts)
	client.mqttClient = mqttClient

	return client
}

// Done returns a channel that is closed if the application is stopped
func (client *Client) Done() <-chan struct{} {
	client.doneMu.Lock()
	if client.done == nil {
		client.done = make(chan struct{})
	}
	d := client.done
	client.doneMu.Unlock()
	return d
}

// Connected returns a bool if the MQTT client is connected or not
func (client *Client) Connected() bool {
	return client.connected
}

func (client *Client) setConnected() {
	metricsConnectionState.Set(1)
	client.connected = true
}

func (client *Client) setDisconnected() {
	metricsConnectionState.Set(0)
	client.connected = false
}

func (client *Client) incReconnectAttempt() {
	client.reconnectMu.Lock()
	client.reconnectCount++
	metricsCurrentReconnectAttempts.Set(float64(client.reconnectCount))
	metricsTotalReconnectAttempts.Inc()
	client.reconnectMu.Unlock()
}

func (client *Client) resetReconnectAttempt() {
	client.reconnectMu.Lock()
	client.reconnectCount = 0
	metricsCurrentReconnectAttempts.Set(0)
	client.reconnectMu.Unlock()
}

// Stop stops the server and calls StopWithContext()
func (client *Client) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.StopWithContext(ctx)
	return err
}

// StopWithContext takes a context and stops the server
func (client *Client) StopWithContext(ctx context.Context) error {
	// Check if MQTT client already has been (or is being) stopped
	client.stoppingMu.Lock()
	if client.stopping {
		client.stoppingMu.Unlock()
		return nil
	}

	client.stopping = true
	client.stoppingMu.Unlock()

	// If server.done already has been closed, it would cause an error
	select {
	case <-client.done:
	default: // Channel is not closed, close it
		close(client.done)
	}

	// Stop the MQTT client
	c := make(chan struct{})
	go func() {
		defer close(c)

		unsubToken := client.mqttClient.Unsubscribe(client.topic)

		unsubMessage := fmt.Sprintf("Unsubscribed from topic: %s", client.topic)
		if unsubToken.Error() != nil {
			unsubMessage = fmt.Sprintf("Unable to gracefully unsubscribe from topic: %s", client.topic)
		}

		client.statusClient.Print(unsubMessage, unsubToken.Error())

		client.mqttClient.Disconnect(250)
		client.statusClient.Print("Disconnected from mqtt broker, stopping client", nil)
	}()

	var err error
	select {
	case <-c:
		err = nil
	case <-ctx.Done():
		err = ctx.Err()
	}

	return err
}

// Start starts the MQTT client
func (client *Client) Start() error {
	token := client.mqttClient.Connect()

	<-token.Done()
	if token.Error() != nil {
		client.statusClient.Print("Unable to connect to mqtt broker", token.Error())
		_ = client.Stop()
		return token.Error()
	}

	return nil
}

func (client *Client) messageHandler(c pahomqtt.Client, m pahomqtt.Message) {
	metricsTotalMessages.Inc()
	message := string(m.Payload())
	client.messageClient.Print(message)
}

func (client *Client) onConnectHandler(c pahomqtt.Client) {
	client.statusClient.Print("Connected to mqtt broker", nil)

	subToken := c.Subscribe(client.topic, byte(client.qos), client.messageHandler)

	<-subToken.Done()
	if subToken.Error() != nil {
		client.statusClient.Print(fmt.Sprintf("Unable to subscribe to topic: %s", client.topic), subToken.Error())
		_ = client.Stop()
		return
	}

	allowed := subscriptionAllowed(subToken, client.topic)
	if !allowed {
		client.statusClient.Print(fmt.Sprintf("Subscription not allowed to topic: %s", client.topic), fmt.Errorf("subscription not allowed"))
		_ = client.Stop()
		return
	}

	client.statusClient.Print(fmt.Sprintf("Subscription started to topic: %s", client.topic), nil)

	client.setConnected()
	if client.reconnectCount > 0 {
		client.resetReconnectAttempt()
	}
}

func (client *Client) connectionLostHandler(c pahomqtt.Client, e error) {
	client.setDisconnected()
	client.statusClient.Print("Connection lost to mqtt broker", e)
}

func (client *Client) reconnectHandler(c pahomqtt.Client, co *pahomqtt.ClientOptions) {
	// setDisconnected() isn't needed here as OnReconnecting is called as the same time as OnConnectionLost: https://github.com/eclipse/paho.mqtt.golang/blob/master/client.go#L491
	client.incReconnectAttempt()
	client.statusClient.Print(fmt.Sprintf("Reconnecting to mqtt broker, attempt: %d", client.reconnectCount), nil)
}

func subscriptionAllowed(token pahomqtt.Token, topic string) bool {
	subscriptionToken, ok := token.(*pahomqtt.SubscribeToken)
	if !ok {
		return false
	}

	result := subscriptionToken.Result()
	res, found := result[topic]
	if !found {
		return false
	}

	if res >= 128 {
		return false
	}

	return true
}
