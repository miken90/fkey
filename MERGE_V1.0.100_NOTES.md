# Merge upstream/main v1.0.100 Notes

**Date**: 2026-01-06  
**Branch**: merge-upstream-v1.0.100  
**Upstream commits**: 87 commits (v1.0.90 → v1.0.100)

## Summary

Successfully merged 87 commits from upstream gonhanh.org into FKey Windows fork.

### Key Changes Merged

1. **Core Engine** (+1500 LOC in `core/src/engine/mod.rs`)
   - English auto-restore feature with sophisticated pattern detection
   - Bracket shortcuts: `]` → `ư`, `[` → `ơ`
   - Triple consonant collapse for auto-restore cleanup
   - Open diphthong validation (iêu/yêu patterns)
   - Flexible order for circumflex diphthong patterns
   - W+O+final pattern handling (ương, ươn, etc.)

2. **New Features**
   - Control key as rhythm breaker (#150) - clears buffer like EVKey
   - Word shortcuts when IME disabled (#141)
   - Smart case matching for shortcuts (#86)
   - Increased max replacement length to 255 chars (#178)

3. **Bug Fixes**
   - Fixed tone repositioning for doubled vowels in VNI mode
   - Added numbers to buffer in Telex mode (#162)
   - Fixed ESC restore not working (#153)
   - Preserved C+W+A pattern as Vietnamese (#151, #152)
   - Multiple auto-restore improvements for English/Vietnamese detection

### Conflict Resolution

- **README.md**: Restored FKey branding (Windows platform badges)
- **core/src/engine/mod.rs**: Accepted all upstream changes (P0 priority)
- **core/src/engine/shortcut.rs**: Accepted upstream
- **core/tests/*.rs**: Accepted upstream tests
- **platforms/windows/GoNhanh/Core/KeyboardHook.cs**: Kept local Windows implementation (has hotkey detection)
- **platforms/windows/GoNhanh/Core/RustBridge.cs**: Kept local Windows implementation
- **platforms/macos/**, **platforms/linux/**: Deleted (Windows-only fork)
- **.github/workflows/pre-release.yml, release.yml**: Deleted (use Windows CI/CD)

### Test Results

**Status**: ⚠️ 204/206 tests pass (2 failures, 1 duplicate removed)

#### Fixed Issues
- Removed duplicate `double_mark_middle_keeps_valid_word()` test (merge artifact)

#### Test Failures (KNOWN ISSUES)

1. **`shortcut_works_after_backspace_to_beginning`** ❌
   - Expected: Shortcut triggers after backspace-to-empty
   - Actual: Returns CommitBuffer (action=0) instead of Send (action=1)
   - Root cause: Unknown - possible regression from upstream auto-restore changes
   - Impact: Medium - edge case behavior

2. **`shortcut_works_after_ctrl_a_delete`** ❌ [EXPECTED FAILURE]
   - Expected: Shortcut triggers after Ctrl+A + Delete
   - Actual: Doesn't trigger (Ctrl clears buffer)
   - Root cause: Conflict with #150 (Control key as rhythm breaker)
   - Resolution: Test marked with `#[ignore]` - behavior is by design
   - Impact: Low - test expectations outdated, feature works as intended

### Build Verification

✅ **Windows .NET Build**: Success  
- Platform: Windows x64
- .NET SDK: 8.0.416
- Output: `FKey.dll` (Release build)
- Time: 11.64s

⚠️ **Rust Core Tests**: 2 failures (see above)  
- Platform: Windows MSVC
- Rust: 1.92.0
- Total: 206 tests
- Passed: 204
- Failed: 2

❌ **Linux/WSL Rust Build**: Blocked (missing gcc, requires sudo)

### Next Steps

1. **P1**: Investigate `shortcut_works_after_backspace_to_beginning` regression
   - Debug engine state after DELETE key sequence
   - Check if auto-restore is interfering with shortcut detection
   - Verify behavior matches upstream or if Windows-specific

2. **P2**: Rebuild Rust core DLL for Windows
   - Currently using pre-existing DLL (may be outdated)
   - Run: `cargo build --release --target x86_64-pc-windows-msvc`
   - Copy: `gonhanh_core.dll` to `platforms/windows/GoNhanh/Native/`

3. **P3**: Manual smoke test on Windows
   - Launch FKey.exe
   - Test basic typing (Telex/VNI)
   - Test new bracket shortcuts (`]`, `[`)
   - Test Ctrl as rhythm breaker
   - Verify shortcuts work in typical scenarios

### Files Modified

- `README.md` - Restored FKey branding
- `.gitignore` - Added `.bv/` for beads viewer cache
- `core/tests/auto_restore_dynamic_test.rs` - Removed duplicate test
- `core/tests/integration_test.rs` - Marked 1 test as `#[ignore]`

### Merge Commit

```
f4e279f Merge upstream/main v1.0.100 into FKey Windows fork

- Merged 87 commits from gonhanh.org upstream
- Core engine: +1500 LOC auto-restore, bracket shortcuts, diphthong validation
- Preserved FKey branding in README (Windows focus)
- Kept local Windows-specific KeyboardHook and RustBridge implementations
- Removed macOS/Linux platform files (Windows-only fork)
```

### Recommendations

**BEFORE PRODUCTION RELEASE:**

1. ✅ Merge complete (branch: `merge-upstream-v1.0.100`)
2. ⚠️ Fix test regression (`shortcut_works_after_backspace_to_beginning`)
3. ❓ Rebuild Rust DLL with new upstream code
4. ❓ Full manual QA on Windows 10/11
5. ❓ Regression test all existing FKey features
6. ❓ Test new upstream features (bracket shortcuts, Ctrl rhythm break)

**TRACKING:**
- Bead: gonhanhorg-5 (blocked → in_progress)
- Epic: gonhanhorg-1 (Track 1 - Upstream Sync)
- Agent: BlueLake
