# FKey Linux

Vietnamese Input Method for Linux (X11).

## Requirements

### Build Dependencies

```bash
# Ubuntu/Debian
sudo apt install build-essential libgtk-3-dev libx11-dev xdotool

# Fedora
sudo dnf install gtk3-devel libX11-devel xdotool

# Arch
sudo pacman -S gtk3 libx11 xdotool
```

### Runtime Dependencies

- **xdotool** - Required for Unicode text injection
- **X11** - Wayland not yet supported

## Build

```bash
# Build everything (Rust core + Go app)
make all

# Or step by step:
make rust-core  # Build Rust IME engine
make build      # Build Go application
```

## Run

```bash
# Run locally (for development)
make run

# Or directly:
LD_LIBRARY_PATH=../../../core/target/release ./fkey
```

## Install

```bash
# Install to /usr/local/bin
make install
```

## Usage

1. **Toggle IME**: `Ctrl+Space` (default)
2. **Right-click tray icon** for menu
3. **Left-click tray icon** to toggle on/off

### Input Methods

- **Telex**: `aa` → `â`, `aw` → `ă`, `ow` → `ơ`, `s/f/r/x/j` for tones
- **VNI**: `a6` → `â`, `a8` → `ă`, `o7` → `ơ`, `1-5` for tones

## Configuration

Config file: `~/.config/fkey/config.toml`

```toml
enabled = true
input_method = 0  # 0=Telex, 1=VNI
modern_tone = true
esc_restore = true
toggle_hotkey = "Ctrl+Space"
```

## Known Limitations

1. **X11 only** - Wayland not supported yet
2. **xdotool required** - For reliable Unicode injection
3. **Some apps may not work** - Apps with custom input handling

## Roadmap

- [ ] Wayland support via wlroots protocols
- [ ] IBus integration for better compatibility
- [ ] Custom hotkey configuration UI
- [ ] App-specific profiles
