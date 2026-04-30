# Makefile for GKE MCP Server
# Provides convenient shortcuts for common development tasks

.PHONY: help build test clean presubmit

# Default target - show help
.DEFAULT_GOAL := help

help: ## Display available commands
	@echo "GKE MCP Server - Available Commands"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Build the project (TS and UI)
	@echo "Building..."
	npm run build
	@echo "✓ Built"

test: ## Run tests
	@echo "Running tests..."
	npm run test

clean: ## Remove build artifacts
	@echo "Cleaning up..."
	@rm -rf dist/
	@rm -rf ui/dist/
	@echo "✓ Cleaned"

presubmit: ## Run all presubmit checks
	@echo "Running presubmit checks..."
	@npm run test
	@echo "✓ All presubmit checks passed"
