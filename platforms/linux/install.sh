#!/bin/bash
# FKey Linux Installer

set -e

INSTALL_DIR="/usr/local/bin"
LIB_DIR="/usr/local/lib"
DESKTOP_DIR="$HOME/.local/share/applications"

echo "=== FKey Vietnamese IME Installer ==="
echo ""

# Check for xdotool
if ! command -v xdotool &> /dev/null; then
    echo "ERROR: xdotool is required but not installed."
    echo "Install with: sudo apt install xdotool"
    exit 1
fi

# Install binary
echo "Installing fkey to $INSTALL_DIR..."
sudo install -Dm755 fkey "$INSTALL_DIR/fkey"

# Install library
echo "Installing libgonhanh_core.so to $LIB_DIR..."
sudo install -Dm644 libgonhanh_core.so "$LIB_DIR/libgonhanh_core.so"
sudo ldconfig

# Install desktop file
echo "Installing desktop file..."
mkdir -p "$DESKTOP_DIR"
install -Dm644 fkey.desktop "$DESKTOP_DIR/fkey.desktop"

echo ""
echo "=== Installation complete! ==="
echo ""
echo "To start FKey:"
echo "  1. Run 'fkey' from terminal, or"
echo "  2. Find 'FKey' in your applications menu"
echo ""
echo "The system tray icon will appear when running."
echo "Use Ctrl+Space to toggle Vietnamese input on/off."
echo ""
echo "To uninstall:"
echo "  sudo rm /usr/local/bin/fkey /usr/local/lib/libgonhanh_core.so"
echo "  rm ~/.local/share/applications/fkey.desktop"
