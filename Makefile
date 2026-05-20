BINARY   ?= mns
VERSION  ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.1.0")
LDFLAGS  := -s -w -X github.com/bladeacer/mmsync/config.AppVersion=$(VERSION)

.DEFAULT_GOAL := help

.PHONY: help build test lint gowatch snapshot tag

help: ## Show this help
	@printf "\nUsage: make <target>\n\n"
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf "\n"

build: ## Build the mns binary
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

test: ## Run all tests
	go test ./... -v -count=1

lint: ## Run golangci-lint
	golangci-lint run ./...

gowatch: ## Start gowatch for hot-reload development
	gowatch

snapshot: ## Test goreleaser locally (builds all platforms)
	goreleaser release --snapshot --clean

tag: ## Create an annotated git tag for a new release
	@read -p "Enter version (e.g. v0.2.0): " TAG; \
	git tag -a "$$TAG" -m "Release $$TAG" && \
	echo "Created tag $$TAG. Push with: git push origin $$TAG"
