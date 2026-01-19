# FKey Vietnamese IME - Agent Instructions

## Environment: WSL on Windows (Hybrid Development)

This project is developed in **WSL** with platform-specific build environments:

| Platform | Edit Code | Build | Test |
|----------|-----------|-------|------|
| **Windows** | WSL | Windows PowerShell | Windows |
| **Linux** | WSL | WSL (native) | WSL/Linux |

### Path Conventions
- WSL paths: `/mnt/c/WORKSPACES/2026/gonhanh.org/...`
- Windows paths: `C:\WORKSPACES\2026\gonhanh.org\...`
- **Windows builds (PowerShell)**: Use Windows-style paths (`C:\...`)
- **Linux builds (WSL bash)**: Use WSL paths (`/mnt/c/...`)

---

## Project Structure

```
fkey/
├── core/                          # Rust core engine (Vietnamese IME logic)
│   ├── src/
│   ├── tests/
│   └── Cargo.toml
├── platforms/
│   ├── windows-wails/             # Windows: Wails v3 Go (~5MB)
│   │   ├── main.go
│   │   ├── core/                  # Go wrapper for Rust DLL
│   │   │   ├── bridge.go          # FFI to gonhanh_core.dll
│   │   │   ├── keyboard_hook.go   # Low-level keyboard hook (Win32)
│   │   │   └── text_sender.go     # SendInput Unicode injection
│   │   ├── services/
│   │   │   ├── settings.go        # Registry-based settings
│   │   │   └── updater.go         # Auto-update checker
│   │   ├── frontend/              # WebView2 UI (HTML/JS/CSS)
│   │   ├── build.ps1              # Build script
│   │   └── wails.json
│   └── linux/                     # Linux: GTK3 + X11 (MVP)
│       ├── main.go
│       ├── core/
│       │   ├── bridge.go          # FFI to libgonhanh_core.so
│       │   ├── keyboard_x11.go    # X11 keyboard hook
│       │   └── text_sender.go     # xdotool text injection
│       ├── config/
│       │   └── config.go          # TOML config (~/.config/fkey/)
│       ├── ui/
│       │   └── tray.go            # GTK3 system tray
│       ├── Makefile
│       └── README.md
├── .claude/skills/                # Agent skills
│   └── release-github/            # GitHub release automation
└── AGENTS.md
```

---

## Commit Message Conventions

### Platform Prefixes

All commits MUST include a platform prefix:

| Prefix | Platform | Example |
|--------|----------|---------|
| `[win]` | Windows only | `[win] fix: auto-update batch file path` |
| `[linux]` | Linux only | `[linux] feat: add system tray icon` |
| `[core]` | Rust core (both platforms) | `[core] fix: tone placement algorithm` |
| `[all]` | All platforms | `[all] docs: update README` |

### Commit Type Prefixes

After platform prefix, use conventional commit type:

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat:` | New feature | Minor (x.Y.0) |
| `fix:` | Bug fix | Patch (x.y.Z) |
| `refactor:` | Code refactoring | Patch |
| `docs:` | Documentation | None |
| `test:` | Tests | None |
| `chore:` | Maintenance | None |

### Examples

```bash
# Windows bug fix → v2.2.2 (patch bump)
git commit -m "[win] fix: Smart Paste hotkey detection order"

# Windows new feature → v2.3.0 (minor bump)
git commit -m "[win] feat: add Smart Paste for mojibake fix"

# Linux feature → v0.2.0-linux
git commit -m "[linux] feat: GTK3 system tray"

