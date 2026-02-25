# FKey Vietnamese IME — Codebase Summary

> **Last updated:** 2026-02-25

## 1. Project Overview

FKey is a Vietnamese Input Method Editor (IME) for Windows, built as a two-layer system: a **Rust core engine** handling Vietnamese phonology, diacritics, tone placement, and spelling validation, and a **Go/Wails v3 Windows app** providing the system tray UI, low-level keyboard hook, and text injection via Win32 APIs. The entire application ships as a single ~5 MB executable with the Rust DLL embedded, zero runtime dependencies on the Rust side, and minimal Go dependencies.

---

## 2. Directory Structure

```
fkey/
├── core/                              # Rust core engine (Vietnamese IME logic)
│   ├── src/
│   │   ├── lib.rs                     # FFI C-ABI exports (~916 lines)
│   │   ├── utils.rs                   # String/char utilities (~522 lines)
│   │   ├── engine/
│   │   │   ├── mod.rs                 # Main Engine struct, keystroke processing (~8106 lines)
│   │   │   ├── buffer.rs              # Keystroke buffer management
│   │   │   ├── shortcut.rs            # User-defined abbreviations (~897 lines)
│   │   │   ├── syllable.rs            # Vietnamese syllable parsing
│   │   │   ├── transform.rs           # Diacritic/tone transformation
│   │   │   └── validation.rs          # Vietnamese spelling validation (~676 lines)
│   │   ├── data/
│   │   │   ├── chars.rs               # Vietnamese character maps (mark/tone)
│   │   │   ├── vowel.rs               # Vowel phonology tables (~824 lines)
│   │   │   ├── keys.rs                # macOS keycode constants
│   │   │   ├── english_dict.rs        # 100k English word dictionary
│   │   │   ├── telex_doubles.rs       # Telex double-key patterns (~10k lines)
│   │   │   ├── dictionary.rs          # Vietnamese word validation (HashSet, ~0.5MB)
│   │   │   ├── constants.rs           # Shared constants
│   │   │   └── dictionaries/          # Dictionary files (vi.dic, keep.dic)
│   │   ├── input/
│   │   │   ├── mod.rs                 # Input method trait/types
│   │   │   ├── telex.rs               # Telex input method
│   │   │   └── vni.rs                 # VNI input method
│   │   └── updater/
│   │       └── mod.rs                 # Version parsing
│   ├── tests/                         # 24 test files, ~15k lines
│   └── Cargo.toml                     # Zero runtime dependencies
│
├── platforms/
│   └── windows-wails/                 # Windows app (Go + Wails v3)
│       ├── main.go                    # App entry, tray menu, init
│       ├── bindings.go                # Frontend ↔ Go bindings
│       ├── core/                      # Go wrapper for Rust DLL + Win32
│       │   ├── bridge.go              # Rust DLL FFI bridge
│       │   ├── keyboard_hook.go       # Win32 low-level keyboard hook
│       │   ├── ime_loop.go            # IME processing pipeline
│       │   ├── text_sender.go         # SendInput Unicode text injection
│       │   ├── app_detector.go        # App detection, injection profiles
│       │   ├── coalescer.go           # Keystroke coalescing
│       │   ├── smart_paste.go         # Mojibake fix via Ctrl+Shift+V
│       │   ├── elevation.go           # UAC elevation/de-elevation
│       │   ├── clipboard.go           # Clipboard operations
│       │   ├── format_handler.go      # Unicode text formatting
│       │   └── format_hotkeys.go      # Format hotkey detection
│       ├── services/
│       │   ├── settings.go            # Registry-based settings (HKCU\SOFTWARE\FKey)
│       │   ├── updater.go             # GitHub auto-updater
│       │   └── formatting.go          # Formatting config (JSON)
│       ├── frontend/                  # WebView2 UI (HTML/JS/CSS)
│       ├── tests/
│       │   └── fkey_test.go           # Go unit tests (27 tests)
│       ├── build.ps1                  # PowerShell build script
│       ├── dll_embed.go               # Embeds Rust DLL into Go binary
│       ├── icons.go                   # Runtime tray icon generation
│       └── go.mod                     # Go 1.25, Wails v3 alpha.60
│
├── assets/                            # Logo, banner images
├── docs/                              # Documentation
├── VERSION                            # Current version
├── AGENTS.md                          # Agent instructions
└── README.md                          # User-facing README (Vietnamese)
```

---

## 3. Module Map

### Rust Core (`core/`)

