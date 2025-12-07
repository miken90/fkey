.PHONY: help all test format build clean setup install release

# Auto-versioning
TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)
VER := $(subst v,,$(TAG))
NEXT := $(shell echo $(VER) | awk -F. '{print $$1"."$$2"."$$3+1}')

# Default target
.DEFAULT_GOAL := help

help: ## Show this help
	@echo "⚡ GoNhanh - Vietnamese Input Method Engine"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "\033[1;34mDevelopment:\033[0m"
	@grep -E '^(test|format|build|clean):.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[1;32m%-12s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "\033[1;33mSetup & Install:\033[0m"
	@grep -E '^(setup|install):.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[1;32m%-12s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "\033[1;31mRelease:\033[0m"
	@grep -E '^(release|all):.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[1;32m%-12s\033[0m %s\n", $$1, $$2}'

all: test build ## Run test + build

test: ## Run tests
	@cd core && cargo test

format: ## Format & lint
	@cd core && cargo fmt && cargo clippy -- -D warnings

build: format ## Build core + macos app
	@./scripts/build-core.sh
	@./scripts/build-macos.sh

clean: ## Clean build artifacts
	@cd core && cargo clean
	@rm -rf platforms/macos/build

setup: ## Setup dev environment
	@./scripts/setup.sh

install: build ## Install app to /Applications
	@cp -r platforms/macos/build/Release/GoNhanh.app /Applications/

release: ## Tag & push new version
	@echo "$(TAG) → v$(NEXT)"
	@git add -A && git commit -m "release: v$(NEXT)" --allow-empty
	@git tag v$(NEXT) && git push origin main v$(NEXT)
	@echo "→ https://github.com/khaphanspace/gonhanh.org/releases"
