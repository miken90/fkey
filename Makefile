.PHONY: help build clean test core macos setup install lint release

.DEFAULT_GOAL := help

# Version from git tag or default
VERSION ?= $(shell git describe --tags 2>/dev/null || echo "dev")

help:
	@echo "GoNhanh $(VERSION) - Makefile commands:"
	@echo ""
	@echo "Development:"
	@echo "  make setup       - Setup dev environment"
	@echo "  make test        - Run tests"
	@echo "  make lint        - Run fmt + clippy"
	@echo "  make build       - Build everything (test → core → macos)"
	@echo "  make clean       - Clean build artifacts"
	@echo ""
	@echo "Build:"
	@echo "  make core        - Build Rust core only"
	@echo "  make macos       - Build macOS app"
	@echo "  make install     - Install to /Applications"
	@echo ""
	@echo "Release:"
	@echo "  make release v=1.0.0  - Tag and push release"

build: test core macos

core:
	@./scripts/build-core.sh

macos:
	@./scripts/build-macos.sh

test:
	@echo "🧪 Running tests..."
	@cd core && cargo test

lint:
	@echo "🔍 Linting..."
	@cd core && cargo fmt -- --check && cargo clippy -- -D warnings
	@echo "✅ Lint passed!"

clean:
	@echo "🧹 Cleaning..."
	@cd core && cargo clean
	@rm -rf platforms/macos/build
	@echo "✅ Clean complete!"

setup:
	@echo "🔧 Setting up..."
	@./scripts/setup.sh

install: build
	@echo "📦 Installing GoNhanh..."
	@cp -r platforms/macos/build/Release/GoNhanh.app /Applications/
	@echo "✅ Installed to /Applications/GoNhanh.app"

# Get current and next versions
CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
CURRENT := $(shell echo $(CURRENT_TAG) | sed 's/v//')
MAJOR := $(shell echo $(CURRENT) | cut -d. -f1)
MINOR := $(shell echo $(CURRENT) | cut -d. -f2)
PATCH := $(shell echo $(CURRENT) | cut -d. -f3)

# Release commands
release-patch:
	@$(MAKE) _release NEW_V="$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1)))"

release-minor:
	@$(MAKE) _release NEW_V="$(MAJOR).$(shell echo $$(($(MINOR)+1))).0"

release-major:
	@$(MAKE) _release NEW_V="$(shell echo $$(($(MAJOR)+1))).0.0"

_release:
	@echo "🚀 Releasing v$(NEW_V)... (was $(CURRENT_TAG))"
	@git add -A
	@git commit -m "chore: release v$(NEW_V)" --allow-empty
	@git tag -a v$(NEW_V) -m "Release v$(NEW_V)"
	@git push origin main v$(NEW_V)
	@echo "✅ Released v$(NEW_V)!"
	@echo "👉 https://github.com/khaphanspace/gonhanh.org/releases/tag/v$(NEW_V)"