| Module | Purpose | Key Types / Functions |
|--------|---------|----------------------|
| `lib.rs` | C-ABI FFI boundary | `process_key()`, `create_engine()`, `destroy_engine()` — exports consumed by Go via DLL |
| `utils.rs` | String/char helpers | Unicode normalization, char classification, tone/mark detection |
| **engine/** | | |
| `engine/mod.rs` | Central Engine struct | `Engine`, `process_key()`, `handle_backspace()`, `reset()` — main keystroke pipeline |
| `engine/buffer.rs` | Keystroke buffer | Tracks raw input, composed output, cursor position |
| `engine/shortcut.rs` | Abbreviation expansion | User-defined shortcuts, trigger condition matching |
| `engine/syllable.rs` | Syllable parsing | Splits Vietnamese words into onset/nucleus/coda/tone components |
| `engine/transform.rs` | Diacritic/tone ops | Applies/removes marks (ă, ơ, ê…) and tones (sắc, huyền, hỏi, ngã, nặng) |
| `engine/validation.rs` | Spelling rules | Validates Vietnamese syllable structure, consonant clusters, vowel combos |
| **data/** | | |
| `data/chars.rs` | Character maps | Mark → base char mappings, tone → char mappings |
| `data/vowel.rs` | Vowel phonology | Vowel combination tables, tone placement rules per vowel cluster |
| `data/english_dict.rs` | English dictionary | ~100k words for English auto-restore detection |
| `data/dictionary.rs` | Vietnamese dictionary | HashSet-based word validation (~0.5MB), keep list for auto-restore exceptions |
| `data/telex_doubles.rs` | Telex patterns | Double-key reversal patterns (e.g., `aa` → `â` → `aa`) |
| **input/** | | |
| `input/mod.rs` | Input method trait | `InputMethod` trait definition |
| `input/telex.rs` | Telex method | Maps Telex keystrokes (s→sắc, f→huyền, w→ư/ơ, etc.) |
| `input/vni.rs` | VNI method | Maps VNI number keystrokes (1→sắc, 2→huyền, etc.) |

### Go/Wails Platform (`platforms/windows-wails/`)

| Module | Purpose | Key Types / Functions |
|--------|---------|----------------------|
| `main.go` | App entry | Wails app init, system tray menu, window management |
| `bindings.go` | JS ↔ Go bridge | `SettingsService`, `FormattingService` — methods callable from frontend |
| **core/** | | |
| `core/bridge.go` | Rust FFI | `ProcessKey()`, `NewEngine()` — Go wrappers around `gonhanh_core.dll` |
| `core/keyboard_hook.go` | Keyboard hook | Win32 `SetWindowsHookEx(WH_KEYBOARD_LL)`, key event dispatch |
| `core/ime_loop.go` | IME pipeline | Goroutine processing keystroke → engine → text output |
| `core/text_sender.go` | Text injection | `SendInput()` Unicode injection, backspace simulation |
| `core/app_detector.go` | App profiles | Detects foreground app, selects injection strategy (SendInput vs clipboard) |
| `core/coalescer.go` | Coalescing | Batches rapid keystrokes for apps like Discord |
| `core/smart_paste.go` | Mojibake fix | Detects and fixes UTF-8 → CP1252 mojibake via clipboard |
| `core/elevation.go` | UAC handling | Elevates/de-elevates process for admin app input |
| `core/clipboard.go` | Clipboard | Read/write clipboard for paste-based injection |
| `core/format_handler.go` | Text formatting | Unicode bold/italic/strikethrough transforms |
| `core/format_hotkeys.go` | Format hotkeys | Ctrl+B/I/U/S hotkey detection |
| **services/** | | |
| `services/settings.go` | Settings | Read/write `HKCU\SOFTWARE\FKey` registry keys |
| `services/updater.go` | Auto-update | Checks GitHub Releases, downloads `.exe`, applies update |
| `services/formatting.go` | Format config | Loads/saves `formatting.json` |

---

## 4. Key Files Quick Reference

### Core Algorithm Files (start here to understand the engine)

| File | Why It Matters |
|------|----------------|
| `core/src/engine/mod.rs` | **The heart** — 8k lines of keystroke processing, tone/mark logic, undo, English detection |
| `core/src/engine/transform.rs` | How diacritics and tones are applied to characters |
| `core/src/engine/validation.rs` | Vietnamese spelling rules that determine valid output |
| `core/src/engine/syllable.rs` | How input is parsed into Vietnamese syllable components |
| `core/src/data/vowel.rs` | Vowel phonology tables driving tone placement |
| `core/src/input/telex.rs` | Telex key mappings (the most popular input method) |

### Platform Integration Files

| File | Why It Matters |
|------|----------------|
| `platforms/windows-wails/core/bridge.go` | FFI boundary between Go and Rust DLL |
| `platforms/windows-wails/core/keyboard_hook.go` | How keystrokes are intercepted at OS level |
| `platforms/windows-wails/core/text_sender.go` | How Vietnamese text is injected into target apps |
| `platforms/windows-wails/core/app_detector.go` | App-specific injection strategies (terminals, browsers, etc.) |
| `platforms/windows-wails/main.go` | App lifecycle, tray menu, window creation |

### Test Files

| File | Coverage Area |
|------|---------------|
| `core/tests/integration_test.rs` | Full typing sequences (~3697 lines) |
| `core/tests/bug_reports_test.rs` | Regression tests from user reports (~1923 lines) |
| `core/tests/english_auto_restore_test.rs` | English word detection (~1421 lines) |
| `core/tests/typing_test.rs` | Keystroke-by-keystroke typing simulation |
| `platforms/windows-wails/tests/fkey_test.go` | Go unit tests (27 tests, ~782 lines) |

---

## 5. Stats

| Metric | Value |
|--------|-------|
| **Total LOC** (excl. test data) | ~20k |
| Rust core LOC | ~20k (engine ~8k, data ~12k) |
| Go platform LOC | ~4k |
| Rust test LOC | ~15k |
| Go test LOC | ~800 |
| Rust test files | 23 |
| Go test count | 27 |
| Runtime dependencies (Rust) | 0 |
| Runtime dependencies (Go) | Wails v3, golang.org/x/sys |
| Final binary size | ~5 MB (single .exe, DLL embedded) |
| Current version | See `VERSION` file |

---

## 6. Technology Stack

| Layer | Technology | Version | Purpose |
|-------|-----------|---------|---------|
| Core engine | Rust | stable | Vietnamese IME logic, phonology, validation |
| DLL interface | C-ABI FFI | — | Rust ↔ Go boundary |
| Windows app | Go | 1.25 | System integration, keyboard hook, text injection |
| UI framework | Wails | v3 alpha.60 | WebView2 wrapper, system tray, bindings |
| Frontend | HTML/CSS/JS | — | Settings UI in WebView2 |
| Keyboard hook | Win32 API | — | `SetWindowsHookEx(WH_KEYBOARD_LL)` |
| Text injection | Win32 API | — | `SendInput()` Unicode events |
| Settings | Windows Registry | — | `HKCU\SOFTWARE\FKey` |
| Build | PowerShell | — | `build.ps1` orchestrates Rust + Go builds |
| Package | Single .exe | — | DLL embedded via `go:embed` |

---

## 7. Data Flow

```
Keystroke (Win32 hook)
  → keyboard_hook.go (intercept)
  → ime_loop.go (dispatch)
  → bridge.go (FFI call)
  → lib.rs → Engine::process_key()
      → input/telex.rs or vni.rs (map key)
      → engine/transform.rs (apply mark/tone)
      → engine/validation.rs (check spelling)
      → engine/syllable.rs (parse syllable)
  ← EngineResult { committed_text, buffer_display, backspaces }
  → text_sender.go (SendInput or clipboard inject)
  → Target application receives Vietnamese text
```

---

## 8. Settings Storage

All settings stored in Windows Registry at `HKEY_CURRENT_USER\SOFTWARE\FKey`:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| InputMethod | DWORD | 0 | 0=Telex, 1=VNI |
| ModernTone | DWORD | 1 | Modern tone placement |
| Enabled | DWORD | 1 | IME active |
| AutoStart | DWORD | 0 | Start with Windows |
| SkipWShortcut | DWORD | 0 | Skip w→ư in Telex |
| EscRestore | DWORD | 1 | ESC restores raw input |
| FreeTone | DWORD | 0 | Free tone placement |
| EnglishAutoRestore | DWORD | 0 | Auto-restore English words |
| AutoCapitalize | DWORD | 0 | Auto-capitalize |
| ToggleHotkey | String | "0,5" | Hotkey (keycode,modifiers) |
| CoalescingApps | String | "discord,..." | Apps using coalesced injection |
| ShowOSD | DWORD | 0 | OSD on toggle |
| SmartPaste | DWORD | 1 | Mojibake auto-fix |
| RunAsAdmin | DWORD | 0 | Admin privileges |

---

## 9. Build & Test Commands

```bash
# Rust tests
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\core'; cargo test 2>&1"

# Rust DLL build
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\core'; cargo build --release 2>&1"

# Go tests
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\platforms\windows-wails'; go test ./... 2>&1"

# Full Windows build (dev)
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\platforms\windows-wails'; .\build.ps1 2>&1"

# Full Windows build (release)
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\platforms\windows-wails'; .\build.ps1 -Release 2>&1"
```
