.PHONY: start build lint test format
VERSION=latest

start:
	@go run cmd/smartassistant/main.go -c ./app.yaml

# make build-all VERSION=1.0.0
build-all: build build-supervisor

build:
	docker build -f build/docker/Dockerfile --build-arg VERSION=$(VERSION) --build-arg GIT_COMMIT=$(shell git log -1 --format=%h) -t smartassistant:$(VERSION) .

build-supervisor:
	docker build -f build/docker/supervisor.Dockerfile --build-arg VERSION=$(VERSION) --build-arg GIT_COMMIT=$(shell git log -1 --format=%h) -t supervisor:$(VERSION) .

push:
	docker image tag smartassistant:$(VERSION) docker.yctc.tech/smartassistant:$(VERSION)
	docker push docker.yctc.tech/smartassistant:$(VERSION)
	docker image tag supervisor:$(VERSION) docker.yctc.tech/supervisor:$(VERSION)
	docker push docker.yctc.tech/supervisor:$(VERSION)

run:
	docker-compose -f build/docker/docker-compose.yaml up

lint:
	@golangci-lint run ./...

test:
	go test -cover -v ./modules/...
	go test -cover -v ./pkg/...

format:
	@find -type f -name '*.go' | $(XARGS) gofmt -s -w

.PHONY: install.golangci-lint
install.golangci-lint:
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: install.goimports
install.goimports:
	@go get -u golang.org/x/tools/cmd/goimports

build-plugin-demo:
	docker build -f build/docker/demo.Dockerfile -t demo-plugin  .

