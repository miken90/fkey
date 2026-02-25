# System Architecture

## Overview

FKey is a Vietnamese Input Method Editor (IME) for Windows with a two-layer architecture: a **Rust core engine** that handles Vietnamese text processing, and a **Go/Wails v3 platform layer** that integrates with Windows via Win32 APIs. The layers communicate through a C-ABI FFI bridge—no CGo required.

---

## Architecture Layers

```
┌─────────────────────────────────────────────────────┐
│                   WebView2 Frontend                  │
│                 (HTML / CSS / JS)                     │
├─────────────────────────────────────────────────────┤
│              Wails v3 Bindings (bindings.go)         │
├──────────┬──────────────────────────────┬───────────┤
│ Services │         Core (Go)            │  main.go  │
│----------│------------------------------│-----------|
│ settings │  ime_loop    keyboard_hook   │  tray UI  │
│ updater  │  text_sender app_detector    │  menus    │
│ format   │  smart_paste coalescer       │  events   │
│          │  elevation   clipboard       │           │
├──────────┴──────────┬───────────────────┴───────────┤
│                     │  FFI Bridge (bridge.go)        │
│                     │  syscall.LoadDLL → ime_*()     │
├─────────────────────┴───────────────────────────────┤
│              Rust Core Engine (gonhanh_core.dll)      │
│  ┌──────────┬──────────┬──────────┬──────────┐      │
│  │  engine/  │  data/   │  input/  │  utils   │      │
│  │  buffer   │  chars   │  telex   │          │      │
│  │  shortcut │  vowel   │  vni     │          │      │
│  │  syllable │  keys    │          │          │      │
│  │  transform│  english │          │          │      │
│  │  validate │  dict    │          │          │      │
│  └──────────┴──────────┴──────────┴──────────┘      │
└─────────────────────────────────────────────────────┘
```

---

## Rust Core Engine (`core/`)

The engine is a pure Rust library with **zero runtime dependencies**. It compiles to a DLL (`cdylib`) and exposes a C-ABI FFI interface.

### Engine Architecture

**Validation-first, pattern-based** approach:

1. Keystroke enters buffer
2. Engine scans entire buffer for Vietnamese patterns (not case-by-case)
3. Validates against Vietnamese spelling rules before applying transforms
4. Applies diacritic marks and tones using longest-match-first strategy
5. Returns `Result` struct with action (None/Send/Restore), backspaces, and replacement chars

### Key Modules

| Module | File | Purpose |
|--------|------|---------|
| **Engine** | `engine/mod.rs` | Main `Engine` struct, `on_key()` processing pipeline |
| **Buffer** | `engine/buffer.rs` | Fixed-size keystroke buffer (`MAX=256`), no heap alloc per keystroke |
| **Shortcut** | `engine/shortcut.rs` | User abbreviations with trigger conditions (Immediate/OnWordBoundary) |
| **Syllable** | `engine/syllable.rs` | Vietnamese syllable parsing and decomposition |
| **Transform** | `engine/transform.rs` | Diacritic/tone placement and transformation |
| **Validation** | `engine/validation.rs` | Vietnamese spelling validation, foreign word detection |
| **Chars** | `data/chars.rs` | Character maps for marks (ă, â, ê, ô, ơ, ư, đ) and tones |
| **Vowel** | `data/vowel.rs` | Vowel phonology tables for tone placement rules |
| **Input** | `input/telex.rs`, `input/vni.rs` | Input method keystroke-to-diacritic mappings |
| **Dictionary** | `data/english_dict.rs` | 100k English words for auto-restore feature |

### FFI Interface (`lib.rs`)

All exports use `#[no_mangle] pub extern "C" fn`:

| Function | Purpose |
|----------|---------|
| `ime_init()` | Initialize engine (call once) |
| `ime_key(key, caps, ctrl)` | Process keystroke |
| `ime_key_ext(key, caps, ctrl, shift)` | Process with shift info |
| `ime_key_with_char(key, caps, ctrl, shift, char_code)` | Process with actual Unicode char |
| `ime_method(method)` | Set input method (0=Telex, 1=VNI) |
| `ime_enabled(enabled)` | Enable/disable processing |
| `ime_clear()` | Clear buffer on word boundary |
| `ime_modern(modern)` | Toggle modern tone placement |
| `ime_free_tone(enabled)` | Toggle free tone mode |
| `ime_esc_restore(enabled)` | Toggle ESC restore |
| `ime_english_auto_restore(enabled)` | Toggle English auto-restore |
| `ime_auto_capitalize(enabled)` | Toggle auto-capitalize |
| `ime_add_shortcut(trigger, replacement)` | Add text shortcut |
| `ime_remove_shortcut(trigger)` | Remove shortcut |
| `ime_restore_word(word)` | Restore word to buffer for editing |
| `ime_get_buffer()` | Get current buffer contents |
| `ime_free(ptr)` | Free result memory |

### Result Struct

```c
struct Result {
    uint32 chars[256];  // UTF-32 codepoints
    uint8  action;      // 0=None, 1=Send, 2=Restore
    uint8  backspace;   // chars to delete
    uint8  count;       // valid chars in array
    uint8  flags;       // bit 0: key_consumed
};
```

---

## Windows Platform (`platforms/windows-wails/`)

Go application using Wails v3 framework with WebView2 for the settings UI.

### Core Components

