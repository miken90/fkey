---
name: release-github
description: >-
  Build and upload releases to GitHub Releases using standardized build scripts.
  Automates version tagging, portable package creation, and release publishing.
  Use when releasing new versions or creating distribution packages.
license: MIT
version: 3.0.0
---

# GitHub Release Skill

Automate building and publishing releases to GitHub using project's standardized build pipeline.

## When to Use This Skill

Use this skill when:
- Releasing a new version of FKey
- Creating portable distribution packages
- Publishing release to GitHub with auto-generated notes
- Tagging versions with semantic versioning

## Usage

```
/release-github <version>
```

Example:
```
/release-github 2.0.0
```

## What It Does

1. **Build Portable Package** - Uses `platforms/windows-wails/build.ps1`
   - Builds Rust core DLL (if needed)
   - Builds Go/Wails executable with version injection
   - Applies UPX compression (if available)
   - Optionally signs binaries
   - Creates portable ZIP (~5MB)

2. **Generate Release Notes** - Auto-generates from commits since last tag

3. **Create GitHub Release** - Uses `gh` CLI to:
   - Create version tag (e.g., `v2.0.0`)
   - Upload portable ZIP
   - Publish release notes

## Prerequisites

- `gh` CLI installed and authenticated (`gh auth login`)
- Rust toolchain (`cargo`)
- Go 1.21+ (`go`)
- PowerShell 5.1+
- Write access to repository
- Optional: UPX for compression (`winget install upx`)
- Optional: Windows SDK for signing (`signtool.exe`)

## Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| version | Yes | Semantic version (e.g., 2.0.0) |

## Flags

| Flag | Description |
|------|-------------|
| `-SkipBuild` | Skip build step (use existing ZIP) |
| `-Draft` | Create as draft release |
| `-Sign` | Sign binaries with code signing certificate |

## Output

- Local: `platforms/windows-wails/build/bin/FKey-v{version}-portable.zip`
- GitHub: `https://github.com/{owner}/{repo}/releases/tag/v{version}`

## Build Process

Uses `platforms/windows-wails/build.ps1`:

### ⚠️ Pre-Build Checklist (IMPORTANT)

**Before building, ALWAYS verify version:**

1. **Check git tag**: `git describe --tags --abbrev=0`
2. **Verify `winres.json`** matches tag version:
   - `RT_MANIFEST.#1.0409.identity.version` → "X.Y.Z.0"
   - `RT_VERSION.#1.0409.fixed.file_version` → "X.Y.Z.0"
   - `RT_VERSION.#1.0409.fixed.product_version` → "X.Y.Z.0"
   - `RT_VERSION.#1.0409.info.0409.FileVersion` → "X.Y.Z"
   - `RT_VERSION.#1.0409.info.0409.ProductVersion` → "X.Y.Z"
3. **Update `winres.json`** if mismatched before building

### Build Steps

1. Build Rust core DLL (`cargo build --release`) if needed
2. Build Go executable with ldflags:
   - `-s -w` - Strip symbols
   - `-H=windowsgui` - Windows GUI app
   - `-X main.Version={version}` - Inject version
3. Apply UPX compression (optional, ~4.8MB result)
4. Sign binaries (optional)
5. Create ZIP package

## Release Notes Format

Auto-categorizes commits using conventional commit prefixes:

```markdown
## What's Changed

### ✨ New Features

- Feature description 1
- Feature description 2

### 🐛 Bug Fixes

- Bug fix description

### ⚡ Improvements

- Improvement/refactor description

---

## 📦 Download

| Platform | File | Size |
|----------|------|------|
| Windows (Portable) | [FKey-v2.0.0-portable.zip](...) | ~5 MB |

### Installation

1. Download và giải nén `FKey-v2.0.0-portable.zip`
2. Chạy `FKey.exe`
3. Ứng dụng chạy ở khay hệ thống (system tray)

**Full Changelog**: [compare link]
```

### Commit Prefixes

| Prefix | Section |
|--------|---------|
| `feat:`, `feature:`, `add:`, `new:` | ✨ New Features |
| `fix:`, `bug:`, `hotfix:` | 🐛 Bug Fixes |
| `refactor:`, `perf:`, `chore:`, `docs:`, `test:`, `ci:`, `build:` | ⚡ Improvements |
| (other) | ⚡ Improvements |

## Manual Execution

```powershell
# Full release
.\.claude\skills\release-github\scripts\github-release.ps1 -Version "2.0.0"

# Draft release
.\.claude\skills\release-github\scripts\github-release.ps1 -Version "2.0.0" -Draft

# Skip build (use existing ZIP)
.\.claude\skills\release-github\scripts\github-release.ps1 -Version "2.0.0" -SkipBuild

# With signing
.\.claude\skills\release-github\scripts\github-release.ps1 -Version "2.0.0" -Sign
```

## Direct Build Script

If you only need to build without releasing:

```powershell
cd platforms/windows-wails
.\build.ps1 -Release -Version "2.0.0"

# With signing
.\build.ps1 -Release -Version "2.0.0" -Sign
```

## Integration

This skill integrates with:
- `platforms/windows-wails/build.ps1` - Go/Wails build process
- `gh` CLI - GitHub release creation
- Git tags - Version management
- Semantic versioning - Version numbering
