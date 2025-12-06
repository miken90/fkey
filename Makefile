.PHONY: help build clean test core macos setup install

.DEFAULT_GOAL := help

help:
	@echo "GoNhanh - Makefile commands:"
	@echo "  make build       - Build everything (core + macOS app)"
	@echo "  make core        - Build Rust core only"
	@echo "  make macos       - Build macOS app"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make setup       - Setup development environment"
	@echo "  make install     - Install the app"

build: test core macos

core:
	@./scripts/build-core.sh

macos:
	@./scripts/build-macos.sh

test:
	@echo "🧪 Running tests..."
	@if [ -f "$$HOME/.cargo/env" ]; then source "$$HOME/.cargo/env"; fi && cd core && cargo test

clean:
	@echo "🧹 Cleaning..."
	cd core && cargo clean
	rm -rf platforms/macos/build
	@echo "✅ Clean complete!"

setup:
	@echo "🔧 Setting up..."
	./scripts/setup.sh

install: build
	@echo "📦 Installing GoNhanh..."
	@echo "Please drag platforms/macos/build/Release/GoNhanh.app to /Applications"
