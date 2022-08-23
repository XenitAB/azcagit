SHELL := /bin/bash
.ONESHELL:

TAG = $$(git rev-parse --short HEAD)
IMG ?= ghcr.io/xenitab/azcagit:$(TAG)
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

.PHONY: build
build:
	go build -o bin/azcagit ./src/main.go

.PHONY: test
test: fmt vet
	go test --cover ./...

cover:
	mkdir -p .tmp
	go test -timeout 5m -coverpkg=./src/... -coverprofile=.tmp/coverage.out ./src/...
	go tool cover -html=.tmp/coverage.out

terraform-up:
	cd test/terraform
	terraform init
	terraform apply -auto-approve

run:
	AZURE_TENANT_ID=$${TENANT_ID} AZURE_CLIENT_ID=$${CLIENT_ID} AZURE_CLIENT_SECRET=$${CLIENT_SECRET} go run ./src \
		--resource-group-name $${RG_NAME} \
		--subscription-id $${SUB_ID} \
		--managed-environment-id $${ME_ID} \
		--location westeurope \
		--reconcile-interval "10s" \
		--checkout-path "/tmp/foo" \
		--git-url $${GIT_URL_AND_CREDS} \
		--git-branch "main" \
		--git-yaml-path "yaml/"

docker-build:
	docker build . -t $(IMG)

docker-run: docker-build
	docker run -it --rm -e AZURE_TENANT_ID=$${TENANT_ID} -e AZURE_CLIENT_ID=$${CLIENT_ID} -e AZURE_CLIENT_SECRET=$${CLIENT_SECRET} $(IMG) \
		--resource-group-name $${RG_NAME} \
		--subscription-id $${SUB_ID} \
		--managed-environment-id $${ME_ID} \
		--location westeurope \
		--reconcile-interval "10s" \
		--checkout-path "/tmp/foo" \
		--git-url $${GIT_URL_AND_CREDS} \
		--git-branch "main" \
		--git-yaml-path "yaml/"