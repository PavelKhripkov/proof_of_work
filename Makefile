default: help

help: ## Help.
	@echo "Use \`make <target>\` where <target> one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
	    awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

init: ## Set up required tools.
	rm -rf bin/*
	cd tools && go generate -x -tags=tools

test: ## Run tests and linters.
	go test ./...
	./bin/golangci-lint run

server-up: ## Run server.
	docker compose -f docker-compose-server.yml up --build
server-down: ## Stop server.
	docker compose -f docker-compose-server.yml down

client-up: ## Run client.
	docker compose -f docker-compose-client.yml up --build
client-down: ## Stop client.
	docker compose -f docker-compose-client.yml down
