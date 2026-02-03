## What's Changed

### New Features

- **[core]** Sync with upstream - dictionary-based auto-restore using Hunspell Vietnamese dictionaries
- **[win]** Add Augment CLI support with Unicode backspace mode

### Bug Fixes

- **[win]** Suppress KEYUP for consumed keys - fixes Firefox address bar first character duplication ([#3](https://github.com/miken90/fkey/issues/3))
- **[win]** Revert terminals to ProfileSlow - fixes missing chars in Claude Code
- **[win]** Terminal lag/missing chars - use atomic injection for terminals
- **[win]** Lazy-load WebView2 to reduce RAM from 33MB to 18MB
- **[win]** Update defaults based on user feedback

### Improvements

- **[win]** Remove debug logs from CLI app detection

---

## Core Engine Updates (from upstream)

- Dictionary-based auto-restore using Hunspell Vietnamese dictionaries
- Option to allow foreign consonants (z, w, j, f) as valid initials
- Fix: collapse double vowel using dict priority
- Fix: support special characters from Option-modified keys
- Fix: auto-restore for "perrmission", "hiss" patterns
- Fix: reset auto-capitalize state on cursor change
- Fix: prevent capacity overflow panic in buffer rebuild
- Fix: delayed circumflex handling improvements
- Fix: horn modifier for "Qu-" initial pattern

---

## Download

| Platform | File | Size |
|----------|------|------|
| Windows (Portable) | FKey-v2.2.5-portable.zip | ~5 MB |

### Installation

1. Download and extract `FKey-v2.2.5-portable.zip`
2. Run `FKey.exe`
3. App runs in system tray

**Full Changelog**: https://github.com/miken90/fkey/compare/v2.2.4...v2.2.5
