package mqtt

import (
	"fmt"
	"net"
	"testing"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	hmqBroker "github.com/fhmq/hmq/broker"
	"github.com/stretchr/testify/require"
	h "github.com/xenitab/mqtt-log-stdout/pkg/helper"
	"github.com/xenitab/mqtt-log-stdout/pkg/message"
	"github.com/xenitab/mqtt-log-stdout/pkg/status"
	"go.uber.org/goleak"
)

func TestStart(t *testing.T) {
	errGroup, ctx, cancel := h.NewErrGroupAndContext()
	defer cancel()

	args := []string{""}
	hmqConfig, err := hmqBroker.ConfigureConfig(args)
	require.NoError(t, err)

	mqttBroker, err := hmqBroker.NewBroker(hmqConfig)
	require.NoError(t, err)
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

	statusClient := testNewFakeStatusClient(t)
	messageClient := testNewFakeMessageClient(t)

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

	h.StartService(ctx, errGroup, mqttClient)

	// Check that the mqtt client is connected
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
	token := publishMqttClient.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	numberOfWorkers := 10
	messagesPerWorker := 200
	expectedMessageCount := messagesPerWorker * numberOfWorkers
	publisherErrGroup, _, _ := h.NewErrGroupAndContext()

	for w := 0; w < numberOfWorkers; w++ {
		publisherErrGroup.Go(func() error {
			for i := 0; i < messagesPerWorker; i++ {
				message := fmt.Sprintf("test message %d", i)
				publishToken := publishMqttClient.Publish(opts.Topic, byte(opts.QoS), false, message)

				<-publishToken.Done()
				require.NoError(t, publishToken.Error())
			}

			return nil
		})
	}

	err = h.WaitForErrGroup(publisherErrGroup)
	require.NoError(t, err)

	var messageCount int
	for start := time.Now(); time.Since(start) < 5*time.Second; {
		fakeMessageClient := messageClient.(*testFakeMessage)
		messageCount = len(fakeMessageClient.messages)
		if messageCount == expectedMessageCount {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()

	timeoutCtx, timeoutCancel := h.NewShutdownTimeoutContext()
	defer timeoutCancel()

	h.StopService(timeoutCtx, errGroup, mqttClient)

	err = h.WaitForErrGroup(errGroup)
	require.NoError(t, err)

	require.Equal(t, expectedMessageCount, messageCount)
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

type testFakeMessage struct {
	t        *testing.T
	messages []string
}

func testNewFakeMessageClient(t *testing.T) message.Client {
	t.Helper()

	return &testFakeMessage{
		t:        t,
		messages: []string{},
	}
}

func (client *testFakeMessage) Print(m string) {
	client.t.Helper()

	client.messages = append(client.messages, m)
}

type testFakeStatus struct {
	t *testing.T
}

func testNewFakeStatusClient(t *testing.T) status.Client {
	t.Helper()

	return &testFakeStatus{
		t: t,
	}
}

func (s *testFakeStatus) Print(m string, e error) {
	s.t.Helper()
}
