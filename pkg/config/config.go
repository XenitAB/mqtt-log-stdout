package config

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	NullClient  = Client{}
	NullOptions = Options{}
)

// Options takes the build information and provides the CLI with correct information
type Options struct {
	Version  string
	Revision string
	Created  string
	// DisableExitOnHelp configures if --help should exit or not, used with helpPrinter()
	DisableExitOnHelp bool
}

// Client struct
type Client struct {
	BrokerAddresses   []string
	Topic             string
	QoS               int
	KeepAlive         time.Duration
	ConnectTimeout    time.Duration
	CleanSession      bool
	Username          string
	Password          string
	ClientID          string
	MetricsAddress    string
	MetricsPort       int
	disableExitOnHelp bool
	cliReader         io.Reader
	cliWriter         io.Writer
	cliErrWriter      io.Writer
	version           string
	revision          string
	created           string
}

// NewClient returns the Client or error
func NewClient(opts Options) (Client, error) {
	client := newClient(opts)
	generatedCfg, err := client.generateConfig(os.Args)
	if err != nil {
		return NullClient, err
	}

	return generatedCfg, nil
}

// GenerateMarkdown creates a markdown file with documentation for the application
func GenerateMarkdown(filePath string) error {
	client := newClient(NullOptions)
	err := client.generateMarkdown(filePath)
	return err
}

func newClient(opts Options) *Client {
	return &Client{
		disableExitOnHelp: opts.DisableExitOnHelp,
		cliReader:         os.Stdin,
		cliWriter:         os.Stdout,
		cliErrWriter:      os.Stderr,
		version:           opts.Version,
		revision:          opts.Revision,
		created:           opts.Created,
	}
}

func (client *Client) generateConfig(args []string) (Client, error) {
	app := client.newCLIApp()

	err := app.Run(args)
	if err != nil {
		return NullClient, err
	}

	return *client, nil
}

func (client *Client) generateMarkdown(filePath string) error {
	app := client.newCLIApp()

	md, err := app.ToMarkdown()
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, []byte(md), 0666) // #nosec
	return err
}

func (client *Client) setConfig(cfg Client) {
	client.BrokerAddresses = cfg.BrokerAddresses
	client.Topic = cfg.Topic
	client.QoS = cfg.QoS
	client.KeepAlive = cfg.KeepAlive
	client.ConnectTimeout = cfg.ConnectTimeout
	client.CleanSession = cfg.CleanSession
	client.Username = cfg.Username
	client.Password = cfg.Password
	client.ClientID = cfg.ClientID
	client.MetricsAddress = cfg.MetricsAddress
	client.MetricsPort = cfg.MetricsPort
}

func (client *Client) setIO(reader io.Reader, writer io.Writer, errWriter io.Writer) {
	client.cliReader = reader
	client.cliWriter = writer
	client.cliErrWriter = errWriter
}

func (client *Client) newCLIApp() *cli.App {
	cli.VersionPrinter = client.versionHandler
	cli.HelpPrinter = client.helpPrinter

	app := &cli.App{
		Name:    "mqtt-log-stdout",
		Usage:   "MQTT client that listens to a topic and prints it to stdout",
		Version: client.version,
		Flags:   client.newCLIFlags(),
		Action:  client.newCLIAction,
	}

	app.Writer = client.cliWriter
	app.ErrWriter = client.cliErrWriter
	app.Reader = client.cliReader

	return app
}

func (client *Client) newCLIAction(c *cli.Context) error {
	err := client.setConfigFromCLI(c)
	return err
}

