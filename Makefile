.DEFAULT_GOAL := help

# ==============================================================================
# Variables
# ==============================================================================
CR           ?= icn.vultrcr.com/homincr1
IMAGE_VER    ?= latest
PROGRAM_VER  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
PLATFORM     ?= linux/amd64
DEPLOYMENT   ?= ingress
CONFIG       ?= conf/links.yaml
STATIC_DIR   ?= frontend/dist
PORT         ?= 8080

IMAGE_TAG_BACKEND  := $(CR)/ingress-backend:$(IMAGE_VER)
IMAGE_TAG_FRONTEND := $(CR)/ingress-frontend:$(IMAGE_VER)

# ==============================================================================
# Help
# ==============================================================================
.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target] [VARIABLE=value]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Variables:"
	@echo "  CR          Container Registry (default: $(CR))"
	@echo "  IMAGE_VER   Docker image tag (default: $(IMAGE_VER))"
	@echo "  PROGRAM_VER Application build version (default: $(PROGRAM_VER))"
	@echo "  DEPLOYMENT  Kubernetes deployment name (default: $(DEPLOYMENT))"
	@echo "  CONFIG      Links config file (default: $(CONFIG))"
	@echo "  PORT        HTTP port for local server (default: $(PORT))"

# ==============================================================================
# Development & Testing
# ==============================================================================
.PHONY: dev-fe
dev-fe: ## Run frontend dev server (Vite)
	cd frontend && npm run dev

.PHONY: dev-be
dev-be: ## Run backend server with links config and built frontend
	go run . -static $(STATIC_DIR) -c $(CONFIG) -http $(PORT)

.PHONY: run
run: ## Run backend server with default settings
	go run . -c $(CONFIG) -http $(PORT)

.PHONY: test
test: ## Run Go unit tests
	go test -v ./...

.PHONY: vet
vet: ## Run Go vet
	go vet ./...

# ==============================================================================
# Build & Assets
# ==============================================================================
.PHONY: build-fe
build-fe: ## Build frontend assets into frontend/dist
	cd frontend && npm ci && npm run build

.PHONY: build-be
build-be: ## Build backend Go binary
	go build -ldflags "-X main.programVer=$(PROGRAM_VER)" -o app .

.PHONY: build
build: build-fe build-be ## Build both frontend and backend binary

.PHONY: model
model: ## Convert 3D Gopher STL mesh to quantized binary
	python3 scripts/stl2qbin.py asset/go_gopher_low_0.02.stl asset/go_gopher_low_0.02.bin

# ==============================================================================
# Docker & Deployment (Vultr / Kubernetes)
# ==============================================================================
.PHONY: docker-be
docker-be: ## Build backend Docker image
	docker buildx build --platform $(PLATFORM) --build-arg=PROGRAM_VER=$(PROGRAM_VER) -t $(IMAGE_TAG_BACKEND) .

.PHONY: docker-fe
docker-fe: ## Build frontend Docker image
	docker buildx build --platform $(PLATFORM) -f frontend/Dockerfile -t $(IMAGE_TAG_FRONTEND) .

.PHONY: docker
docker: docker-be docker-fe ## Build both Docker images

.PHONY: push-be
push-be: docker-be ## Build and push backend Docker image
	docker push $(IMAGE_TAG_BACKEND)

.PHONY: push-fe
push-fe: docker-fe ## Build and push frontend Docker image
	docker push $(IMAGE_TAG_FRONTEND)

.PHONY: push
push: push-be push-fe ## Build and push both backend and frontend images

.PHONY: restart
restart: ## Restart k8s deployment and wait for rollout
	kubectl rollout restart deployment $(DEPLOYMENT)
	kubectl rollout status deployment $(DEPLOYMENT)

.PHONY: deploy
deploy: push restart ## Build, push all images, and restart deployment

.PHONY: deploy-be
deploy-be: push-be restart ## Build, push backend image, and restart deployment

.PHONY: deploy-fe
deploy-fe: push-fe restart ## Build, push frontend image, and restart deployment

.PHONY: status
status: ## Check k8s deployment status and pods
	kubectl rollout status deployment $(DEPLOYMENT)
	kubectl get pods -l app=$(DEPLOYMENT)

# Aliases
.PHONY: docker-backend docker-frontend push-backend push-frontend deploy-backend deploy-frontend build-backend build-frontend
docker-backend: docker-be
docker-frontend: docker-fe
push-backend: push-be
push-frontend: push-fe
deploy-backend: deploy-be
deploy-frontend: deploy-fe
build-backend: build-be
build-frontend: build-fe

# ==============================================================================
# Verification & Maintenance
# ==============================================================================
.PHONY: verify
verify: ## Verify local running server endpoints (/social, /llms.txt, markdown)
	@echo "--- Checking /social ---"
	@curl -s localhost:$(PORT)/social | grep -E '<title>|link-card' || echo "Failed to fetch /social"
	@echo "\n--- Checking /llms.txt ---"
	@curl -s localhost:$(PORT)/llms.txt | head -n 10 || echo "Failed to fetch /llms.txt"
	@echo "\n--- Checking Markdown Content Negotiation ---"
	@curl -s -H 'Accept: text/markdown' localhost:$(PORT)/ | head -n 10 || echo "Failed to fetch markdown"

.PHONY: clean
clean: ## Remove build artifacts
	rm -f app
	rm -rf frontend/dist
