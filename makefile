APP=message-system
APP_EXECUTABLE=${APP}
DOCKER_IMAGE=messaging-system:latest

serve-http:
	go run main.go serve-http

migrate:
	go run main.go migrate

test: ## Run tests
	go test -v ./tests/...

deps: ## Install dependencies
	go mod download
	go mod tidy

docker: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-up: ## Start services with Docker Compose
	docker-compose up -d

docker-down: ## Stop Docker Compose services
	docker-compose down -v

docker-logs: ## Show Docker Compose logs
	docker-compose logs -f

lint: ## Run linter
	golangci-lint run

swagger: ## Generate Swagger documentation
	swag init