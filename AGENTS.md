# FKey Vietnamese IME - Agent Instructions

## Environment: WSL on Windows

This project is developed in **WSL** and builds for **Windows only**.

| Platform | Edit Code | Build | Test |
|----------|-----------|-------|------|
| **Windows** | WSL | Windows PowerShell | Windows |

### Path Conventions
- WSL paths: `/mnt/c/WORKSPACES/2026/gonhanh.org/...`
- Windows paths: `C:\WORKSPACES\2026\gonhanh.org\...`
- **Windows builds (PowerShell)**: Use Windows-style paths (`C:\...`)

---

## Project Structure

```
fkey/
├── core/                          # Rust core engine (Vietnamese IME logic)
│   ├── src/
│   ├── tests/
│   └── Cargo.toml
├── platforms/
│   └── windows-wails/             # Windows: Wails v3 Go (~5MB)
│       ├── main.go
│       ├── core/                  # Go wrapper for Rust DLL
│       │   ├── bridge.go          # FFI to gonhanh_core.dll
│       │   ├── keyboard_hook.go   # Low-level keyboard hook (Win32)
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

## Commit Message Conventions

### Platform Prefixes

All commits MUST include a platform prefix:

| Prefix | Platform | Example |
|--------|----------|---------|
| `[win]` | Windows only | `[win] fix: auto-update batch file path` |
| `[core]` | Rust core | `[core] fix: tone placement algorithm` |
| `[all]` | All (docs, config) | `[all] docs: update README` |

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

# Core fix → v2.2.2 (patch bump)
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
br ready              # Show issues ready to work (no blockers)
br list --status=open # All open issues
br show <id>          # Full issue details with dependencies
br create --title="..." --type=task --priority=2
br update <id> --status=in_progress
br close <id> --reason="Completed"
br close <id1> <id2>  # Close multiple issues at once
br sync               # Commit and push changes
```

### Workflow Pattern

1. **Start**: Run `br ready` to find actionable work
2. **Claim**: Use `br update <id> --status=in_progress`
3. **Work**: Implement the task
4. **Complete**: Use `br close <id>`
5. **Sync**: Always run `br sync` at session end

### Key Concepts

- **Dependencies**: Issues can block other issues. `br ready` shows only unblocked work.
- **Priority**: P0=critical, P1=high, P2=medium, P3=low, P4=backlog (use numbers, not words)
- **Types**: task, bug, feature, epic, question, docs
- **Blocking**: `br dep add <issue> <depends-on>` to add dependencies

### Best Practices

- Check `br ready` at session start to find available work
- Update status as you work (in_progress → closed)
- Create new issues with `br create` when you discover tasks
- Use descriptive titles and set appropriate priority/type
- Always `br sync` before ending session

<!-- end-bv-agent-instructions -->
