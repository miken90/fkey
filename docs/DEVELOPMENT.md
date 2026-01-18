# FKey Developer Setup Guide

This guide helps you set up the development environment for FKey Vietnamese IME on Windows.

## Prerequisites

- Windows 10/11 (64-bit)
- PowerShell 5.1+ or PowerShell Core 7+
- Git

## Quick Setup (winget)

Open PowerShell as **Administrator** and run:

```powershell
# 1. Install Rust
winget install Rustlang.Rustup

# 2. Install Go
winget install GoLang.Go

# 3. Install Node.js (required for Wails frontend)
winget install OpenJS.NodeJS.LTS

# 4. Install Windows SDK (for code signing - optional)
winget install Microsoft.WindowsSDK

# 5. Restart PowerShell to refresh PATH, then install Go tools
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
go install github.com/tc-hib/go-winres@latest
```

## Manual Installation

### 1. Rust (for core engine)

Download and install from: https://rustup.rs/

```powershell
# Verify installation
rustc --version   # Should show: rustc 1.xx.x
cargo --version   # Should show: cargo 1.xx.x
```

### 2. Go (for Windows app)

Download and install from: https://go.dev/dl/

```powershell
# Verify installation
go version   # Should show: go version go1.xx.x windows/amd64
```

### 3. Wails v3 CLI

```powershell
go install github.com/wailsapp/wails/v3/cmd/wails3@latest

# Verify installation
wails3 version
```

### 4. go-winres (for Windows resources/icon)

```powershell
go install github.com/tc-hib/go-winres@latest

# Verify installation
go-winres --help
```

### 5. CGO (C compiler for Go)

Wails requires CGO. Install one of:

**Option A: TDM-GCC (Recommended)**
```powershell
winget install TDM.TDM-GCC
```

**Option B: MSYS2 MinGW**
```powershell
winget install MSYS2.MSYS2
# Then in MSYS2 terminal:
pacman -S mingw-w64-x86_64-gcc
# Add to PATH: C:\msys64\mingw64\bin
```

Verify:
```powershell
gcc --version
```

## Project Structure

```
fkey/
├── core/                      # Rust core (Vietnamese IME logic)
│   ├── src/
│   ├── tests/
│   └── Cargo.toml
├── platforms/
│   └── windows-wails/         # Go/Wails Windows app
│       ├── main.go
│       ├── core/              # FFI bridge to Rust DLL
│       ├── frontend/          # WebView2 UI
│       ├── build.ps1          # Build script
│       └── wails.json
└── docs/
```

## Building

### Build Rust Core (DLL)

```powershell
cd C:\path\to\fkey\core
cargo build --release

# Output: target\release\gonhanh_core.dll
```

### Build Windows App

```powershell
cd C:\path\to\fkey\platforms\windows-wails

# Debug build (for development)
.\build.ps1

# Release build (optimized, single exe)
.\build.ps1 -Release -Version "2.0.8"

# Release with code signing
.\build.ps1 -Release -Version "2.0.8" -Sign
```

### Development Mode (hot reload)

```powershell
cd C:\path\to\fkey\platforms\windows-wails
.\build.ps1 -Dev
```

## Running Tests

### Rust Tests

```powershell
cd C:\path\to\fkey\core

# Run all tests
cargo test

# Run specific test
cargo test pattern9_double_f_words

# Run with output
cargo test -- --nocapture
```

### Go Tests

```powershell
cd C:\path\to\fkey\platforms\windows-wails
go test ./...
```

## Code Signing (Optional)

### Create Self-Signed Certificate (for testing)

```powershell
cd C:\path\to\fkey\platforms\windows-wails
.\build.ps1 -CreateCert
```

Note: Self-signed certs still trigger SmartScreen warnings. For trusted distribution, purchase a certificate from DigiCert, Sectigo, or Comodo (~$200-500/year).

### Build with Signing

```powershell
# Using certificate from Windows store
.\build.ps1 -Release -Sign

# Using PFX file
.\build.ps1 -Release -Sign -CertPath "mycert.pfx" -CertPassword "password"
```

## Creating a Release

### Using Release Script

```powershell
cd C:\path\to\fkey
.\.claude\skills\release-github\scripts\github-release.ps1 -Version "2.0.8"
```

This will:
1. Build release executable
2. Create ZIP package
3. Generate release notes from git commits
4. Create GitHub release with assets

### Manual Release

```powershell
# 1. Build
cd platforms\windows-wails
.\build.ps1 -Release -Version "2.0.8"

# 2. Create tag
git tag v2.0.8
git push origin v2.0.8

# 3. Create release on GitHub
gh release create v2.0.8 .\build\bin\FKey.exe --title "FKey v2.0.8"
```

## Troubleshooting

### "go: command not found"

Add Go to PATH:
```powershell
$env:Path += ";C:\Program Files\Go\bin;$env:USERPROFILE\go\bin"
```

Or add permanently via System Properties → Environment Variables.

### "cargo: command not found"

Run in new terminal after Rust installation, or:
```powershell
$env:Path += ";$env:USERPROFILE\.cargo\bin"
```

### CGO errors / "gcc not found"

Install TDM-GCC and ensure it's in PATH:
```powershell
$env:Path += ";C:\TDM-GCC-64\bin"
```

### Wails build fails

Check Wails doctor:
```powershell
wails3 doctor
```

### Windows Defender blocks exe

- Right-click ZIP → Properties → Unblock
- Or add folder to Windows Security exclusions
- UPX compression has been disabled to reduce false positives

## IDE Setup

### VS Code (Recommended)

Install extensions:
- `rust-analyzer` - Rust language support
- `Go` - Go language support
- `Even Better TOML` - Cargo.toml support

### GoLand / IntelliJ

- Install Rust plugin
- Configure Go SDK

## Environment Variables

Optional but recommended in your PowerShell profile (`$PROFILE`):

```powershell
# Go
$env:GOPATH = "$env:USERPROFILE\go"
$env:Path += ";$env:GOPATH\bin"

# Rust
$env:Path += ";$env:USERPROFILE\.cargo\bin"

# Project shortcut
function fkey { cd "C:\path\to\fkey" }
```

## WSL Development (Alternative)

If using WSL for editing:

```bash
# Edit in WSL
cd /mnt/c/path/to/fkey
code .

# Build in Windows PowerShell
powershell.exe -Command "cd 'C:\path\to\fkey\platforms\windows-wails'; .\build.ps1 -Release"
```

## Useful Commands

```powershell
# Check all tool versions
go version; cargo --version; wails3 version; go-winres --help | Select-Object -First 1

# Clean build artifacts
.\build.ps1 -Clean

# Format Rust code
cd core; cargo fmt

# Format Go code
cd platforms\windows-wails; go fmt ./...

# Lint Rust
cd core; cargo clippy

# Lint Go
cd platforms\windows-wails; go vet ./...
```
