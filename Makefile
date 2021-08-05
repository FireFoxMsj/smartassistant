.PHONY: start build lint test format

start:
	@go run cmd/smartassistant/main.go

build:
	docker build -f build/docker/Dockerfile -t smartassistant .

push:
	docker image tag smartassistant 192.168.0.44:5000/smartassistant
	docker push 192.168.0.44:5000/smartassistant

run:
	docker-compose -f build/docker/docker-compose.yaml up

lint:
	@golangci-lint run ./...

test:
	go test -cover -v ./internal/...
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