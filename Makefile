.PHONY: all test lint build clean setup install release

# Auto-versioning
TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)
VER := $(subst v,,$(TAG))
NEXT := $(shell echo $(VER) | awk -F. '{print $$1"."$$2"."$$3+1}')

all: test build

test:
	@cd core && cargo test

lint:
	@cd core && cargo fmt -- --check && cargo clippy -- -D warnings

build:
	@./scripts/build-core.sh
	@./scripts/build-macos.sh

clean:
	@cd core && cargo clean
	@rm -rf platforms/macos/build

setup:
	@./scripts/setup.sh

install: build
	@cp -r platforms/macos/build/Release/GoNhanh.app /Applications/

release:
	@echo "$(TAG) → v$(NEXT)"
	@git add -A && git commit -m "release: v$(NEXT)" --allow-empty
	@git tag v$(NEXT) && git push origin main v$(NEXT)
	@echo "→ https://github.com/khaphanspace/gonhanh.org/releases"
