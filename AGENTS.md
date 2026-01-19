# FKey Vietnamese IME - Agent Instructions

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
fkey/
├── core/                          # Rust core engine (Vietnamese IME logic)
│   ├── src/
│   ├── tests/
│   └── Cargo.toml
├── platforms/
│   └── windows-wails/             # Wails v3 Go (production, ~5MB)
│       ├── main.go
│       ├── core/                  # Go wrapper for Rust DLL
│       │   ├── bridge.go          # FFI to gonhanh_core.dll
│       │   ├── keyboard_hook.go   # Low-level keyboard hook
│       │   └── text_sender.go     # SendInput Unicode injection
│       ├── services/
│       │   ├── settings.go        # Registry-based settings
│       │   └── updater.go         # Auto-update checker
│       ├── frontend/              # WebView2 UI (HTML/JS/CSS)
│       ├── build.ps1              # Build script
│       └── wails.json
├── .claude/skills/                # Agent skills
│   └── release-github/            # GitHub release automation
└── AGENTS.md
```

---

## Rust Core Commands

```bash
# Full test suite
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo test 2>&1"

# Specific test
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo test pattern9_double_f_words 2>&1"

# Build release DLL
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo build --release 2>&1"
```

---

## Go/Wails Commands

```bash
# Build dev (uses version from git tag)
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\platforms\windows-wails'; .\build.ps1 2>&1"

# Build release (uses version from git tag)
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\platforms\windows-wails'; .\build.ps1 -Release 2>&1"

# Run Go tests
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\platforms\windows-wails'; go test ./... 2>&1"
```

### ⚠️ Version Management (IMPORTANT)

**Before building, ALWAYS verify version is correct:**

1. Check current git tag: `git describe --tags --abbrev=0`
2. Verify `winres.json` matches the tag version in these fields:
   - `RT_MANIFEST.#1.0409.identity.version` (e.g., "2.0.9.0")
   - `RT_VERSION.#1.0409.fixed.file_version` (e.g., "2.0.9.0")
   - `RT_VERSION.#1.0409.fixed.product_version` (e.g., "2.0.9.0")
   - `RT_VERSION.#1.0409.info.0409.FileVersion` (e.g., "2.0.9")
   - `RT_VERSION.#1.0409.info.0409.ProductVersion` (e.g., "2.0.9")

3. If mismatched, update `winres.json` before building

The build script reads version from git tag and injects via `-ldflags`, but `winres.json` must also be updated for Windows executable properties.

---

## Settings Storage (Registry)

Settings stored at `HKEY_CURRENT_USER\SOFTWARE\GoNhanh`:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| InputMethod | DWORD | 0 | 0=Telex, 1=VNI |
| ModernTone | DWORD | 1 | Modern tone placement |
| Enabled | DWORD | 1 | IME enabled |
| AutoStart | DWORD | 0 | Start with Windows |
| EscRestore | DWORD | 1 | ESC restores raw input |
| EnglishAutoRestore | DWORD | 0 | Auto-restore English |
| AutoCapitalize | DWORD | 1 | Auto-capitalize |
| ToggleHotkey | String | "32,1" | Hotkey (keycode,modifiers) |

---

## GitHub Release

Use the `release-github` skill:

```
/release-github 2.0.0
```

Or manually:
```powershell
cd C:\WORKSPACES\2026\gonhanh.org
.\.claude\skills\release-github\scripts\github-release.ps1 -Version "2.0.0"
```

---

## Key Test Files

- `core/tests/english_auto_restore_test.rs` - English word auto-restore tests
- `core/tests/integration_test.rs` - Integration tests
- `core/tests/bug_reports_test.rs` - Bug regression tests
- `platforms/windows-wails/tests/fkey_test.go` - Go unit tests (27 tests)
