package config

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	envVarsToClear := []string{
		"MQTT_BROKER_ADDRESSES",
		"MQTT_HOST_1",
		"MQTT_HOST_2",
		"MQTT_HOST_3",
		"MQTT_TOPIC",
		"LOG_TOPIC",
		"MQTT_PORT",
		"MQTT_QOS",
		"MQTT_KEEP_ALIVE",
		"MQTT_CONNECT_TIMEOUT",
		"MQTT_CLEAN_SESSION",
		"MQTT_USERNAME",
		"MQTT_PASSWORD",
		"MQTT_CLIENT_ID",
		"MQTT_CLIENT_ID_RANDOM_SUFFIX",
		"METRICS_ENABLED",
		"METRICS_ADDRESS",
		"METRICS_PORT",
	}

	for _, envVar := range envVarsToClear {
		restore := tempUnsetEnv(envVar)
		defer restore()
	}

	cliClient := newClient(Options{
		DisableExitOnHelp: true,
	})

	baseArgs := []string{"fake-bin"}
	baseWorkingArgs := append(baseArgs, "--mqtt-broker-addresses=test", "--mqtt-topic=fake")

	cases := []struct {
		client              *Client
		args                []string
		expectedHosts       []string
		expectedErrContains string
		outBuffer           bytes.Buffer
		errBuffer           bytes.Buffer
	}{
		{
			client:              cliClient,
			args:                baseArgs,
			expectedErrContains: "Required flags \"mqtt-broker-addresses, mqtt-topic\" not set",
			outBuffer:           bytes.Buffer{},
			errBuffer:           bytes.Buffer{},
		},
		{
			client:              cliClient,
			args:                append(baseArgs, "--mqtt-broker-addresses=abc"),
			expectedErrContains: "Required flag \"mqtt-topic\" not set",
			outBuffer:           bytes.Buffer{},
			errBuffer:           bytes.Buffer{},
		},
		{
			client:              cliClient,
			args:                baseWorkingArgs,
			expectedErrContains: "",
			outBuffer:           bytes.Buffer{},
			errBuffer:           bytes.Buffer{},
		},
		{
			client:              cliClient,
			args:                append(baseArgs, "--mqtt-broker-addresses=test:1883", "--mqtt-topic=fake"),
			expectedErrContains: "",
			outBuffer:           bytes.Buffer{},
			errBuffer:           bytes.Buffer{},
		},
	}

	for _, c := range cases {
		c.client.setIO(&bytes.Buffer{}, &c.outBuffer, &c.errBuffer)
		cfg, err := c.client.generateConfig(c.args)
		if err != nil && c.expectedErrContains == "" {
			t.Errorf("Expected err to be nil: %q", err)
		}

		if err == nil && c.expectedErrContains != "" {
			t.Errorf("Expected err to contain '%s' but was nil", c.expectedErrContains)
		}

		if err != nil && c.expectedErrContains != "" {
			if !strings.Contains(err.Error(), c.expectedErrContains) {
				t.Errorf("Expected err to contain '%s' but was: %q", c.expectedErrContains, err)
			}
		}

		if c.expectedErrContains == "" {
			if cfg.QoS != 0 {
				t.Errorf("Expected cfg.QoS to be '0' but was: %d", cfg.QoS)
			}
		}
	}
}

func tempUnsetEnv(key string) func() {
	oldEnv := os.Getenv(key)
	os.Unsetenv(key)
	return func() { os.Setenv(key, oldEnv) }
}
