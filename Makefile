BINARY      := obsbox
CMD_PKG     := ./cmd/obsbox
MODULE_PATH := github.com/neox5/obsbox
DIST_DIR    := dist

# Container configuration
IMAGE_NAME     := obsbox
IMAGE_TAG      := latest
CONTAINER_NAME := obsbox

PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

# Version from git; falls back to "dev" if describe fails.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# LDFLAGS: Set version + strip debug symbols for smaller binaries
# -s: omit symbol table
# -w: omit DWARF debug info
# Result: ~40-50% size reduction
LDFLAGS := -s -w -X '$(MODULE_PATH)/internal/version.Version=$(VERSION)'

.PHONY: all build build-local clean print-version release post-release test lint help
.PHONY: build-image run-container

all: build

# ---------------------------------------------------------------------
# Multi-platform release build + checksums
# ---------------------------------------------------------------------

build: clean
	@mkdir -p "$(DIST_DIR)"
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		ext=""; \
		[ "$$GOOS" = "windows" ] && ext=".exe"; \
		out="$(DIST_DIR)/$(BINARY)-$${GOOS}-$${GOARCH}$${ext}"; \
		file=$$(basename "$$out"); \
		echo "building $$out (VERSION=$(VERSION))"; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags "$(LDFLAGS)" -o "$$out" $(CMD_PKG); \
		echo "generating $$file.sha256"; \
		( cd "$(DIST_DIR)" && sha256sum "$$file" > "$$file.sha256" ); \
	done

# ---------------------------------------------------------------------
# Local development build
# ---------------------------------------------------------------------

build-local: clean
	@mkdir -p "$(DIST_DIR)"
	@echo "building $(DIST_DIR)/$(BINARY) (VERSION=$(VERSION))"
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o "$(DIST_DIR)/$(BINARY)" $(CMD_PKG)

# ---------------------------------------------------------------------
# Container image build
# ---------------------------------------------------------------------

build-image: ## Build container image with Podman
	podman build -t $(IMAGE_NAME):$(IMAGE_TAG) -f Containerfile .

run-container: ## Run container with config volume mount
	podman run --rm \
		--name $(CONTAINER_NAME) \
		-v $(PWD)/config.yaml:/config/config.yaml:ro \
		-p 9090:9090 \
		$(IMAGE_NAME):$(IMAGE_TAG) \
		-config /config/config.yaml

# ---------------------------------------------------------------------
# Full release verification (tag, tests, binaries, checksums)
# ---------------------------------------------------------------------

release:
	@./scripts/release.sh

# ---------------------------------------------------------------------
# Post-release verification against GitHub latest
# ---------------------------------------------------------------------

post-release:
	@./scripts/post-release.sh

# ---------------------------------------------------------------------

test:
	go test -v ./...

lint:
	go vet ./...
	go fmt ./...

print-version:
	@echo $(VERSION)

clean:
	@echo "removing $(DIST_DIR)/"
	@rm -rf "$(DIST_DIR)"

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'