func (client *Client) newCLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:     "mqtt-broker-addresses",
			Usage:    "The MQTT broker addresses",
			Required: true,
			EnvVars:  []string{"MQTT_BROKER_ADDRESSES", "MQTT_HOST_1", "MQTT_HOST_2", "MQTT_HOST_3"},
		},
		&cli.StringFlag{
			Name:     "mqtt-topic",
			Usage:    "The MQTT topic to output logs for",
			Required: true,
			EnvVars:  []string{"MQTT_TOPIC", "LOG_TOPIC"},
		},
		&cli.IntFlag{
			Name:     "mqtt-port",
			Usage:    "The MQTT port",
			Required: false,
			EnvVars:  []string{"MQTT_PORT"},
			Value:    1883,
		},
		&cli.IntFlag{
			Name:     "mqtt-qos",
			Usage:    "The MQTT QoS",
			Required: false,
			EnvVars:  []string{"MQTT_QOS"},
			Value:    0,
		},
		&cli.IntFlag{
			Name:     "mqtt-keep-alive",
			Usage:    "The MQTT keep alive interval in seconds (0 = disabled)",
			Required: false,
			EnvVars:  []string{"MQTT_KEEP_ALIVE"},
			Value:    0,
		},
		&cli.IntFlag{
			Name:     "mqtt-connect-timeout",
			Usage:    "How long should the MQTT client try to connect to the server? (in seconds)",
			Required: false,
			EnvVars:  []string{"MQTT_CONNECT_TIMEOUT"},
			Value:    1,
		},
		&cli.BoolFlag{
			Name:     "mqtt-clean-session",
			Usage:    "Should the MQTT client initiate a clean session when subscribing to the topic?",
			Required: false,
			EnvVars:  []string{"MQTT_CLEAN_SESSION"},
			Value:    false,
		},
		&cli.StringFlag{
			Name:     "mqtt-username",
			Usage:    "The MQTT username",
			Required: false,
			EnvVars:  []string{"MQTT_USERNAME"},
		},
		&cli.StringFlag{
			Name:     "mqtt-password",
			Usage:    "The MQTT password",
			Required: false,
			EnvVars:  []string{"MQTT_PASSWORD"},
		},
		&cli.StringFlag{
			Name:     "mqtt-client-id",
			Usage:    "The MQTT Client ID (defaults to host name)",
			Required: false,
			EnvVars:  []string{"MQTT_CLIENT_ID"},
		},
		&cli.BoolFlag{
			Name:     "mqtt-client-id-random-suffix",
			Usage:    "Should a suffix be appended to the Client ID",
			Required: false,
			EnvVars:  []string{"MQTT_CLIENT_ID_RANDOM_SUFFIX"},
			Value:    false,
		},
		&cli.StringFlag{
			Name:     "metrics-address",
			Usage:    "The http address metrics should be exposed on",
			Required: false,
			EnvVars:  []string{"METRICS_ADDRESS"},
			Value:    "0.0.0.0",
		},
		&cli.IntFlag{
			Name:     "metrics-port",
			Usage:    "The http port metrics should be exposed on",
			Required: false,
			EnvVars:  []string{"METRICS_PORT"},
			Value:    8080,
		},
	}
}

func (client *Client) setConfigFromCLI(cli *cli.Context) error {
	flagMqttClientID := cli.String("mqtt-client-id")
	flagMqttClientIDRandomSuffix := cli.Bool("mqtt-client-id-random-suffix")
	mqttClientID, err := getClientID(flagMqttClientID, flagMqttClientIDRandomSuffix)
	if err != nil {
		return err
	}

	flagMqttBrokerAddresses := cli.StringSlice("mqtt-broker-addresses")
	flagMqttPort := cli.Int("mqtt-port")
	brokerAddresses := getBrokerAddresses(flagMqttBrokerAddresses, flagMqttPort)

	flagQoS := cli.Int("mqtt-qos")
	qos, err := getQoS(flagQoS)
	if err != nil {
		return err
	}

	keepAlive := time.Duration(cli.Int("mqtt-keep-alive")) * time.Second
	connectTimeout := time.Duration(cli.Int("mqtt-connect-timeout")) * time.Second

	newCfg := Client{
		BrokerAddresses: brokerAddresses,
		Topic:           cli.String("mqtt-topic"),
		QoS:             qos,
		KeepAlive:       keepAlive,
		ConnectTimeout:  connectTimeout,
		CleanSession:    cli.Bool("mqtt-clean-session"),
		Username:        cli.String("mqtt-username"),
		Password:        cli.String("mqtt-password"),
		ClientID:        mqttClientID,
		MetricsAddress:  cli.String("metrics-address"),
		MetricsPort:     cli.Int("metrics-port"),
	}

	client.setConfig(newCfg)

	return nil
}

func (client *Client) versionHandler(c *cli.Context) {
	fmt.Printf("version=%s revision=%s created=%s\n", client.version, client.revision, client.created)
	os.Exit(0)
}

// helpPrinter uses the default HelpPrinterCustom() but adds an os.Exit(0)
func (client *Client) helpPrinter(out io.Writer, templ string, data interface{}) {
	cli.HelpPrinterCustom(out, templ, data, nil)
	if !client.disableExitOnHelp {
		os.Exit(0)
	}
}

func getClientID(clientID string, randomSuffix bool) (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	if clientID == "" {
		clientID = hostname
	}

	if randomSuffix {
		return getRandomClientID(clientID)
	}

	return clientID, nil
}

func getRandomClientID(clientID string) (string, error) {
	randomString, err := generateRandomString(5)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s", clientID, randomString), nil

}

func getBrokerAddresses(brokerAddresses []string, port int) []string {
	var newBrokers []string
	for _, broker := range brokerAddresses {
		b := fmt.Sprintf("tcp://%s:%d", broker, port)
		if strings.Contains(broker, ":") {
			b = fmt.Sprintf("tcp://%s", broker)
		}
		newBrokers = append(newBrokers, b)
	}

	return newBrokers
}

func getQoS(qos int) (int, error) {
	if qos < 0 || qos > 1 {
		return 0, fmt.Errorf("QoS allowed to be 0 or 1, received: %d", qos)
	}

	return qos, nil
}

func generateRandomString(n int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
