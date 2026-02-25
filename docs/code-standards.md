# FKey Code Standards

## General Principles

- **YAGNI** — Don't build it until you need it. No speculative abstractions.
- **KISS** — Prefer straightforward code over clever code. Readability wins.
- **Zero external runtime deps (Rust core)** — The core engine uses only `std`. This keeps the DLL small, auditable, and free of supply-chain risk. Dev-only deps (`rstest`, `serial_test`) are acceptable.
- **Minimal deps (Go)** — Use `golang.org/x/sys` for Windows APIs. Avoid CGo entirely — FFI through `syscall.LoadDLL`.
- **No heap allocation per keystroke** — The core processes input in a fixed-size buffer (`MAX = 256`). Performance-critical paths must not allocate.
- **Validation-first** — Validate Vietnamese spelling constraints before applying transforms. Scan the whole buffer (pattern-based), not case-by-case.

---

## Rust Core Conventions

### Naming

| Element       | Style         | Example                        |
|---------------|---------------|--------------------------------|
| Functions     | `snake_case`  | `process_key`, `find_vowel`    |
| Variables     | `snake_case`  | `tone_mark`, `buffer_len`      |
| Types/Enums   | `PascalCase`  | `Engine`, `ToneType`           |
| Constants     | `UPPER_SNAKE` | `MAX`, `VK_RETURN`             |
| Modules       | `snake_case`  | `mod.rs` pattern               |

### FFI Boundary

- All public FFI functions: `#[no_mangle] pub extern "C" fn ime_*()`.
- Return `#[repr(C)]` structs with fixed-size arrays (`[u32; MAX]`) — no dynamically sized types across FFI.
- Heap-allocated returns use `Box::into_raw`. Caller frees via `ime_free()`.
- **Never panic** across the FFI boundary. Return null pointers or zero-value structs for error/uninitialized states.

### Engine & Thread Safety

- Global `Mutex<Option<Engine>>` singleton for the engine instance.
- `lock_engine()` helper recovers from poisoned mutex automatically.
- All engine access goes through this lock — no direct global mutation.

### Error Handling

- Pattern matching for internal error paths — no `unwrap()` on fallible operations in production code.
- FFI functions return null/zero on failure, never `Result` or `Option` (C has no equivalent).
- Internal helper functions may use `Option`/`Result` freely.

### Data & Constants

- Vietnamese character tables live in `data/chars.rs` as `const` arrays.
- Keycode mappings in `data/keys.rs` as `const` values.
- Prefer `const` over `static` unless interior mutability is needed.

### Module Structure

```
core/src/
├── lib.rs          # FFI exports, engine singleton
├── engine.rs       # Core IME engine
├── buffer.rs       # Fixed-size input buffer
├── data/
│   ├── chars.rs    # Vietnamese character data
│   └── keys.rs     # Keycode constants
└── ...
```

---

## Go/Wails Conventions

### Naming

Standard Go conventions — no project-specific overrides:

| Element    | Style        | Example                          |
|------------|--------------|----------------------------------|
| Exported   | `PascalCase` | `ProcessKey`, `GetSettings`      |
| Unexported | `camelCase`  | `hookProc`, `sendUnicode`        |
| Packages   | `lowercase`  | `core`, `services`               |

### Package Layout

| Package    | Purpose                                        |
|------------|------------------------------------------------|
| `core/`    | Low-level: FFI bridge, keyboard hook, text sender, app detector |
| `services/`| App-level: settings, updater, formatting       |

Keep low-level Win32 interactions in `core/`. Business logic and user-facing features in `services/`.

### FFI (No CGo)

```go
dll = syscall.MustLoadDLL("gonhanh_core.dll")
proc = dll.MustFindProc("ime_process_key")
```

- Load DLL once via `sync.Once` in Bridge initialization.
- All `FindProc` calls happen at init time, not per-keystroke.
- Match C struct layouts exactly, including explicit padding fields.

### Win32 API

