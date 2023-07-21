SHELL := /bin/bash
.ONESHELL:

TAG = $$(git rev-parse --short HEAD)
IMG ?= ghcr.io/xenitab/azcagit:$(TAG)
TEST_ENV_FILE = .tmp/env

ifneq (,$(wildcard $(TEST_ENV_FILE)))
    include $(TEST_ENV_FILE)
    export
endif

.PHONY: all
all: fmt vet lint test

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: build
build:
	CGO_ENABLED=0  go build -installsuffix 'static' -o bin/azcagit ./src/main.go

.PHONY: generate-schema
generate-schema:
	go run ./generate_schema.go

.PHONY: test
test: fmt vet
	go test --cover ./...

.PHONY: cover
cover:
	mkdir -p .tmp
	go test -timeout 5m -coverpkg=./src/... -coverprofile=.tmp/coverage.out ./src/...
	go tool cover -html=.tmp/coverage.out

.PHONY: terraform-up
terraform-up:
	cd test/terraform
	terraform init
	terraform apply -auto-approve -var-file="../../.tmp/lab.tfvars"

.PHONY: terraform-mr-up
terraform-mr-up:
	cd test/terraform-multi-region
	terraform init
	terraform apply -auto-approve -var-file="../../.tmp/lab.tfvars"

.PHONY: run
run:
	# AZURE_TENANT_ID=$${TENANT_ID} AZURE_CLIENT_ID=$${CLIENT_ID} AZURE_CLIENT_SECRET=$${CLIENT_SECRET} \
	go run ./src \
		reconcile \
		--debug \
		--resource-group-name $${RG_NAME} \
		--own-resource-group-name $${OWN_RG_NAME} \
		--subscription-id $${SUB_ID} \
		--managed-environment-id $${ME_ID} \
		--key-vault-name $${KV_NAME} \
		--location westeurope \
		--dapr-topic-name $${DAPR_TOPIC} \
		--reconcile-interval "10s" \
		--git-url $${GIT_URL_AND_CREDS} \
		--git-branch "main" \
		--git-yaml-path "yaml/" \
		--notifications-enabled \
		--environment $${ENV} 

.PHONY: docker-build
docker-build:
	docker build . -t $(IMG)

.PHONY: docker-run
docker-run: docker-build
	docker run -it --rm -e AZURE_TENANT_ID=$${TENANT_ID} -e AZURE_CLIENT_ID=$${CLIENT_ID} -e AZURE_CLIENT_SECRET=$${CLIENT_SECRET} $(IMG) \
		reconcile \
		--debug \
		--resource-group-name $${RG_NAME} \
		--own-resource-group-name $${OWN_RG_NAME} \
		--subscription-id $${SUB_ID} \
		--managed-environment-id $${ME_ID} \
		--key-vault-name $${KV_NAME} \
		--location westeurope \
		--dapr-topic-name $${DAPR_TOPIC} \
		--reconcile-interval "10s" \
		--git-url $${GIT_URL_AND_CREDS} \
		--git-branch "main" \
		--git-yaml-path "yaml/" \
		--notifications-enabled \
		--environment $${ENV} 

.PHONY: k6-http-get
k6-http-get:
	k6 run -e LOAD_TEST_URI=$${LOAD_TEST_URI} test/k6/http_get.js

