#!/bin/bash
# FKey Linux - One-click Install & Test Script
# Usage: curl -fsSL https://raw.githubusercontent.com/miken90/fkey/main/platforms/linux/install-test.sh | bash

set -e

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║       FKey - Vietnamese Input Method for Linux               ║"
echo "║                   Installation Script                        ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Check if running on Linux
if [[ "$(uname)" != "Linux" ]]; then
    echo "❌ This script only runs on Linux"
    exit 1
fi

# Check for X11
if [[ -z "$DISPLAY" ]]; then
    echo "❌ X11 display not found. Please run in a graphical environment."
    exit 1
fi

echo "📦 Installing dependencies..."
sudo apt-get update -qq
sudo apt-get install -y -qq xdotool wget tar

echo ""
echo "⬇️  Downloading FKey v0.1.0..."
cd /tmp
rm -rf fkey-install
mkdir -p fkey-install
cd fkey-install

wget -q --show-progress https://github.com/miken90/fkey/releases/download/v0.1.0-linux/FKey-0.1.0-linux-x86_64.tar.gz

echo ""
echo "📂 Extracting..."
tar -xzf FKey-0.1.0-linux-x86_64.tar.gz
cd FKey-0.1.0-linux-x86_64

echo ""
echo "🔧 Installing..."
./install.sh

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    ✅ Installation Complete!                 ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "🚀 Starting FKey..."
echo ""

# Run FKey
fkey &
FKEY_PID=$!

sleep 2

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                     📝 TEST INSTRUCTIONS                     ║"
echo "╠══════════════════════════════════════════════════════════════╣"
echo "║                                                              ║"
echo "║  1. Open any text editor (gedit, kate, etc.)                 ║"
echo "║                                                              ║"
echo "║  2. Type these words and check results:                      ║"
echo "║                                                              ║"
echo "║     vieetj    →  should become: việt                        ║"
echo "║     xin chaof →  should become: xin chào                    ║"
echo "║     hoaf binh2→  should become: hoà bình                    ║"
echo "║                                                              ║"
echo "║  3. Press Ctrl+Space to toggle IME on/off                    ║"
echo "║                                                              ║"
echo "║  4. Right-click tray icon for settings                       ║"
echo "║                                                              ║"
echo "╠══════════════════════════════════════════════════════════════╣"
echo "║  Press Enter when done testing to see results prompt...     ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

read -p ""

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    📊 TEST RESULTS                           ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "Please answer the following questions (y/n):"
echo ""

read -p "1. Did the tray icon appear? (y/n): " Q1
read -p "2. Did Ctrl+Space toggle work? (y/n): " Q2
read -p "3. Did 'vieetj' become 'việt'? (y/n): " Q3
read -p "4. Did 'xin chaof' become 'xin chào'? (y/n): " Q4
read -p "5. Any errors or issues? (describe or 'no'): " Q5

echo ""
echo "══════════════════════════════════════════════════════════════"
echo "TEST REPORT"
echo "══════════════════════════════════════════════════════════════"
echo "Date: $(date)"
echo "OS: $(lsb_release -d 2>/dev/null | cut -f2 || cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
echo "Desktop: $XDG_CURRENT_DESKTOP"
echo ""
echo "Results:"
echo "  Tray icon: $Q1"
echo "  Toggle hotkey: $Q2"
echo "  Vietnamese input (vieetj): $Q3"
echo "  Vietnamese input (chaof): $Q4"
echo "  Issues: $Q5"
echo "══════════════════════════════════════════════════════════════"
echo ""
echo "Please send this report to: https://github.com/miken90/fkey/issues"
echo ""

# Cleanup
kill $FKEY_PID 2>/dev/null || true

echo "To run FKey again: fkey"
echo "To uninstall: /tmp/fkey-install/FKey-0.1.0-linux-x86_64/uninstall.sh"