| Component | File | Purpose |
|-----------|------|---------|
| **Bridge** | `core/bridge.go` | FFI to Rust DLL via `syscall.LoadDLL`. Translates Windows VK → macOS keycodes |
| **Keyboard Hook** | `core/keyboard_hook.go` | Win32 `WH_KEYBOARD_LL` low-level hook. System-wide keystroke interception |
| **IME Loop** | `core/ime_loop.go` | Orchestrates hook → engine → injection pipeline |
| **Text Sender** | `core/text_sender.go` | `SendInput` API text injection with multiple methods |
| **App Detector** | `core/app_detector.go` | Detects foreground process, selects injection profile |
| **Coalescer** | `core/coalescer.go` | Batches rapid keystrokes for flicker-free injection |
| **Smart Paste** | `core/smart_paste.go` | Ctrl+Shift+V mojibake detection and fix |
| **Elevation** | `core/elevation.go` | UAC elevation/de-elevation via `ShellExecute` |

### Services

| Service | File | Purpose |
|---------|------|---------|
| **Settings** | `services/settings.go` | Windows Registry read/write at `HKCU\SOFTWARE\FKey` |
| **Updater** | `services/updater.go` | GitHub VERSION file check, zip download, batch install |
| **Formatting** | `services/formatting.go` | Unicode text formatting config (bold/italic/underline) |

### Frontend

- `frontend/index.html` — Single-page settings UI
- `frontend/assets/app.js` — Application logic, Wails binding calls
- `frontend/assets/app.css` — Styling
- `bindings.go` — `AppBindings` struct exposes Go methods to JavaScript

---

## Key Data Flows

### 1. Keystroke Processing

```
User keypress
    │
    ▼
Win32 LowLevelKeyboardProc (keyboard_hook.go)
    │  ├─ Skip if: injected (FKEY marker), modifier-only, or disabled
    │  ├─ Check hotkey toggle (Ctrl+Shift, etc.)
    │  └─ Detect Shift/CapsLock state
    │
    ▼
ImeLoop.processKey (ime_loop.go)
    │  ├─ Translate VK → macOS keycode (bridge.go)
    │  └─ Call bridge.ProcessKey()
    │
    ▼
Rust FFI: ime_key_ext (lib.rs → engine/mod.rs)
    │  ├─ Check shortcuts (shortcut.rs)
    │  ├─ Buffer management (buffer.rs)
    │  ├─ Vietnamese validation (validation.rs)
    │  ├─ Pattern matching & transform (transform.rs)
    │  └─ Tone/mark placement (syllable.rs + vowel.rs)
    │
    ▼
Result {action, backspace, chars}
    │
    ▼
Text injection (text_sender.go)
    │  ├─ App detector selects method
    │  ├─ Coalescer batches if needed
    │  └─ SendInput: backspaces → Unicode chars
    │
    ▼
Text appears in active application
```

### 2. Settings Flow

```
User changes setting in WebView2 UI
    │
    ▼
JavaScript → Wails binding → AppBindings.SaveSettings()
    │
    ▼
services/settings.go → Write to Windows Registry
    │
    ▼
ImeLoop.UpdateSettings() → Rust FFI calls (ime_method, ime_modern, etc.)
```

### 3. Auto-Update Flow

```
App starts → 3s delay → updater.CheckForUpdates()
    │
    ▼
Fetch raw.githubusercontent.com/miken90/fkey/main/VERSION
    │
    ▼
Compare versions → If newer:
    │  ├─ Download FKey-vX.X.X-portable.zip
    │  ├─ Create batch script (wait, kill, replace, restart)
    │  └─ Run script, quit app
    │
    ▼
Batch script replaces FKey.exe → Restarts
```

---

## Text Injection Methods

The app detector selects the optimal injection method per application:

| Method | How | Used For |
|--------|-----|----------|
| **Fast** | Separate `SendInput` calls with 5ms delay | Most apps (Notepad, VS Code) |
| **Slow** | Per-character with 5ms key + 20/15ms pre/post delay | Electron apps, browsers |
| **Atomic** | Single `SendInput` call with all inputs | Discord (prevents flicker) |
| **Paste** | Clipboard + `Ctrl+V` | Warp terminal, apps that don't support SendInput |

### App Profiles

Apps are matched by process name (e.g., `discord.exe`, `code.exe`, `warp.exe`). Each profile specifies:
- Injection method
- Whether to coalesce keystrokes
- Coalescing timer (ms)
- Backspace mode (VK_BACK vs Unicode BS)

---

## FFI Bridge Design

The Go→Rust bridge uses **`syscall.LoadDLL`** (no CGo dependency):

1. `dll_embed.go` embeds `gonhanh_core.dll` as `//go:embed`
2. At startup, DLL is extracted to temp directory
3. `bridge.go` loads DLL via `syscall.LoadDLL(path)`
4. Each FFI function is resolved via `FindProc("ime_*")`
5. Calls use `proc.Call(uintptr(arg1), uintptr(arg2), ...)`
6. Result struct parsed from raw bytes at known offsets

**Keycode translation**: The Rust engine uses macOS keycodes internally (historical). `TranslateToMacKeycode()` in `bridge.go` maps Windows VK codes → macOS keycodes before each FFI call.

---

## Dependencies

### Rust Core
- **Runtime**: None (only `std`)
- **Dev**: `rstest` 0.18 (parameterized tests), `serial_test` 3.0 (sequential test execution)

### Go Platform
- **Framework**: `wails/v3` v3.0.0-alpha.60
- **System**: `golang.org/x/sys` (Windows Registry, process APIs)
- **Text**: `golang.org/x/text` (encoding for mojibake fix)
- **Build**: Go 1.25, PowerShell

### System Requirements
- Windows 10/11 (64-bit)
- WebView2 Runtime (for settings UI)
- No admin rights required (optional elevation)
