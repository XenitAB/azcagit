TAG = $$(git rev-parse --short HEAD)
IMG ?= ghcr.io/xenitab/node-ttl:$(TAG)

all: fmt vet lint

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

.PHONY: test
test: fmt vet
	go test --cover ./...

docker-build:
	docker build -t ${IMG} .
