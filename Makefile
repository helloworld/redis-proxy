.PHONY: help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

build: ## Build the container. 
	docker-compose build

run: ## Run container. See README for options
	docker-compose up -d

stop: ## Stop docker container
	docker-compose down -v

test: ## Run Tests. Runs within docker container
	docker-compose run go go test -v ./...
	docker-compose down
