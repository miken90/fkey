# FKey Linux - Integration Testing Checklist

## Prerequisites

```bash
# Install dependencies
sudo apt update && sudo apt install -y \
  build-essential \
  libayatana-appindicator3-dev \
  libx11-dev \
  xdotool

# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env

# Install Go 1.22+
# https://go.dev/dl/
```

## Build Steps

```bash
# Clone and checkout
git clone https://github.com/miken90/fkey.git
cd fkey
git checkout feature/linux-port

# Build Rust core
cd core
cargo build --release
cd ..

# Build Linux app
cd platforms/linux
export CGO_LDFLAGS="-L../../core/target/release -lgonhanh_core"
go build -o fkey .

# Run (requires X11)
export LD_LIBRARY_PATH="../../core/target/release:$LD_LIBRARY_PATH"
./fkey
```

## Test Checklist

### Basic Functionality
- [ ] Binary starts without error
- [ ] System tray icon appears
- [ ] Tray tooltip shows "FKey - Tiếng Việt (Bật) - Telex"

### Toggle Hotkey
- [ ] Press Ctrl+Space → log shows "IME toggled: false"
- [ ] Press Ctrl+Space again → log shows "IME toggled: true"
- [ ] Tray menu updates to reflect state

### Vietnamese Input (Telex)
Open gedit or any text editor:
- [ ] Type "vieetj" → "việt"
- [ ] Type "Vieejt Nam" → "Việt Nam"
- [ ] Type "xin chaof" → "xin chào"
- [ ] Type "hoaf binh2" → "hoà bình" (modern tone)
- [ ] Type "tôi" → "tôi" (with circumflex)

### VNI Mode
- [ ] Switch to VNI via tray menu
- [ ] Type "vie6t5" → "việt"
- [ ] Type "xin cha2o" → "xin chào"

### Special Keys
- [ ] ESC restores raw text (if enabled)
- [ ] Backspace works correctly mid-word
- [ ] Space commits word

### Tray Menu
- [ ] "✓ Tiếng Việt" / "Bật Tiếng Việt" toggles
- [ ] Kiểu gõ → Telex/VNI switches method
- [ ] "Bỏ dấu kiểu mới" checkbox works
- [ ] "ESC khôi phục ký tự gốc" checkbox works
- [ ] "Thoát" exits cleanly

### Config Persistence
- [ ] Change settings via tray
- [ ] Quit and restart
- [ ] Settings are preserved
- [ ] Check ~/.config/fkey/config.toml exists

## Known Limitations (MVP)

- X11 only (no Wayland support)
- US QWERTY keyboard layout assumed
- Hotkey fixed at Ctrl+Space
- No autostart setup yet

## Reporting Issues

If tests fail, note:
1. Ubuntu version
2. Desktop environment (GNOME, KDE, etc.)
3. Error messages from terminal
4. Steps to reproduce
