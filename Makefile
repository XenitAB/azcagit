SHELL := /bin/bash
.ONESHELL:

TAG = $$(git rev-parse --short HEAD)
IMG ?= ghcr.io/xenitab/aca-gitops-engine:$(TAG)
TEST_ENV_FILE = .tmp/env

ifneq (,$(wildcard $(TEST_ENV_FILE)))
    include $(TEST_ENV_FILE)
    export
endif

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
		--resource-group-name $${RG_NAME} \
		--subscription-id $${SUB_ID} \
		--managed-environment-id $${ME_ID} \
		--location westeurope \
		--reconcile-interval "10s" \
		--checkout-path "/tmp/foo" \
		--git-url $${GIT_URL_AND_CREDS} \
		--git-branch "main" \
		--git-yaml-path "yaml/"
