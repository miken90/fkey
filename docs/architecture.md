# Architecture

## Overview

GoNhanh uses a **hybrid architecture** with a shared Rust core and platform-native UIs.

```
┌─────────────────────────────────────┐
│        Platform UI Layer            │
│  ┌──────────┐      ┌──────────┐    │
│  │  macOS   │      │ Windows  │    │
│  │ SwiftUI  │      │   WPF    │    │
│  └─────┬────┘      └────┬─────┘    │
└────────┼──────────────────┼─────────┘
         │                  │
         │  FFI (C ABI)     │
         │                  │
┌────────▼──────────────────▼─────────┐
│       Rust Core Library             │
│  ┌─────────┐  ┌──────────┐         │
│  │ Engine  │  │ Keyboard │         │
│  └─────────┘  └──────────┘         │
└─────────────────────────────────────┘
```

## Components

### Rust Core (`core/`)

**Responsibilities:**
- Vietnamese text conversion (Telex/VNI)
- Keyboard event listening
- Configuration management
- Business logic

**Key modules:**
- `engine.rs` - Vietnamese conversion algorithms
- `keyboard.rs` - Cross-platform keyboard hooks
- `config.rs` - Configuration storage
- `lib.rs` - FFI exports

**Build artifacts:**
- macOS: `libgonhanh_core.a` (static library)
- Windows: `gonhanh_core.lib` (static library)

### macOS Platform (`platforms/macos/`)

**Technology:** SwiftUI + Cocoa

**Components:**
- `App.swift` - Application entry point
- `MenuBar.swift` - System tray controller
- `SettingsView.swift` - Settings window (SwiftUI)
- `RustBridge.swift` - FFI bridge to Rust

**Features:**
- Native macOS UI
- Menu bar integration
- System-native settings window
- Auto-start support

### Windows Platform (`platforms/windows/`)

**Technology:** WPF/WinUI3 (planned)

**Components:**
- `App.xaml` - Application entry
- `MainWindow.xaml` - Settings window
- `RustInterop.cs` - P/Invoke bridge

## Data Flow

### Keyboard Input Processing

```
1. User types key
   ↓
2. Rust keyboard hook captures event
   ↓
3. Engine processes input (Telex/VNI)
   ↓
4. Send Vietnamese character to OS
   ↓
5. OS displays character
```

### Configuration

```
1. User changes settings in UI
   ↓
2. UI calls Rust FFI
   ↓
3. Rust saves to TOML file
   ↓
4. Changes applied immediately
```

## Performance Optimizations

- **Static linking**: No runtime dependencies
- **Zero-copy FFI**: Minimal overhead between Rust and native
- **Lazy loading**: Settings window only created when needed
- **Release optimizations**: LTO, size optimization

## Security

- **Memory safety**: Rust prevents memory bugs
- **Minimal permissions**: Only keyboard access needed
- **No network**: Fully offline
- **Open source**: Auditable code
