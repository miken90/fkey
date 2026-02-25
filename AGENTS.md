# FKey Vietnamese IME - Agent Instructions

## Environment: WSL on Windows

This project is developed in **WSL** and builds for **Windows only**.

| Platform | Edit Code | Build | Test |
|----------|-----------|-------|------|
| **Windows** | WSL | Windows PowerShell | Windows |

### Path Conventions
- WSL paths: `/mnt/d/WORKSPACES/PERSONAL/fkey/...`
- Windows paths: `D:\WORKSPACES\PERSONAL\fkey\...`
- **Windows builds (PowerShell)**: Use Windows-style paths (`D:\...`)

---

## Documentation

> **Read `docs/` first** before exploring codebase. Scout only when docs are missing or stale.

| Doc | Purpose |
|-----|---------|
| `docs/system-architecture.md` | Architecture layers, data flows, FFI bridge |
| `docs/codebase-summary.md` | Module map, directory tree, stats |
| `docs/code-standards.md` | Rust/Go conventions, commit format, build/test commands |
| `docs/project-overview-pdr.md` | Project overview, features, requirements |
| `docs/project-roadmap.md` | Milestones, future features |

---

## Quick Reference Commands

```bash
# Rust: full test suite
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\core'; cargo test 2>&1"

# Rust: specific test
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\core'; cargo test <test_name> 2>&1"

# Rust: build release DLL
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\core'; cargo build --release 2>&1"

# Go: run tests
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\platforms\windows-wails'; go test ./... 2>&1"

# Go: build (dev)
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\platforms\windows-wails'; .\build.ps1 2>&1"

# Go: build (release)
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\platforms\windows-wails'; .\build.ps1 -Release 2>&1"
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

---

## GitHub Release

Use the `release-github` skill:

```
/release-github 2.0.0
```

Or manually:
```powershell
cd D:\WORKSPACES\PERSONAL\fkey
.\.claude\skills\release-github\scripts\github-release.ps1 -Version "2.0.0"
```

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