# Core fix (affects both) → bump both platforms
git commit -m "[core] fix: tone placement for 'oa' vowels"
```

### Versioning Rules

| Change Type | Version Bump | Example |
|-------------|--------------|---------|
| New feature | `x.Y+1.0` | 2.2.0 → 2.3.0 |
| Bug fix | `x.y.Z+1` | 2.2.0 → 2.2.1 |
| Breaking change | `X+1.0.0` | 2.2.0 → 3.0.0 |

**Platform-specific versions:**
- Windows: `v2.3.0` (no suffix)
- Linux: `v0.2.0-linux` (with `-linux` suffix)

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

---

## Linux Build Commands (WSL Native)

**Prerequisites** (run once in WSL):
```bash
# Ubuntu/Debian
sudo apt update && sudo apt install -y build-essential libgtk-3-dev libx11-dev xdotool

# Install Rust if not present
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

**Build Commands** (run in WSL bash):
```bash
# Build Rust core for Linux
cd /mnt/c/WORKSPACES/2026/gonhanh.org/core
cargo build --release

# Build Linux app
cd /mnt/c/WORKSPACES/2026/gonhanh.org/platforms/linux
make deps      # Install Go dependencies
make build     # Build binary

# Run for testing (requires X11/WSLg)
make run
```

**Testing on WSL**:
- WSL2 with WSLg supports X11 apps natively
- Older WSL needs VcXsrv or X410 on Windows with `export DISPLAY=:0`

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

### Windows Release (Manual)

Use the `release-github` skill:

```
/release-github 2.0.0
```

Or manually:
```powershell
cd C:\WORKSPACES\2026\gonhanh.org
.\.claude\skills\release-github\scripts\github-release.ps1 -Version "2.0.0"
```

### Linux Release (GitHub Actions)

1. Go to: **Actions** → **Release Linux** → **Run workflow**
2. Enter version (e.g., `0.1.0`)
3. Check "prerelease" for beta versions
4. Click **Run workflow**

The workflow will:
- Build Rust core + Go app on Ubuntu
- Create `FKey-{version}-linux-x86_64.tar.gz`
- Create GitHub Release with tag `v{version}-linux`

**Release tags:**
- Windows: `v2.0.9` (no suffix)
- Linux: `v0.1.0-linux` (with `-linux` suffix)

---

## Key Test Files

- `core/tests/english_auto_restore_test.rs` - English word auto-restore tests
- `core/tests/integration_test.rs` - Integration tests
- `core/tests/bug_reports_test.rs` - Bug regression tests
- `platforms/windows-wails/tests/fkey_test.go` - Go unit tests (27 tests)

<!-- bv-agent-instructions-v1 -->

---

## Beads Workflow Integration

This project uses [beads_viewer](https://github.com/Dicklesworthstone/beads_viewer) for issue tracking. Issues are stored in `.beads/` and tracked in git.

### Essential Commands

```bash
# View issues (launches TUI - avoid in automated sessions)
bv

# CLI commands for agents (use these instead)
bd ready              # Show issues ready to work (no blockers)
bd list --status=open # All open issues
bd show <id>          # Full issue details with dependencies
bd create --title="..." --type=task --priority=2
bd update <id> --status=in_progress
bd close <id> --reason="Completed"
bd close <id1> <id2>  # Close multiple issues at once
bd sync               # Commit and push changes
```

### Workflow Pattern

1. **Start**: Run `bd ready` to find actionable work
2. **Claim**: Use `bd update <id> --status=in_progress`
3. **Work**: Implement the task
4. **Complete**: Use `bd close <id>`
5. **Sync**: Always run `bd sync` at session end

### Key Concepts

- **Dependencies**: Issues can block other issues. `bd ready` shows only unblocked work.
- **Priority**: P0=critical, P1=high, P2=medium, P3=low, P4=backlog (use numbers, not words)
- **Types**: task, bug, feature, epic, question, docs
- **Blocking**: `bd dep add <issue> <depends-on>` to add dependencies

### Best Practices

- Check `bd ready` at session start to find available work
- Update status as you work (in_progress → closed)
- Create new issues with `bd create` when you discover tasks
- Use descriptive titles and set appropriate priority/type
- Always `bd sync` before ending session

<!-- end-bv-agent-instructions -->
