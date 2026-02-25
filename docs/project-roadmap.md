# FKey Vietnamese IME — Project Roadmap

## Current Status

**Version:** 2.3.0  
**Platform:** Windows (x64)  
**Architecture:** Rust core engine + Go/Wails v3 shell + WebView2 UI  
**Distribution:** Single portable exe (~5 MB, embedded DLL)

FKey is a stable, feature-rich Vietnamese IME for Windows. The core engine handles input composition, tone placement, and dictionary lookup. The platform shell manages keyboard hooks, text injection, and system integration.

---

## Completed Milestones (v1.0 → v2.3.0)

### Core Engine
- Telex & VNI input methods with full Vietnamese coverage
- Modern and traditional tone placement algorithms
- Free tone placement mode
- English auto-restore with 100k-word dictionary
- ESC restores raw (pre-composition) input
- Auto-capitalize after sentence boundaries
- Text shortcuts and abbreviations engine

### Windows Platform
- Low-level keyboard hook (Win32 `SetWindowsHookEx`)
- Unicode text injection via `SendInput`
- App-specific injection profiles (Discord, terminals, browsers)
- Coalesced text injection for apps that batch input events
- Smart Paste — clipboard mojibake detection and fix
- Unicode text formatting (bold, italic, underline)
- Custom toggle hotkeys (any modifier + key combination)
- System tray with context menu
- Settings UI via WebView2 (Wails v3)
- Auto-start with Windows (registry)
- Auto-update from GitHub Releases
- Run as Administrator with UAC elevation/de-elevation
- Single-exe portable distribution (embedded `gonhanh_core.dll`)
- Registry-based settings persistence

---

## Future Features

### Short-Term (v2.4 – v2.5)

| Feature | Description | Complexity |
|---------|-------------|------------|
| ARM64 Windows | Build and test on Windows ARM64 devices | Low |
| Multi-monitor OSD | Show input mode overlay on the active monitor | Low |
| Dictionary customization UI | Add/remove words from the auto-restore dictionary | Medium |
| Emoji shortcuts | Type `:smile:` → 😊 via shortcut engine | Medium |

### Medium-Term (v2.6 – v3.0)

| Feature | Description | Complexity |
|---------|-------------|------------|
| Spell checking | Suggest corrections for misspelled Vietnamese words | High |
| Typing statistics | Track words/min, accuracy, language ratio | Medium |
| Cloud sync | Sync settings and shortcuts across devices | Medium |
| Additional input methods | VIQR, VNI extended, custom mappings | Medium |
| Plugin system | Allow third-party extensions for the shortcut engine | High |

### Long-Term (v3.x+)

| Feature | Description | Complexity |
|---------|-------------|------------|
| macOS platform | Native macOS IME using the shared Rust core | High |
| Linux/IBus platform | IBus-based IME for Linux desktops | High |
| Touch keyboard | On-screen Vietnamese keyboard for tablets | High |
| Predictive input | AI-assisted word and phrase prediction | Very High |

---

## Architecture Readiness

The current architecture already supports several future directions:

- **Cross-platform core:** The Rust core is platform-agnostic. It already handles macOS keycodes internally, indicating early design for multi-platform support. A new platform shell (macOS, Linux) only needs to implement keyboard hooks and text injection.
- **Plugin-ready shortcut engine:** The text shortcuts system can be extended to support emoji, templates, and custom transformations without modifying the core composition logic.
- **Modular injection layer:** App-specific profiles demonstrate the injection layer is pluggable — adding new app behaviors requires only configuration, not code changes.
- **FFI boundary:** The clean C FFI between Rust and Go means the core can be consumed by any language (Swift, Python, C++) for new platform shells.

---

## Contributing

### Getting Started

1. Clone the repo and read `AGENTS.md` for build instructions
2. Rust core: `cargo test` in `core/` to verify the engine
3. Windows shell: run `build.ps1` in `platforms/windows-wails/`

### Guidelines

- **Commit format:** `[platform] type: description` (e.g., `[core] feat: emoji shortcut engine`)
- **Test coverage:** All core logic changes must include tests in `core/tests/`
- **Platform isolation:** Never import platform-specific code into the core engine
- **Minimal diffs:** Keep PRs focused on a single feature or fix

### Key Test Suites

| Suite | Path | Purpose |
|-------|------|---------|
| Integration | `core/tests/integration_test.rs` | End-to-end composition |
| Bug regressions | `core/tests/bug_reports_test.rs` | Reported bug fixes |
| English restore | `core/tests/english_auto_restore_test.rs` | Dictionary matching |
| Windows shell | `platforms/windows-wails/tests/fkey_test.go` | Go unit tests |
