# FKey Linux Build Makefile

BINARY_NAME=fkey
VERSION?=0.1.0
RUST_TARGET=../../../core/target/release

.PHONY: all build clean deps rust-core run install uninstall

all: deps rust-core build

# Install Go dependencies
deps:
	go mod tidy

# Build Rust core library
rust-core:
	cd ../../../core && cargo build --release
	@echo "Rust core built: $(RUST_TARGET)/libgonhanh_core.so"

# Build Go binary
build:
	CGO_ENABLED=1 \
	CGO_LDFLAGS="-L$(RUST_TARGET) -lgonhanh_core" \
	go build -ldflags "-X main.Version=$(VERSION)" -o $(BINARY_NAME) .

# Run locally (for testing)
run: build
	LD_LIBRARY_PATH=$(RUST_TARGET):$$LD_LIBRARY_PATH ./$(BINARY_NAME)

# Install to system
install: build
	sudo install -Dm755 $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	sudo install -Dm644 $(RUST_TARGET)/libgonhanh_core.so /usr/local/lib/libgonhanh_core.so
	sudo ldconfig
	@echo "Installing desktop file..."
	install -Dm644 fkey.desktop ~/.local/share/applications/fkey.desktop
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

# Uninstall
uninstall:
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	sudo rm -f /usr/local/lib/libgonhanh_core.so
	rm -f ~/.local/share/applications/fkey.desktop
	sudo ldconfig

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	cd ../../../core && cargo clean

# Development: rebuild and run
dev: rust-core build run

# Check dependencies
check-deps:
	@echo "Checking dependencies..."
	@which xdotool > /dev/null || echo "WARNING: xdotool not found (install: sudo apt install xdotool)"
	@pkg-config --exists gtk+-3.0 || echo "WARNING: GTK3 dev not found (install: sudo apt install libgtk-3-dev)"
	@pkg-config --exists x11 || echo "WARNING: X11 dev not found (install: sudo apt install libx11-dev)"
	@which cargo > /dev/null || echo "WARNING: Rust not found (install: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh)"
	@echo "Done."
