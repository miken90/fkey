# Architecture

## Overview

GoNhanh uses a **Rust core + native UI** architecture:

```
┌─────────────────────────────────────┐
│         Platform UI Layer           │
│  ┌──────────┐      ┌──────────┐    │
│  │  macOS   │      │ Windows  │    │
│  │ SwiftUI  │      │   WPF    │    │
│  └─────┬────┘      └────┬─────┘    │
└────────┼────────────────┼──────────┘
         │    FFI (C ABI) │
┌────────▼────────────────▼──────────┐
│         Rust Core Library          │
│  ┌─────────────────────────────┐   │
│  │  Engine (Telex/VNI)         │   │
│  │  - Buffer management        │   │
│  │  - Phonology-based rules    │   │
│  │  - Unicode output           │   │
│  └─────────────────────────────┘   │
└────────────────────────────────────┘
```

## Core Library (`core/`)

### Modules

```
core/src/
├── lib.rs              # FFI exports (ime_*)
├── data/
│   ├── keys.rs         # Key code constants
│   ├── chars.rs        # Unicode character mappings
│   └── vowel.rs        # Phonology system
├── engine/
│   ├── mod.rs          # Main engine logic
│   └── buffer.rs       # Typing buffer
└── input/
    ├── mod.rs          # Method trait
    ├── telex.rs        # Telex rules
    └── vni.rs          # VNI rules
```

### FFI Interface

```c
// Initialize/cleanup
void ime_init();
void ime_clear();
void ime_free(Result* r);

// Configuration
void ime_method(uint8_t method);  // 0=Telex, 1=VNI
void ime_enabled(bool enabled);
void ime_modern(bool modern);     // true=oà, false=òa

// Key processing
Result* ime_key(uint16_t key, bool caps, bool ctrl);
```

### Result Structure

```rust
#[repr(C)]
pub struct Result {
    pub chars: [u32; 32],  // Unicode output
    pub action: u8,        // 0=None, 1=Send, 2=Restore
    pub backspace: u8,     // Chars to delete
    pub count: u8,         // Chars to send
}
```

## Engine Flow

```
1. Key press captured by platform
   ↓
2. Platform calls ime_key(key, caps, ctrl)
   ↓
3. Engine processes:
   - Add to buffer
   - Check for modifiers (tone, mark, đ)
   - Apply phonology rules for mark placement
   ↓
4. Return Result with:
   - backspace: how many chars to delete
   - chars[]: new Unicode chars to insert
   ↓
5. Platform sends backspaces + new text
```

## Platform: macOS (`platforms/macos/`)

### Components

| File | Purpose |
|------|---------|
| `App.swift` | Entry point, app lifecycle |
| `MenuBar.swift` | System tray, keyboard hooks |
| `SettingsView.swift` | SwiftUI settings UI |
| `RustBridge.swift` | FFI bridge to Rust |

### Key Hook Flow

```
CGEventTap (Accessibility)
    ↓
RustBridge.processKey()
    ↓
ime_key() → Result
    ↓
CGEvent.post() backspaces
    ↓
CGEvent.post() new chars
```

## Build Artifacts

| Platform | Output |
|----------|--------|
| macOS | `libgonhanh_core.a` (universal: arm64 + x86_64) |
| Windows | `gonhanh_core.lib` (planned) |

## Performance

- **Memory**: ~25 MB RAM
- **Binary**: ~3 MB
- **Startup**: ~200ms
- **Latency**: <1ms per keystroke

## Security

- **Memory safe**: Rust prevents buffer overflows
- **Minimal permissions**: Only keyboard access (Accessibility)
- **Offline**: No network, no telemetry
- **Open source**: Auditable code
