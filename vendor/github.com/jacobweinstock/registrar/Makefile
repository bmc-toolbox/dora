help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: test cover lint goimports ## run all checks
	go vet ./...

.PHONY: test
test: ## Run unit tests
	go test -v -covermode=count ./...

.PHONY: cover
cover: ## Run unit tests with coverage report
	go test -coverprofile=cover.out ./...
	go tool cover -func=cover.out
	rm -rf cover.out

.PHONY: lint
lint:  ## Run linting
	@echo be sure golangci-lint is installed: https://golangci-lint.run/usage/install/
	golangci-lint run

.PHONY: goimports
goimports: ## run goimports updating files in place
	@echo be sure goimports is installed
	goimports -w .

.PHONY: goimports-check
goimports-check: ## run goimports displaying diffs
	@echo be sure goimports is installed
	goimports -d . | (! grep .)
