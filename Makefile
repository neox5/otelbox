.PHONY: help build build-image run-container clean

# Configuration
BINARY_NAME=obsbox
IMAGE_NAME=obsbox
IMAGE_TAG=latest
CONTAINER_NAME=obsbox

# Build flags
LDFLAGS=-ldflags="-s -w"
CGO_ENABLED=0

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Compile optimized binary for local development
	CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/obsbox

build-image: ## Build container image with Podman
	podman build -t $(IMAGE_NAME):$(IMAGE_TAG) -f Containerfile .

run-container: ## Run container with config volume mount
	podman run --rm \
		--name $(CONTAINER_NAME) \
		-v $(PWD)/config.yaml:/config/config.yaml:ro \
		-p 9090:9090 \
		$(IMAGE_NAME):$(IMAGE_TAG) \
		-config /config/config.yaml

clean: ## Remove build artifacts
	rm -f $(BINARY_NAME)
	go clean
