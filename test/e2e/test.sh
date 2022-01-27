#!/bin/bash

set -e

timestamp() {
  date +"%T"
}

publish_messages() {
    for i in `seq 1 $1`; do
        echo "End-to-end test message (publisher-${2}): ${i}"
    done | mosquitto_pub -l -h localhost -p 1883 -q 1 -t "test/log_entry" -i "publisher-${2}"
}

export -f publish_messages

ITERATIONS=20
WORKERS=5
MESSAGES_PER_WORKER=200
EXPECTED_NUM_MESSAGES=$(expr ${ITERATIONS} \* ${MESSAGES_PER_WORKER})

echo "$(timestamp) [TEST] Starting to publish messages. Iterations: ${ITERATIONS}, Workers: ${WORKERS}, Messages per worker: ${MESSAGES_PER_WORKER}"
seq 1 ${ITERATIONS} | parallel -j${WORKERS} "publish_messages ${MESSAGES_PER_WORKER} {}"
echo "$(timestamp) [TEST] Messages published to mqtt"

for i in `seq 1 100`; do
    NUM_MESSAGES_RECEIVED=$(docker logs mqtt-log-stdout-e2e | grep "End-to-end test message" | wc -l)
    if [[ ${NUM_MESSAGES_RECEIVED} -eq ${EXPECTED_NUM_MESSAGES} ]]; then
        break
    fi
    sleep 0.1
done

for i in `seq 1 ${ITERATIONS}`; do
    WORKER_MESSAGES=$(docker logs mqtt-log-stdout-e2e | grep "End-to-end test message (publisher-${i})" | wc -l)
    echo "$(timestamp) [TEST] Publisher #${i}: ${WORKER_MESSAGES}"
done

if [[ ${NUM_MESSAGES_RECEIVED} -ne ${EXPECTED_NUM_MESSAGES} ]]; then
    echo "$(timestamp) [TEST] FAILURE: Expected ${EXPECTED_NUM_MESSAGES} messages received. Was: ${NUM_MESSAGES_RECEIVED}"
    exit 1
fi

METRICS_COUNTER=$(curl -s localhost:8080/metrics | grep "mqtt_client_total_messages" | grep -v "#" | awk '{print $2}')
echo "$(timestamp) [TEST] Metrics counter: ${METRICS_COUNTER}"
if [[ ${METRICS_COUNTER} -ne ${EXPECTED_NUM_MESSAGES} ]]; then
    echo "$(timestamp) [TEST] FAILURE: Expected metrics counter to be ${EXPECTED_NUM_MESSAGES}. Was: ${METRICS_COUNTER}"
    exit 1
fi