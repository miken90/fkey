#!/bin/bash
set -e

echo "Installing FKey dependencies..."

sudo apt-get update
sudo apt-get install -y \
  build-essential \
  libayatana-appindicator3-dev \
  libx11-dev \
  libxtst-dev \
  libxinerama-dev \
  libx11-xcb-dev \
  libxkbcommon-dev \
  libxkbcommon-x11-dev \
  xdotool \
  gedit

echo "Building Rust core..."
cd /workspaces/fkey/core
cargo build --release

echo "Building Go app..."
cd /workspaces/fkey/platforms/linux
export CGO_LDFLAGS="-L../../core/target/release -lgonhanh_core"
go mod tidy
go build -o fkey .

echo ""
echo "=== Setup Complete ==="
echo "To test FKey:"
echo "1. Open port 6080 in browser (noVNC desktop)"
echo "2. Password: fkey"
echo "3. Open terminal in desktop"
echo "4. Run: cd /workspaces/fkey/platforms/linux && LD_LIBRARY_PATH=../../core/target/release ./fkey"
echo "5. Open gedit and test typing Vietnamese"
