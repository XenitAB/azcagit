SHELL := /bin/bash
.ONESHELL:

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

terraform-up:
	cd test/terraform
	terraform init
	terraform apply -auto-approve

run:
	go run ./src \
		--resource-group-name "rg-aca-tenant" \
		--subscription-id "2a6936a5-fc30-492a-ab19-ec59068b5b96" \
		--managed-environment-id "/subscriptions/2a6936a5-fc30-492a-ab19-ec59068b5b96/resourceGroups/rg-aca-platform/providers/Microsoft.App/managedEnvironments/me-container-apps" \
		--location westeurope \
		--reconcile-interval "10s" \
		--checkout-path "/tmp/foo" \
		--git-url "https://github.com/simongottschlag/aca-test-yaml.git" \
		--git-branch "main" \
		--git-yaml-path "yaml/"
