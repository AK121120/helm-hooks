.PHONY: build test test-e2e clean install plugin-install

GO := $(shell command -v go 2>/dev/null || echo /usr/local/go/bin/go)

# Build the binary
build:
	$(GO) build -o helm-hooks ./cmd/helm-hooks

# Run unit tests
test:
	$(GO) test ./... -v

# Run E2E regression tests
test-e2e: build
	./scripts/test-e2e.sh

# Run all tests
test-all: test test-e2e

# Clean build artifacts
clean:
	rm -f helm-hooks
	rm -f helm-hooks-*

# Install as Helm 4 plugin (development mode)
plugin-install: build
	helm plugin uninstall helm-hooks 2>/dev/null || true
	helm plugin install .

# Uninstall plugin
plugin-uninstall:
	helm plugin uninstall helm-hooks

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GO) build -o helm-hooks-linux-amd64 ./cmd/helm-hooks
	GOOS=linux GOARCH=arm64 $(GO) build -o helm-hooks-linux-arm64 ./cmd/helm-hooks
	GOOS=darwin GOARCH=amd64 $(GO) build -o helm-hooks-darwin-amd64 ./cmd/helm-hooks
	GOOS=darwin GOARCH=arm64 $(GO) build -o helm-hooks-darwin-arm64 ./cmd/helm-hooks
	GOOS=windows GOARCH=amd64 $(GO) build -o helm-hooks-windows-amd64.exe ./cmd/helm-hooks

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  test           - Run unit tests"
	@echo "  test-e2e       - Run E2E regression tests"
	@echo "  test-all       - Run all tests"
	@echo "  clean          - Remove build artifacts"
	@echo "  plugin-install - Install as Helm 4 plugin"
	@echo "  build-all      - Build for all platforms"
