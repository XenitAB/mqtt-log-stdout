SHELL := /bin/bash

TAG = dev
IMG ?= mqtt-log-stdout:$(TAG)
TEST_ENV_FILE = tmp/test_env
VERSION ?= "v0.0.0-dev"
REVISION ?= ""
CREATED ?= ""

ifneq (,$(wildcard $(TEST_ENV_FILE)))
    include $(TEST_ENV_FILE)
    export
endif

.PHONY: all
.SILENT: all
all: tidy lint fmt vet test build

.PHONY: lint
.SILENT: lint
lint:
	golangci-lint run

.PHONY: fmt
.SILENT: fmt
fmt:
	go fmt ./...

.PHONY: tidy
.SILENT: tidy
tidy:
	go mod tidy

.PHONY: vet
.SILENT: vet
vet:
	go vet ./...

.PHONY: test
.SILENT: test
test:
	mkdir -p tmp
	go test -timeout 1m ./... -cover

.PHONY: start-mqtt
.SILENT: start-mqtt
start-mqtt:
	docker run -p 1883:1883 -v "$$(pwd)"/test/vernemq:/mnt -e "DOCKER_VERNEMQ_VMQ_ACL__ACL_FILE=/mnt/vmq.acl" -e "DOCKER_VERNEMQ_ACCEPT_EULA=yes" -e "DOCKER_VERNEMQ_ALLOW_ANONYMOUS=on" -e "DOCKER_VERNEMQ_LOG__CONSOLE__LEVEL=debug" --name vernemq1 -d vernemq/vernemq

.PHONY: stop-mqtt
.SILENT: stop-mqtt
stop-mqtt:
	-docker stop $$(docker ps -f name=vernemq1 -q)
	docker rm $$(docker ps -a -f name=vernemq1 -q)

.PHONY: cover
.SILENT: cover
cover:
	go test -timeout 1m -coverpkg=./... -coverprofile=tmp/coverage.out ./...
	go tool cover -html=tmp/coverage.out	

.PHONY: run
.SILENT: run
run:
	go run cmd/mqtt-log-stdout/main.go --mqtt-broker-addresses=127.0.0.1 --mqtt-topic "test/log_entry"

.PHONY: gen-docs
.SILENT: gen-docs
gen-docs:
	go run cmd/gen-docs/main.go

.PHONY: build
.SILENT: build
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-w -s -X main.Version=$(VERSION) -X main.Revision=$(REVISION) -X main.Created=$(CREATED)" -o bin/mqtt-log-stdout cmd/mqtt-log-stdout/main.go

.PHONY: e2e
.SILENT: e2e
e2e:
	./test/e2e/prepare.sh $(IMG)
	./test/e2e/test.sh
	./test/e2e/cleanup.sh
