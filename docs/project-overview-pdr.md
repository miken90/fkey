# FKey Vietnamese IME — Project Overview

> **Version**: 2.3.0 · **License**: BSD-3-Clause · **Platform**: Windows 10/11 (64-bit)
> **Repository**: [github.com/miken90/fkey](https://github.com/miken90/fkey)

---

## 1. Project Summary

FKey is a free, open-source Vietnamese Input Method Editor (IME) for Windows. Built on the Gõ Nhanh engine by Kha Phan, FKey delivers a lightweight (~5MB), portable typing experience that runs quietly in the system tray. It requires no installation, collects no data, and processes keystrokes in under 1ms.

FKey supports both Telex and VNI input methods, automatically restores mistyped English words from a 100k-word dictionary, and includes Smart Paste to fix Vietnamese mojibake from the clipboard.

---

## 2. Problem Statement

Existing Vietnamese IMEs on Windows suffer from several pain points:

| Problem | Impact |
|---------|--------|
| Bloated installers with bundled adware | Users distrust the software and risk malware |
| Telemetry and data collection | Privacy concerns for users in enterprise/government |
| Poor compatibility with modern apps | Text injection fails in terminals, Electron apps, admin windows |
| No mojibake recovery | Users manually retype garbled Vietnamese text from clipboard |
| Stale development / closed source | No community fixes, no transparency |
| Heavy resource usage | Background processes consume 50–200MB RAM |

FKey addresses all of these by being open-source, portable, ad-free, telemetry-free, and engineered for broad app compatibility with multiple text injection strategies.

---

## 3. Key Features

| Feature | Description |
|---------|-------------|
| **Telex & VNI Input** | Two standard Vietnamese input methods |
| **Smart Paste** | `Ctrl+Shift+V` — detects and fixes Vietnamese mojibake from clipboard |
| **English Auto-Restore** | 100k-word dictionary; words like `text`, `expect`, `user` auto-restore on Space |
| **ESC Restore** | Press ESC to undo diacritics and recover raw input |
| **Auto-Capitalize** | Capitalizes the first letter after sentence-ending punctuation |
| **Tone Placement** | Modern (`hoà`) or traditional (`hòa`) placement modes |
| **Free Tone Mode** | Bypasses tone validation for unrestricted placement |
| **Custom Toggle Hotkey** | Configurable on/off hotkey (default: `Ctrl+Shift`) |
| **Text Shortcuts** | User-defined abbreviations that expand on trigger |
| **Auto-Start** | Optional launch at Windows startup |
| **Auto-Update** | Checks GitHub Releases for new versions |
| **Run as Admin** | Elevates to inject text into admin-level windows |
| **App-Specific Profiles** | Tailored text injection per app (Discord atomic, terminal paste, etc.) |
| **System Tray** | Minimizes to tray with on/off icon and context menu |
| **Settings UI** | WebView2-based settings panel |

---

## 4. Technical Requirements

- **OS**: Windows 10 version 1809+ or Windows 11 (64-bit)
- **Runtime**: WebView2 Runtime (pre-installed on Windows 10 21H2+ and all Windows 11)
- **Privileges**: Standard user; optional Administrator for elevated app support
- **Single Instance**: Enforced via named mutex — only one FKey process runs at a time

---

## 5. Non-functional Requirements

| Requirement | Target |
|-------------|--------|
| Binary size | ~5MB portable zip |
| Memory usage | ~18MB RAM |
| Keystroke latency | <1ms processing overhead |
| Network access | None, except optional update check to GitHub |
| Telemetry | None — zero data collection |
| Advertising | None |
| Installation | Portable — extract and run, no installer |
| Offline operation | Fully functional without internet |
| Privacy | No logs, no analytics, no external calls |

---

## 6. Target Users

- **Vietnamese speakers on Windows** who type daily in Vietnamese and English
- **Privacy-conscious users** who want an IME with no telemetry or bundled software
- **Developers and power users** who need reliable input in terminals, IDEs, and Electron apps
- **Enterprise/government users** who require auditable, open-source input tools
- **Users switching between Vietnamese and English** who benefit from auto-restore of English words

---

## 7. Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Core Engine | **Rust** | Vietnamese text processing, tone placement, dictionary lookup. Zero runtime dependencies, compiles to a ~5k DLL. |
| App Framework | **Go + Wails v3** | Application shell, system tray, WebView2 integration. Produces a single `.exe`. |
| Keyboard Hook | **Win32 API** (`WH_KEYBOARD_LL`) | Intercepts keystrokes at the OS level before any application receives them. |
| Text Injection | **Win32 `SendInput`** | Injects processed Unicode text. Multiple strategies per app (char-by-char, atomic, clipboard paste). |
| Settings Storage | **Windows Registry** | All settings stored under `HKCU\SOFTWARE\FKey`. No config files to manage. |
| Auto-Update | **GitHub Releases** | Compares local `VERSION` against remote; downloads portable zip if newer. |
| UI | **HTML / CSS / JS** | Settings panel rendered in WebView2. Lightweight, no frontend framework. |
| Build System | **PowerShell + Cargo** | `build.ps1` orchestrates Rust DLL compilation and Go binary linking. Development in WSL, builds target Windows. |

---

## 8. Distribution & Installation

### Download

Portable zip available from [GitHub Releases](https://github.com/miken90/fkey/releases).

### Installation Steps

1. Download `FKey-v2.3.0-portable.zip` from the latest release
2. Extract to any folder (e.g., `C:\Tools\FKey\`)
3. Run `FKey.exe`
4. FKey appears in the system tray — ready to type

### Updates

FKey checks GitHub for new versions on launch. When an update is available, it downloads the new zip and replaces itself. No manual intervention required.

### Uninstallation

1. Close FKey from the system tray
2. Delete the FKey folder
3. (Optional) Remove registry keys at `HKCU\SOFTWARE\FKey`
