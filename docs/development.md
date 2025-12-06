# Development Guide

## Prerequisites

### Required
- **Rust** 1.70+ ([Install](https://rustup.rs))
- **macOS** 13+ (for macOS development)
- **Xcode** 15+ (for macOS development)

### Optional
- **Visual Studio** 2022+ (for Windows development)

## Setup

```bash
# Clone repository
git clone https://github.com/yourusername/gonhanh.org
cd gonhanh.org

# Run setup script
./scripts/setup.sh

# Install Rust targets
rustup target add aarch64-apple-darwin
rustup target add x86_64-apple-darwin
```

## Building

### Build Rust Core

```bash
./scripts/build-core.sh
```

This creates a universal binary at `platforms/macos/libgonhanh_core.a`

### Build macOS App

**Option 1: Using Xcode**

1. Open Xcode
2. Create new macOS App project:
   - Product Name: `GoNhanh`
   - Organization ID: `org.gonhanh`
   - Interface: `SwiftUI`
   - Language: `Swift`
   - Location: `platforms/macos/`

3. Add Swift files:
   - Drag all `.swift` files from `platforms/macos/GoNhanh/` to project

4. Link Rust library:
   - Select project in navigator
   - Go to "Build Phases" → "Link Binary With Libraries"
   - Click "+" → Add `libgonhanh_core.a`

5. Update Info.plist:
   - Copy from `platforms/macos/GoNhanh/Info.plist`

6. Build: `Cmd + B`

**Option 2: Using xcodebuild**

```bash
./scripts/build-macos.sh
```

## Testing

### Test Rust Core

```bash
cd core
cargo test
```

### Test Vietnamese Engine

```bash
cd core
cargo test engine
```

## Project Structure

```
gonhanh.org/
├── core/              # Rust core library
│   ├── src/
│   │   ├── lib.rs    # FFI exports
│   │   ├── engine.rs # Vietnamese conversion
│   │   ├── keyboard.rs
│   │   └── config.rs
│   └── tests/
│
├── platforms/
│   └── macos/        # macOS native app
│       └── GoNhanh/
│           ├── App.swift
│           ├── MenuBar.swift
│           ├── SettingsView.swift
│           └── RustBridge.swift
│
└── scripts/          # Build scripts
```

## Debugging

### macOS

1. Open Xcode project
2. Set breakpoints in Swift code
3. Run: `Cmd + R`

### Rust Core

```bash
cd core
RUST_LOG=debug cargo test -- --nocapture
```

## Common Issues

### "Library not found"

Make sure you built the Rust core:
```bash
./scripts/build-core.sh
```

### "Undefined symbols"

Check that `libgonhanh_core.a` is linked in Xcode:
- Build Phases → Link Binary With Libraries

### Keyboard hook not working

Grant Accessibility permissions:
- System Settings → Privacy & Security → Accessibility
- Add GoNhanh

## Release Build

```bash
# Build optimized binary
cd core
cargo build --release --target aarch64-apple-darwin
cargo build --release --target x86_64-apple-darwin

# Create universal binary
./scripts/build-core.sh

# Build macOS app in release mode
cd platforms/macos
xcodebuild -scheme GoNhanh -configuration Release
```

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md)
