help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[32m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build: ## Build image
	docker compose build

up: ## Up containers
	docker compose up -d --remove-orphans

down: ## Down containers
	docker compose down --remove-orphans

logs: ## Show logs
	docker compose logs

logsf: ## Follow logs
	docker compose logs -f

vet-server: ## Run go vet server
	docker compose run --rm --no-deps server go vet ./...

test-server: ## Run go test server
	docker compose run --rm server go test -v ./...

test-race-server: ## Run go race test server
	docker compose run --rm --no-deps server go test -v -race ./...

vet-agent: ## Run go vet agent
	docker compose run --rm --no-deps agent go vet ./...

test-agent: ## Run go test agent
	docker compose run --rm --no-deps agent go test -v ./...

test-race-agent: ## Run go race test agent
	docker compose run --rm --no-deps agent go test -v -race ./...

cpu-pprof-server: ## Capture CPU pprof server
	docker compose kill -s SIGUSR1 server

cpu-pprof-agent: ## Capture CPU pprof agent
	docker compose kill -s SIGUSR1 agent

mem-pprof-server: ## Capture memory pprof server
	docker compose kill -s SIGUSR2 server

mem-pprof-agent: ## Capture memory pprof agent
	docker compose kill -s SIGUSR2 agent