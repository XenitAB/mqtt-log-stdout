#!/bin/bash

echo "[CLEANUP] Docker clean up started."

docker stop $(docker ps -f name=vernemq-e2e -q) 1>/dev/null 2>&1
docker rm $(docker ps -a -f name=vernemq-e2e -q) 1>/dev/null 2>&1
docker stop $(docker ps -f name=mqtt-log-stdout -q) 1>/dev/null 2>&1
docker rm $(docker ps -a -f name=mqtt-log-stdout -q) 1>/dev/null 2>&1
docker network rm endtoend 1>/dev/null 2>&1

echo "[CLEANUP] Docker clean up finished."

exit 0
