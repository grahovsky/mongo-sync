GIT_HASH := $(shell git log --format="%h" -n 1)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

BIN := "./bin/mongo-sync"
DOCKER_IMG := mongo-sync:$(GIT_BRANCH)

LDFLAGS := -X main.release=$(GIT_BRANCH) -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/*.go

run: build
	$(BIN) --config ./configs/config.yaml


build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run -v $$(pwd)/configs:/configs $(DOCKER_IMG) 

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2

lint: install-lint-deps
	golangci-lint run --timeout=90s ./...

test:
	go test -race ./internal/... -count 100

generate: 
	go generate ./...


.PHONY: build run build-img run-img test lint
