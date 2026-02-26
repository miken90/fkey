---
name: release-github
description: >-
  Build and publish FKey releases to GitHub. Automates version bump, release build,
  ZIP packaging, and GitHub release creation. Trigger: "release X.Y.Z"
license: MIT
version: 4.0.0
---

# GitHub Release Skill

## Auto-Trigger

This skill triggers automatically when the user says:
- "release X.Y.Z"
- "release version X.Y.Z"
- "publish X.Y.Z"

**Agent action**: Run the release script directly — do NOT ask for confirmation:

```bash
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org'; .\.claude\skills\release-github\scripts\github-release.ps1 -Version 'X.Y.Z' 2>&1"
```

## What It Does

1. **Version bump** — Updates `VERSION` file and `winres.json` (all 5 version fields)
2. **Release build** — Rust DLL + Go exe with `-s -w -H=windowsgui -X main.Version=X.Y.Z`
3. **ZIP package** — Single `FKey.exe` (~4.4 MB portable ZIP)
4. **Release notes** — Auto-generated from commits since last tag (conventional commits)
5. **GitHub release** — Tag `vX.Y.Z`, upload ZIP, publish notes via `gh` CLI

## Prerequisites

- `gh` CLI authenticated (`gh auth login`)
- Rust toolchain, Go 1.25+, PowerShell, `go-winres`

## Flags

| Flag | Description |
|------|-------------|
| `-SkipBuild` | Skip build, use existing ZIP |
| `-Draft` | Create as draft release |
| `-Sign` | Sign binary with code signing cert |
| `-DryRun` | Preview release notes, no changes |

## Output

- Local: `platforms/windows-wails/build/bin/FKey-vX.Y.Z-portable.zip`
- GitHub: `https://github.com/miken90/fkey/releases/tag/vX.Y.Z`

## Important Notes

- `build.ps1` auto-updates `winres.json` version fields when `-Version` is passed
- Release script uses **hashtable splatting** to pass `-Release` switch to `build.ps1`
- VERSION file is committed + pushed automatically by the script
- Only pushes to `origin` (miken90), never upstream