- Use `windows.LazyDLL` / `NewProc` for system DLLs (`user32.dll`, `kernel32.dll`).
- Define Windows constants (`VK_*`, `WM_*`, `WH_*`) as `const` blocks.
- Struct layouts must match Windows C headers — use `unsafe.Sizeof` to verify when in doubt.

### Settings (Registry)

- All settings stored at `HKEY_CURRENT_USER\SOFTWARE\FKey`.
- Use `golang.org/x/sys/windows/registry` for access.
- Support migration from legacy `GoNhanh` registry key.
- Settings methods return sensible defaults on read failure — never crash on missing keys.

### Error Handling

- Return `error` from functions that can fail. Log errors but keep the app running where possible.
- Critical failures (DLL not found, hook install failure) may terminate the app.
- Non-critical failures (registry read, update check) are logged and swallowed.

### Wails Bindings

- `AppBindings` struct exposes methods to the frontend WebView2 UI.
- Settings returned as `map[string]interface{}` for JS interop.
- Keep binding methods thin — delegate to `services/` or `core/`.

---

## Commit Message Format

### Structure

```
[platform] type: description
```

### Platform Prefixes

| Prefix   | Scope                  |
|----------|------------------------|
| `[core]` | Rust core engine       |
| `[win]`  | Windows Go/Wails app   |
| `[all]`  | Cross-cutting (docs, config, CI) |

### Commit Types

| Type        | Description         | Version Bump   |
|-------------|---------------------|----------------|
| `feat:`     | New feature         | Minor (x.Y.0)  |
| `fix:`      | Bug fix             | Patch (x.y.Z)  |
| `refactor:` | Code restructuring  | Patch          |
| `docs:`     | Documentation       | None           |
| `test:`     | Tests only          | None           |
| `chore:`    | Maintenance/tooling | None           |

### Examples

```
[core] fix: tone placement for 'oa' vowel clusters
[win] feat: add Smart Paste for mojibake fix
[all] docs: update code standards
```

### Versioning

- **Minor bump** (`x.Y+1.0`): new user-facing feature
- **Patch bump** (`x.y.Z+1`): bug fix or refactor
- **Major bump** (`X+1.0.0`): breaking change (rare)

---

## Testing Standards

### Rust Core

- Test files live in `core/tests/` as integration-style tests.
- All tests use `#[serial]` (from `serial_test`) because they share the global engine singleton.
- Use `rstest` for parameterized tests with `#[rstest]` + `#[case]`.
- Bug regressions go in `bug_reports_test.rs` with a comment referencing the original report.
- Run all tests:
  ```bash
  powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\core'; cargo test 2>&1"
  ```
- Run a specific test:
  ```bash
  powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\core'; cargo test test_name 2>&1"
  ```

### Go/Wails

- Tests in `platforms/windows-wails/tests/` and alongside source files.
- Standard `testing` package. No external test frameworks.
- Run all tests:
  ```bash
  powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\platforms\windows-wails'; go test ./... 2>&1"
  ```

### Test Environment

- All tests run on **Windows** (via PowerShell), even when editing code in WSL.
- Do not use mocks or fake data to make tests pass — test real behavior.
- New features require tests. Bug fixes require a regression test.

---

## Build & Release

### Development Build

```bash
# Rust core DLL
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\core'; cargo build --release 2>&1"

# Windows app (version from git tag)
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\platforms\windows-wails'; .\build.ps1 2>&1"
```

### Release Build

```bash
powershell.exe -Command "cd 'D:\WORKSPACES\PERSONAL\fkey\platforms\windows-wails'; .\build.ps1 -Release 2>&1"
```

### Version Checklist

Before building a release:

1. Verify git tag: `git describe --tags --abbrev=0`
2. Update `winres.json` version fields to match:
   - `RT_MANIFEST` identity version
   - `RT_VERSION` file/product versions (both `x.y.z.0` and `x.y.z` forms)
3. Build injects version via `-ldflags -X main.Version=x.x.x`

### GitHub Release

Use the `release-github` skill or run manually:

```powershell
.\.claude\skills\release-github\scripts\github-release.ps1 -Version "x.y.z"
```
