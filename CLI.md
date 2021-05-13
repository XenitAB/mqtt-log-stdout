% mqtt-log-stdout 8

# NAME

mqtt-log-stdout - MQTT client that listens to a topic and prints it to stdout

# SYNOPSIS

mqtt-log-stdout

```
[--metrics-address]=[value]
[--metrics-port]=[value]
[--mqtt-broker-addresses]=[value]
[--mqtt-clean-session]
[--mqtt-client-id-random-suffix]
[--mqtt-client-id]=[value]
[--mqtt-connect-timeout]=[value]
[--mqtt-keep-alive]=[value]
[--mqtt-password]=[value]
[--mqtt-port]=[value]
[--mqtt-qos]=[value]
[--mqtt-topic]=[value]
[--mqtt-username]=[value]
```

**Usage**:

```
mqtt-log-stdout [GLOBAL OPTIONS] command [COMMAND OPTIONS] [ARGUMENTS...]
```

# GLOBAL OPTIONS

**--metrics-address**="": The http address metrics should be exposed on (default: 0.0.0.0)

**--metrics-port**="": The http port metrics should be exposed on (default: 8080)

**--mqtt-broker-addresses**="": The MQTT broker addresses

**--mqtt-clean-session**: Should the MQTT client initiate a clean session when subscribing to the topic?

**--mqtt-client-id**="": The MQTT Client ID (defaults to host name)

**--mqtt-client-id-random-suffix**: Should a suffix be appended to the Client ID

**--mqtt-connect-timeout**="": How long should the MQTT client try to connect to the server? (in seconds) (default: 1)

**--mqtt-keep-alive**="": The MQTT keep alive interval in seconds (0 = disabled) (default: 0)

**--mqtt-password**="": The MQTT password

**--mqtt-port**="": The MQTT port (default: 1883)

**--mqtt-qos**="": The MQTT QoS (default: 0)

**--mqtt-topic**="": The MQTT topic to output logs for

**--mqtt-username**="": The MQTT username

