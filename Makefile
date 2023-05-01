HOST_ARCH := $(shell file `which docker` | awk '{print $$NF}')

ifeq ($(HOST_ARCH), arm64)
	DOCKER_PLATFORM := --platform linux/amd64
endif

GO_LINT?=
GOCMD?=CGO_ENABLED=0 go
GOCMD_TEST?=VOI_ENV=test

##@ General
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
.PHONY: lint
lint: ## Run linters.
	golangci-lint -c .ci/golangci.yaml run

.PHONY: test
test: up ## Run tests.
	CGO_ENABLED=0 go test ./... -cover
	make down

.PHONY: up
up: ## Run docker-compose up.
	docker-compose -f .docker/docker-compose.yaml up -d --remove-orphans
	@sleep 1 # pg sometimes takes a second to be reachable after it's up

.PHONY: down
down: ## Run docker-compose down.
	docker-compose -f .docker/docker-compose.yaml down -v --remove-orphans

.PHONY: mod
mod:
	$(GOCMD) get -u -t ./...
	$(GOCMD) mod tidy
	#make vendor

.PHONY: clean
clean: ## Remove compiled binaries.
	$(GOCMD) clean

