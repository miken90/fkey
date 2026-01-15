# GoNhanh Vietnamese IME - Agent Instructions

## Environment: WSL on Windows (Hybrid Development)

This project is developed in **WSL** but compiled/tested on **Windows**.

- **Edit code**: WSL (using Claude/Amp)
- **Build/Test Go/Wails**: Windows PowerShell
- **Build/Test Rust**: Windows PowerShell

### Path Conventions
- WSL paths: `/mnt/c/WORKSPACES/2026/gonhanh.org/...`
- Windows paths: `C:\WORKSPACES\2026\gonhanh.org\...`
- **Inside PowerShell commands**: Always use Windows-style paths (`C:\...`)

---

## Project Structure

```
gonhanh.org/
├── core/                          # Rust core engine (Vietnamese IME logic)
│   ├── src/
│   ├── tests/
│   └── Cargo.toml
├── platforms/
│   └── windows-wails/             # Wails v3 Go (production, ~5MB)
│       ├── main.go
│       ├── app.go
│       ├── core/                  # Go wrapper for Rust DLL
│       │   ├── bridge.go          # FFI to gonhanh_core.dll
│       │   ├── keyboard_hook.go   # Low-level keyboard hook
│       │   └── text_sender.go     # SendInput Unicode injection
│       ├── services/
│       │   └── settings.go        # Registry-based settings (compatible)
│       ├── frontend/              # WebView2 UI (HTML/JS/CSS)
│       │   ├── index.html
│       │   ├── settings.html
│       │   └── assets/
│       ├── build/                 # Build artifacts
│       └── wails.json
├── docs/
├── scripts/
└── AGENTS.md
```

---

## Rust Core Commands

```bash
# Full test suite
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo test 2>&1"

# Specific test
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo test pattern9_double_f_words 2>&1"

# Test file
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo test --test english_auto_restore_test 2>&1"

# Build release DLL
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo build --release 2>&1"

# Copy DLL to Wails app
cp /mnt/c/WORKSPACES/2026/gonhanh.org/core/target/release/gonhanh_core.dll /mnt/c/WORKSPACES/2026/gonhanh.org/platforms/windows-wails/
```

---

## Go/Wails Commands (Windows PowerShell)

```bash
# Check Go version
powershell.exe -Command "go version 2>&1"

# Build Wails app (dev)
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\platforms\windows-wails'; wails3 dev 2>&1"

# Build Wails app (release)
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\platforms\windows-wails'; wails3 build 2>&1"

# Run Go tests
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\platforms\windows-wails'; go test ./... 2>&1"
```

---

## Settings Storage (Registry - Backward Compatible)

Settings are stored in Windows Registry at `HKEY_CURRENT_USER\SOFTWARE\GoNhanh`:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| InputMethod | DWORD | 0 | 0=Telex, 1=VNI |
| ModernTone | DWORD | 1 | Modern tone placement |
| Enabled | DWORD | 1 | IME enabled |
| FirstRun | DWORD | 1 | First run flag |
| AutoStart | DWORD | 0 | Start with Windows |
| SkipWShortcut | DWORD | 0 | Skip w→ư in Telex |
| EscRestore | DWORD | 1 | ESC restores raw input |
| FreeTone | DWORD | 0 | Free tone placement |
| EnglishAutoRestore | DWORD | 0 | Auto-restore English |
| AutoCapitalize | DWORD | 1 | Auto-capitalize |
| ToggleHotkey | String | "32,1" | Hotkey (keycode,modifiers) |

Shortcuts stored at `HKEY_CURRENT_USER\SOFTWARE\GoNhanh\Shortcuts`.

---

## Key Test Files

- `core/tests/english_auto_restore_test.rs` - English word auto-restore tests
- `core/tests/integration_test.rs` - Integration tests
- `core/tests/bug_reports_test.rs` - Bug regression tests

---

## Migration Roadmap (Wails v3)

### Phase 0: Environment Setup ✅
- [x] Install Go 1.21+ on Windows
- [x] Install Wails v3 CLI
- [x] Verify WebView2 runtime

### Phase 1: Keyboard Hook Prototype ✅
- [x] Standalone Go keyboard hook test
- [x] SendInput Unicode injection test

### Phase 2: Rust FFI Integration ✅
- [x] Load gonhanh_core.dll via syscall.LoadDLL
- [x] Map all RustBridge.cs functions to Go

### Phase 3: Core IME Logic ✅
- [x] Port KeyboardHook.cs to Go (`core/keyboard_hook.go`)
- [x] Port TextSender.cs to Go (`core/text_sender.go`)
- [x] Port AppDetector.cs to Go (`core/app_detector.go`)

### Phase 4: Wails + System Tray ✅
- [x] System tray icon (with Vietnamese ON/OFF icons)
- [x] Toggle On/Off (click tray or hotkey)
- [x] Settings menu (input method, settings window)

### Phase 5: Settings UI ✅
- [x] HTML/JS settings window (`frontend/index.html`, `frontend/assets/app.js`)
- [x] Load/save Registry (backward compatible with GoNhanh)
- [x] Hotkey recording and persistence

### Phase 6: Auto-Update ✅
- [x] GitHub Releases API check (`services/updater.go`)
- [x] Background check at startup
- [x] "Kiểm tra cập nhật..." menu option with dialogs
- [x] Open release page in browser
- [x] Auto version injection via `-ldflags` (`build.ps1`)

### Phase 7: Build Optimization ✅
- [x] Go build flags (-ldflags="-s -w -trimpath")
- [x] UPX compression (winget install upx)
- [x] **Result: ~4.8MB** (target was < 8MB)

### Phase 8: Testing & Release ✅
- [x] Unit tests (`tests/fkey_test.go` - 27 tests)
  - Settings: ParseHotkey, FormatHotkey, HotkeyRoundTrip, DefaultSettings
  - Updater: IsNewerVersion, IsWindowsAsset
  - Keycode: TranslateToMacKeycode (letters, numbers, special, punctuation, unmapped)
  - IME: ImeResult_GetText, InputMethodConstants, ImeActionConstants, DefaultImeSettings
  - Keyboard Hook: IsLetterKey, IsNumberKey, IsRelevantKey, KeyboardShortcutMatches
  - App Detector: ExtractProcessName, DetermineMethod (slow/fast apps)
  - Text Sender: InjectionMethodConstants, Delays
- [x] Test across apps (Wave Terminal, Claude Code, AMP Code, CLI tools) ✅
- [x] Code signing support (`build.ps1 -Sign`)
  - Self-signed: `.\build.ps1 -CreateCert` then `.\build.ps1 -Release -Sign`
  - PFX file: `.\build.ps1 -Release -Sign -CertPath 'cert.pfx' -CertPassword 'xxx'`
  - Note: Self-signed still triggers SmartScreen; purchase certificate for full trust

---

## Known Issues

- None currently
