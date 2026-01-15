# FKey Documentation

## Quick Start

1. Download [FKey-portable.zip](https://github.com/miken90/fkey/releases/latest)
2. Extract and run `FKey.exe`
3. App runs in system tray

## Features

- **Telex & VNI** input methods
- **Auto-restore English** words (text, expect, user...)
- **ESC to restore** raw input
- **Auto-capitalize** after sentence
- **Custom hotkeys**

## Settings

Settings stored in Windows Registry at `HKEY_CURRENT_USER\SOFTWARE\FKey`

| Setting | Default | Description |
|---------|---------|-------------|
| InputMethod | 0 | 0=Telex, 1=VNI |
| ModernTone | 1 | Modern tone placement |
| Enabled | 1 | IME enabled |
| AutoStart | 0 | Start with Windows |
| EscRestore | 1 | ESC restores raw input |
| EnglishAutoRestore | 0 | Auto-restore English words |
| AutoCapitalize | 1 | Auto-capitalize |

## Build

```powershell
# Build Rust core
cd core
cargo build --release

# Build Windows app
cd platforms/windows-wails
.\build.ps1 -Release -Version "2.0.0"
```

## License

[BSD-3-Clause](../LICENSE)
