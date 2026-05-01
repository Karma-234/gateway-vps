.PHONY: help certs secrets build up down restart logs clean test shell buildx

help:
	@echo "========================================"
	@echo "ISO 8583 + Fineract Gateway - Makefile"
	@echo "========================================"
	@echo ""
	@echo "Usage:"
	@echo "  make build          - Build the gateway image"
	@echo "  make up             - Start all services"
	@echo "  make down           - Stop all services"
	@echo "  make restart        - Restart gateway only"
	@echo "  make logs           - Follow gateway logs"
	@echo "  make certs          - Generate fresh TLS certificates"
	@echo "  make secrets        - Generate fresh secrets"
	@echo "  make clean          - Clean everything"
	@echo "  make test           - Test health & metrics endpoints"
	@echo ""

certs:
	@echo "Generating TLS certificates..."
	@./scripts/generate-certs.sh

secrets:
	@echo "Generating secrets..."
	@./scripts/generate-secrets.sh

DOCKER_COMPOSE = docker compose
SERVICE = gateway-vps
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

shell:
	@echo "Opening shell in gateway container..."
	$(DOCKER_COMPOSE) exec $(SERVICE) sh
build:
	@echo "Building gateway image with version: $(VERSION)"
	$(DOCKER_COMPOSE) build --build-arg VERSION=$(VERSION) --build-arg BUILD_TIME=$(BUILD_TIME) $(SERVICE)
up:
	@echo "Starting all services..."
	$(DOCKER_COMPOSE) up --build -d
down:
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down
restart:
	@echo "Restarting gateway service..."
	$(DOCKER_COMPOSE) restart $(SERVICE)
logs:
	@echo "Following gateway logs..."
	$(DOCKER_COMPOSE) logs -f $(SERVICE)
clean:
	@echo "🧹 Cleaning up..."
	docker compose down -v --rmi all
	docker system prune -f
	rm -rf certs/
	rm -rf secrets/
	@echo "Cleaned!"
buildx: 
	docker buildx build --platform linux/amd64,linux/arm64 \
    	--build-arg VERSION=$(VERSION) \
    	-t iso-fineract-gateway:$(VERSION) .

.DEFAULT_GOAL := help


