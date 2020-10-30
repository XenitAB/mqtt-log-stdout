# mqtt-log-stdout

Log messages from a MQTT topicLog messages from a MQTT topic

## Usage

The project includes a Docker image with a static binary that needs the following environment variables:

```
# at least one of
MQTT_HOST_1=
MQTT_HOST_2=
MQTT_HOST_3=
# the port to use when connecting to mqtt
MQTT_PORT=1883
# the topic to listen on
LOG_TOPIC=$share/logger/client/+/v1/log
```

## Development

You need [esy](https://github.com/esy/esy), you can install the beta using [npm](https://npmjs.com):

    npm install -g esy@latest

> NOTE: Make sure `esy --version` returns at least `0.6.0` for this project to build.

Then run the `esy` command from this project root to install and build depenencies.
After you make some changes to source code, you can re-run project's build
again with the same simple `esy` command.

When you want to test the application you can run these commands in separate windows

    esy x MqttLogStdoutApp
    esy x MqttLogStdoutAppPublish

> NOTE: You need to have a mqtt running, there is a docker-compose.yml in the repo that runs this
