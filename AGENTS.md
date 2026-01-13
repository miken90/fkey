# GoNhanh Vietnamese IME - Agent Instructions

## Environment: WSL on Windows

This project runs in WSL (Windows Subsystem for Linux) but the Rust toolchain is installed on the Windows side. Use PowerShell to run Rust/Cargo commands:

```bash
# Run cargo commands via PowerShell (required on WSL)
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo test 2>&1"

# Run specific test
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo test pattern9_double_f_words 2>&1"

# Run test file
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo test --test english_auto_restore_test 2>&1"

# Build
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo build 2>&1"

# Check
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo check 2>&1"
```

**Important**: Always use Windows-style paths (`C:\...`) inside the PowerShell command, not WSL paths (`/mnt/c/...`).

## Project Structure

- `core/` - Rust core engine (Vietnamese IME logic)
- `platforms/` - Platform-specific implementations (macOS, Windows)
- `docs/` - Documentation
- `scripts/` - Build and utility scripts

## Key Test Files

- `core/tests/english_auto_restore_test.rs` - English word auto-restore tests
- `core/tests/integration_test.rs` - Integration tests
- `core/tests/bug_reports_test.rs` - Bug regression tests

## Common Commands

```bash
# Full test suite
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo test 2>&1"

# Typecheck only
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo check 2>&1"

# Build release DLL
powershell.exe -Command "cd 'C:\WORKSPACES\2026\gonhanh.org\core'; cargo build --release 2>&1"

# Copy DLL to Windows app (after building)
cp /mnt/c/WORKSPACES/2026/gonhanh.org/core/target/release/gonhanh_core.dll /mnt/c/WORKSPACES/2026/gonhanh.org/platforms/windows/GoNhanh/Native/
```

## Known Issues

- None currently
