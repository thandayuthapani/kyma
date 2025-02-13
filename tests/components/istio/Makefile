# Inspired by:
# https://betterprogramming.pub/my-ultimate-makefile-for-golang-projects-fcc8ca20c9bb
SHELL=/bin/bash

GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_NAME=example
VERSION?=0.0.0

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all test provision-k3d kyma-deploy test-k3d

all: help

## Test:

test: ## Run the tests of the project
ifeq ($(EXPORT_RESULT), true)
	GO111MODULE=off go get -u github.com/jstemmer/go-junit-report
	$(eval OUTPUT_OPTIONS = | tee /dev/tty | go-junit-report -set-exit-code > junit-report.xml)
endif
	$(GOTEST) -v -race ./... $(OUTPUT_OPTIONS)

test-k3d: provision-k3d kyma-deploy test ## Run tests on k3d cluster

provision-k3d: ## Provision k3d cluster
	kyma provision k3d --port 8080:80@loadbalancer --ci

kyma-deploy: ## Deploy Kyma cluster-essentials and istio components on k8s cluster KUBECONFIG var points to
	kyma deploy -s main --component cluster-essentials --component istio --ci

test-bdd: ## Run BDD-test using godog
	godog run

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)