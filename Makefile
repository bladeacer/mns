BINARY   ?= mns
VERSION  ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.1.0")
LDFLAGS  := -s -w -X github.com/bladeacer/mns/config.AppVersion=$(VERSION)

.DEFAULT_GOAL := help

.PHONY: help build test lint gowatch snapshot tag fmt

help: ## Show this help
	@printf "\nUsage: make <target>\n\n"
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf "\n"

build: ## Build the mns binary
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

test: ## Run all tests with coverage and badge
	go test -coverpkg=./... -coverprofile=coverage.out ./tests/... -count=1
	@go tool cover -func=coverage.out
	@go-test-coverage -p coverage.out -b coverage.svg 2>/dev/null || true

cover: ## Run tests with code coverage (alias for test)
	$(MAKE) test

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Formats source code
	rg --no-ignore --files --type go | xargs gofmt -w -s

gowatch: ## Start gowatch for hot-reload development
	gowatch

snapshot: ## Test goreleaser locally (builds all platforms)
	goreleaser release --snapshot --clean

tag: ## Bump version in config, commit, create and push an annotated git tag
	@CURRENT=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	MAJOR=$$(echo "$$CURRENT" | sed 's/^v//' | cut -d. -f1); \
	MINOR=$$(echo "$$CURRENT" | sed 's/^v//' | cut -d. -f2); \
	SUGGEST="v$$MAJOR.$$(($$MINOR + 1)).0"; \
	read -p "Enter version [$$SUGGEST]: " TAG; \
	TAG=$${TAG:-$$SUGGEST}; \
	VER=$$(echo "$$TAG" | sed 's/^v//'); \
	if git rev-parse "$$TAG" >/dev/null 2>&1; then \
		echo "Tag $$TAG already exists, pushing..."; \
	else \
		sed -i "s/var AppVersion = \".*\"/var AppVersion = \"$$VER\"/" config/config.go; \
		git add config/config.go; \
		git commit -m "chore: bump version to $$TAG"; \
		git tag -a "$$TAG" -m "Release $$TAG" && echo "Created tag $$TAG."; \
	fi; \
	git push origin "$$TAG"
