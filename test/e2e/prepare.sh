#!/bin/bash
set -e

IMG=$1

mkdir -p tmp

timestamp() {
  date +"%T"
}

# Cleanup
echo "$(timestamp) [PREPARE] Docker clean up started."
set +e
docker stop $(docker ps -f name=vernemq-e2e -q) 1>/dev/null 2>&1
docker rm $(docker ps -a -f name=vernemq-e2e -q) 1>/dev/null 2>&1
docker stop $(docker ps -f name=mqtt-log-stdout-e2e -q) 1>/dev/null 2>&1
docker rm $(docker ps -a -f name=mqtt-log-stdout-e2e -q) 1>/dev/null 2>&1
docker network rm endtoend 1>/dev/null 2>&1
set -e
echo "$(timestamp) [PREPARE] Docker clean up finished."


if [[ "${CI}" == "true" ]]; then
    sudo apt-get install mosquitto-clients parallel 1>/dev/null
    echo "$(timestamp) [PREPARE] Installed pre-requisites in CI"
fi

if [[ "${CI}" != "true" ]]; then
    docker build -t ${IMG} . 1>/dev/null
    VERSION="dev"
    echo "$(timestamp) [PREPARE] Built mqtt-log-stdout outside of CI"
fi

docker network create --driver bridge endtoend 1>/dev/null
echo "$(timestamp) [PREPARE] Starting vernemq"
docker run --network endtoend -p 1883:1883 -v "$(pwd)"/test/vernemq:/mnt -e "DOCKER_VERNEMQ_VMQ_ACL__ACL_FILE=/mnt/vmq.acl" -e "DOCKER_VERNEMQ_ACCEPT_EULA=yes" -e "DOCKER_VERNEMQ_ALLOW_ANONYMOUS=on" --name vernemq-e2e -d vernemq/vernemq 1>/dev/null 2>&1
echo "$(timestamp) [PREPARE] Started vernemq"

echo "$(timestamp) [PREPARE] Validating connection to mqtt"
VALID_CONNECTION=false
for i in `seq 1 100`; do
    set +e
    mosquitto_pub -h localhost -p 1883 -q 1 -t "test/validate_connection" -i "validate-connection" -m "validate-connection" 1>/dev/null 2>/dev/null
    EXIT_CODE=$?
    set -e
    if [[ ${EXIT_CODE} -eq 0 ]]; then
        VALID_CONNECTION=true
        break
    fi
    sleep 0.1
done

if [ "${VALID_CONNECTION}" = false ]; then
    echo "$(timestamp) [PREPARE] FAILURE: Not able to establish connection to MQTT server."
    exit 1
fi

echo "$(timestamp) [PREPARE] Connection to mqtt validated"

echo MQTT_BROKER_ADDRESSES=vernemq-e2e > tmp/e2e_env
echo MQTT_TOPIC=test/log_entry >> tmp/e2e_env

docker run --network endtoend --env-file ./tmp/e2e_env -p 8080:8080 --name mqtt-log-stdout-e2e -d ${IMG} 1>/dev/null
echo "$(timestamp) [PREPARE] mqtt-log-stdout started"
