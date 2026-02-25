# FKey Documentation

## Quick Start

1. Download [FKey-portable.zip](https://github.com/miken90/fkey/releases/latest)
2. Extract and run `FKey.exe`
3. App runs in system tray

## Features

- **Telex & VNI** input methods
- **Smart Paste** (Ctrl+Shift+V) — fixes Vietnamese mojibake from clipboard
- **Auto-restore English** words (text, expect, user...)
- **Dictionary-based auto-restore** — uses Hunspell Vietnamese dictionaries
- **ESC to restore** raw input
- **Auto-capitalize** after sentence
- **Custom hotkeys**
- **Run as Administrator** — type Vietnamese in admin apps
- **Terminal support** — optimized for Warp, Claude Code, Augment CLI

## Settings

Settings stored in Windows Registry at `HKEY_CURRENT_USER\SOFTWARE\FKey`

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| InputMethod | DWORD | 0 | 0=Telex, 1=VNI |
| ModernTone | DWORD | 1 | Modern tone placement |
| Enabled | DWORD | 1 | IME enabled |
| AutoStart | DWORD | 0 | Start with Windows |
| SkipWShortcut | DWORD | 0 | Skip w→ư in Telex |
| EscRestore | DWORD | 1 | ESC restores raw input |
| FreeTone | DWORD | 0 | Free tone placement |
| EnglishAutoRestore | DWORD | 0 | Auto-restore English words |
| AutoCapitalize | DWORD | 0 | Auto-capitalize |
| ToggleHotkey | String | "0,5" | Hotkey (keycode,modifiers) — default Ctrl+Shift |
| CoalescingApps | String | "discord,..." | Apps using coalesced text injection |
| ShowOSD | DWORD | 0 | Show OSD when switching |
| SmartPaste | DWORD | 1 | Smart paste mojibake fix |
| RunAsAdmin | DWORD | 0 | Run with admin privileges |

## Project Structure

```
fkey/
├── core/                      # Rust core engine
│   ├── src/
│   │   ├── engine/            # Vietnamese IME engine
│   │   ├── data/              # Dictionaries, word lists
│   │   └── input/             # Input method definitions
│   └── tests/                 # Comprehensive test suite
├── platforms/
│   └── windows-wails/         # Windows app (Wails v3)
│       ├── core/              # Go FFI bridge, keyboard hook
│       ├── services/          # Settings, updater
│       └── frontend/          # WebView2 UI
└── VERSION                    # Current version
```

## Build

```powershell
# Build Rust core
cd core
cargo build --release

# Build Windows app
cd platforms\windows-wails
.\build.ps1 -Release
```

## License

[BSD-3-Clause](../LICENSE)
