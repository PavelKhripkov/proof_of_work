default: help

help: ## Help.
	@echo "Use \`make <target>\` where <target> one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
	    awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'


test: ## Run tests and linters.
	go test ./...
	./bin/golangci-lint run

server-up: ## Run server.
server-down: ## Stop server.

client-up: ## Run client.
client-down: ## Stop client.
