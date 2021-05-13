package mqtt

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	hmqBroker "github.com/fhmq/hmq/broker"
	"github.com/xenitab/mqtt-log-stdout/pkg/message"
	"github.com/xenitab/mqtt-log-stdout/pkg/status"
	"go.uber.org/goleak"
)

func TestStart(t *testing.T) {
	args := []string{""}
	hmqConfig, err := hmqBroker.ConfigureConfig(args)
	if err != nil {
		t.Errorf("Expected err to be nil: %q", err)
	}

	mqttBroker, err := hmqBroker.NewBroker(hmqConfig)
	if err != nil {
		t.Errorf("Expected err to be nil: %q", err)
	}
	mqttBroker.Start()

	mockBroker := net.JoinHostPort(hmqConfig.Host, hmqConfig.Port)

	// Check that the in-memory broker is started
	for start := time.Now(); time.Since(start) < 5*time.Second; {
		conn, err := net.Dial("tcp", mockBroker)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	statusClient := newFakeStatusClient()
	messageClient := newFakeMessageClient()

	opts := Options{
		BrokerAddresses: []string{mockBroker},
		Topic:           "fake-topic",
		QoS:             0,
		ClientID:        "sub-client",
		Username:        "",
		Password:        "",
		CleanSession:    false,
		KeepAlive:       time.Duration(0 * time.Second),
		ConnectTimeout:  time.Duration(1 * time.Second),
		StatusClient:    statusClient,
		MessageClient:   messageClient,
	}

	mqttClient := NewClient(opts)
	err = mqttClient.Start()
	if err != nil {
		t.Errorf("Expected err to be nil: %q", err)
	}

	for start := time.Now(); time.Since(start) < 5*time.Second; {
		if mqttClient.Connected() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Publish message here
	publishHost := fmt.Sprintf("tcp://%s", mockBroker)
	connOpts := pahomqtt.NewClientOptions().SetClientID("pub-client").SetCleanSession(true).SetKeepAlive(opts.KeepAlive).AddBroker(publishHost)

	publishMqttClient := pahomqtt.NewClient(connOpts)
	if token := publishMqttClient.Connect(); token.Wait() && token.Error() != nil {
		t.Errorf("Expected err to be nil: %q", token.Error())
	}

	expectedMessageCount := 200
	for i := 0; i < expectedMessageCount; i++ {
		message := fmt.Sprintf("test message %d", i)
		publishToken := publishMqttClient.Publish(opts.Topic, byte(opts.QoS), false, message)

		<-publishToken.Done()
		if publishToken.Error() != nil {
			t.Errorf("Expected err to be nil: %q", publishToken.Error())
		}
	}

	var messageCount int
	for start := time.Now(); time.Since(start) < 5*time.Second; {
		fakeMessageClient := messageClient.(*fakeMessage)
		messageCount = len(fakeMessageClient.messages)
		if messageCount == expectedMessageCount {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = mqttClient.StopWithContext(ctx)
	if err != nil {
		t.Errorf("Expected err to be nil: %q", err)
	}

	if messageCount != expectedMessageCount {
		t.Errorf("Expected messageCount to be %d but was %d", expectedMessageCount, messageCount)
	}

}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("github.com/eclipse/paho%2emqtt%2egolang.(*client).startCommsWorkers.func2"),
		goleak.IgnoreTopFunction("github.com/eclipse/paho%2emqtt%2egolang.startComms.func2"),
		goleak.IgnoreTopFunction("github.com/eclipse/paho%2emqtt%2egolang.startComms.func1"),
		goleak.IgnoreTopFunction("github.com/eclipse/paho%2emqtt%2egolang.startOutgoingComms.func1"),
		goleak.IgnoreTopFunction("github.com/eclipse/paho%2emqtt%2egolang.startIncomingComms.func1"),
		goleak.IgnoreTopFunction("github.com/eclipse/paho%2emqtt%2egolang.(*client).startCommsWorkers.func1"),
		goleak.IgnoreTopFunction("github.com/eclipse/paho%2emqtt%2egolang.(*router).matchAndDispatch.func1"),
		goleak.IgnoreTopFunction("github.com/eclipse/paho%2emqtt%2egolang.keepalive"),
		goleak.IgnoreTopFunction("github.com/fhmq/hmq/pool.startWorker.func1"),
		goleak.IgnoreTopFunction("sync.runtime_Semacquire"),
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
		goleak.IgnoreTopFunction("github.com/patrickmn/go-cache.(*janitor).Run"),
	)
}

type fakeMessage struct {
	messages []string
}

func newFakeMessageClient() message.Client {
	return &fakeMessage{
		messages: []string{},
	}
}

func (client *fakeMessage) Print(m string) {
	client.messages = append(client.messages, m)
}

type fakeStatus struct{}

func newFakeStatusClient() status.Client {
	return &fakeStatus{}
}

func (s *fakeStatus) Print(m string, e error) {}
