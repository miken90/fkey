---
name: release-github
description: Build and upload releases to GitHub Releases using standardized build scripts. Automates version tagging, portable package creation, and release publishing. Use when releasing new versions or creating distribution packages.
license: MIT
version: 2.0.0
---

# GitHub Release Skill

Automate building and publishing releases to GitHub using project's standardized build pipeline.

## When to Use This Skill

Use this skill when:
- Releasing a new version of GoNhanh
- Creating portable distribution packages
- Publishing release to GitHub with auto-generated notes
- Tagging versions with semantic versioning

## Usage

```
/release-github <version>
```

Example:
```
/release-github 1.5.9
```

## What It Does

1. **Build Portable Package** - Uses `platforms/windows/build-release.ps1`
   - Stops running instances
   - Cleans previous builds
   - Builds self-contained single-file executable
   - Creates portable ZIP (~64MB)

2. **Generate Release Notes** - Auto-generates from commits since last tag

3. **Create GitHub Release** - Uses `gh` CLI to:
   - Create version tag (e.g., `v1.5.9`)
   - Upload portable ZIP
   - Publish release notes

## Prerequisites

- `gh` CLI installed and authenticated (`gh auth login`)
- .NET 8 SDK (`dotnet`)
- PowerShell 5.1+
- Write access to repository

## Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| version | Yes | Semantic version (e.g., 1.5.9) |

## Flags

| Flag | Description |
|------|-------------|
| `-SkipBuild` | Skip build step (use existing ZIP) |
| `-Draft` | Create as draft release |

## Output

- Local: `platforms/windows/GoNhanh/bin/Release/net8.0-windows/win-x64/publish/GoNhanh-v{version}-portable.zip`
- GitHub: `https://github.com/{owner}/{repo}/releases/tag/v{version}`

## Build Process

Uses `platforms/windows/build-release.ps1`:
1. Stop GoNhanh.exe processes
2. Clean `bin/Release` directory
3. Build with `dotnet publish`:
   - Configuration: Release
   - Runtime: win-x64
   - Self-contained: Yes
   - Single file: Yes
   - Version: {version}
4. Create ZIP with `GoNhanh.exe`

## Release Notes Format

```markdown
## 📦 What's New in v{version}

- feat: description
- fix: description
...

## 💾 Download

- **Windows Portable**: [link] (~XX MB)

## 🔧 Installation

1. Download zip
2. Extract and run GoNhanh.exe
3. App runs in system tray

**Full Changelog**: [compare link]
```

## Manual Execution

```powershell
# Full release
.\.claude\skills\release-github\scripts\github-release.ps1 -Version "1.5.9"

# Draft release
.\.claude\skills\release-github\scripts\github-release.ps1 -Version "1.5.9" -Draft

# Skip build (use existing ZIP)
.\.claude\skills\release-github\scripts\github-release.ps1 -Version "1.5.9" -SkipBuild
```

## Direct Build Script

If you only need to build without releasing:

```powershell
cd platforms/windows
.\build-release.ps1 -Version "1.5.9"
```

## Integration

This skill integrates with:
- `platforms/windows/build-release.ps1` - Standardized build process
- `gh` CLI - GitHub release creation
- Git tags - Version management
- Semantic versioning - Version numbering
