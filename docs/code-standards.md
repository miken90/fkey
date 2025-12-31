# FKey: Code Standards & Guidelines

## Rust Coding Standards

### Formatting & Linting
- **Formatter**: `cargo fmt` (automatic, non-negotiable)
- **Linter**: `cargo clippy -- -D warnings` (no warnings allowed)
- **Pre-commit**: Format check runs automatically on all PRs

### Code Style
- **Naming**: snake_case for functions/variables, CamelCase for types
- **Comments**: Inline comments for "why", not "what"
- **Documentation**: Public items require rustdoc comments (`///`)
- **Module Layout**:
  ```rust
  //! Module-level documentation

  // Imports
  use crate::module::Item;

  // Constants
  pub const MAX_BUFFER: usize = 64;

  // Type definitions
  pub struct MyStruct { }

  // Public functions
  pub fn my_function() { }

  // Private implementation
  fn private_helper() { }

  #[cfg(test)]
  mod tests { }
  ```

### Zero-Dependency Philosophy
- **Core** (`core/src/`): **Absolutely no external dependencies** in production
- **Rationale**: FFI library must be lightweight and self-contained
- **Allowed**: Only `rstest` in dev-dependencies for test parametrization
- **Example**: Character tables are hardcoded, not loaded from files

### API Guidelines
- **FFI Safety**: All public functions marked `extern "C"` are unsafe by contract
- **Memory**: FFI results must be freed by caller (`ime_free(ptr)`)
- **Error Handling**: Return `null` or default value, never panic (C compatibility)
- **Documentation**: C/FFI comments in code blocks (see lib.rs examples)

### Testing
- **Coverage**: Every module must have `#[cfg(test)] mod tests { }` and integration tests in `core/tests/`
- **Parametrization**: Use `#[rstest]` for multiple test cases
- **Integration**: `core/tests/` directory for full pipeline tests (700+ tests across 15 test files)
- **Test Files**:
  - `unit_test.rs` - Individual module tests
  - `typing_test.rs` - Full keystroke sequences (Telex + VNI)
  - `engine_test.rs` - Engine state + initialization
  - `integration_test.rs` - End-to-end keystroke→output
  - `paragraph_test.rs` - Multi-word paragraph typing
  - `auto_capitalize_test.rs` - Auto-capitalization edge cases
  - `auto_restore_dynamic_test.rs` - Dynamic auto-restore patterns
  - `english_auto_restore_test.rs` - English word detection
  - `english_auto_restore_toggle_test.rs` - Toggle behavior
  - `english_words_test.rs` - English vocabulary patterns
  - `bug_reports_test.rs` - Regression tests for reported bugs
  - `dynamic_test.rs` - Dynamic input scenarios
  - `permutation_test.rs` - Input permutation coverage
  - `revert_auto_restore_test.rs` - Revert behavior
  - `shortcut_test.rs` - Shortcut expansion tests
- **Naming**: `test_feature_case_expected` (e.g., `test_telex_a_s_returns_á`)
- **Run**: `cd core && cargo test`

### Examples
```rust
/// Convert Vietnamese vowel to uppercase.
///
/// # Arguments
/// * `ch` - Lowercase Vietnamese vowel (a-z)
///
/// # Returns
/// Uppercase equivalent using Unicode rules
pub fn to_uppercase(ch: char) -> char {
    ch.to_uppercase().next().unwrap_or(ch)
}

#[cfg(test)]
mod tests {
    use super::*;
    use rstest::rstest;

    #[rstest]
    #[case('á', 'Á')]
    #[case('ơ', 'Ơ')]
    #[case('a', 'A')]
    fn test_vietnamese_uppercase(#[case] input: char, #[case] expected: char) {
        assert_eq!(to_uppercase(input), expected);
    }
}
```

## C# Coding Standards (Windows Platform)

### Style Guide
- **Authority**: [Microsoft C# Coding Conventions](https://docs.microsoft.com/en-us/dotnet/csharp/fundamentals/coding-style/coding-conventions)
- **Formatting**: Follow Visual Studio default formatting (4-space indentation)
- **Naming**: PascalCase for public members, camelCase for private fields with `_` prefix

### File Organization
```csharp
// MARK: - Usings
using System;
using System.Windows;

// MARK: - Namespace
namespace GoNhanh.Core
{
    // MARK: - Class
    public class KeyboardHook
    {
        // MARK: - Fields
        private IntPtr _hookId;

        // MARK: - Properties
        public bool IsEnabled { get; set; }

        // MARK: - Constructor
        public KeyboardHook() { }

        // MARK: - Public Methods
        public void Start() { }

        // MARK: - Private Methods
        private void ProcessKey() { }
    }
}
```

### Threading & Async
- **UI Thread**: All WPF UI updates must be on Dispatcher thread
- **Background Work**: Use async/await pattern
- **Thread-Safe**: Use lock or concurrent collections
- **Example**:
  ```csharp
  Application.Current.Dispatcher.Invoke(() => {
      // Update UI
  });
  ```

### Text Injection Modes
- **Fast Mode** (default): 2ms batch delay for standard Win32 apps
- **Slow Mode** (app-aware): 20ms + 15ms + 5ms per char for Electron/terminals
- **App Detection**: AppDetector.cs checks foreground window process
- **Slow Apps List**: Wave, Windows Terminal, cmd, PowerShell, Chrome, VS Code, Cursor, Notion, Slack, Discord, Obsidian, Figma

### Error Handling
- **Try-Catch**: Handle exceptions at boundary (hooks, FFI calls)
- **Logging**: Use Debug.WriteLine for development, structured logging for production
- **User Messages**: Show MessageBox for critical errors only

## FFI (Foreign Function Interface) Conventions

### C ABI Compatibility
- **Representation**: `#[repr(C)]` for all shared structs
- **Types**: Use fixed-size types (u8, u16, u32, not usize)
- **Alignment**: Match Rust layout exactly in Swift struct

### Struct Layout (ImeResult Example)
```rust
// Rust
#[repr(C)]
pub struct Result {
    pub chars: [u32; 32],    // Fixed-size array (128 bytes)
    pub action: u8,           // 1 byte
    pub backspace: u8,        // 1 byte
    pub count: u8,            // 1 byte
    pub _pad: u8,             // 1 byte padding
}
```

```csharp
// C# - MUST match Rust layout byte-for-byte
[StructLayout(LayoutKind.Sequential)]
public struct ImeResult
{
    [MarshalAs(UnmanagedType.ByValArray, SizeConst = 32)]
    public uint[] chars;     // 32 elements
    public byte action;
    public byte backspace;
    public byte count;
    public byte _pad;
}
```

### Pointer Management
- **Ownership**: Function that allocates owns the pointer
- **Deallocation**: Caller must call `ime_free(ptr)` to deallocate
- **Safety**: Use try/finally to guarantee cleanup

```csharp
IntPtr resultPtr = RustBridge.ime_key(keyCode, caps, ctrl);
try
{
    if (resultPtr != IntPtr.Zero)
    {
        var result = Marshal.PtrToStructure<ImeResult>(resultPtr);
        // Process result...
    }
}
finally
{
    if (resultPtr != IntPtr.Zero)
        RustBridge.ime_free(resultPtr);
}
```

### Function Declarations
```csharp
// Import with exact name and signature
[DllImport("gonhanh_core.dll", CallingConvention = CallingConvention.Cdecl)]
public static extern IntPtr ime_key(ushort key, bool caps, bool ctrl);

// Safety: Check for null, use try/finally for cleanup
IntPtr resultPtr = ime_key(keyCode, caps, ctrl);
try
{
    if (resultPtr != IntPtr.Zero)
    {
        var result = Marshal.PtrToStructure<ImeResult>(resultPtr);
        // Safe to use
    }
}
finally
{
    if (resultPtr != IntPtr.Zero)
        ime_free(resultPtr);
}
```

## Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
| Type | Purpose | Example |
|------|---------|---------|
| **feat** | New feature | `feat(engine): add shortcut table expansion` |
| **fix** | Bug fix | `fix(transform): correct ư placement in compound vowels` |
| **docs** | Documentation | `docs(ffi): clarify memory ownership in ime_free` |
| **style** | Formatting | `style(rust): apply cargo fmt to engine module` |
| **test** | Tests | `test(validation): add edge case for invalid syllables` |
| **chore** | Build/CI | `chore(ci): update GitHub Actions workflow` |
| **refactor** | Code reorganization | `refactor(buffer): optimize circular buffer lookup` |

### Scope
- `engine`, `input`, `data`, `ffi`, `ui`, `macos`, `ci`, `docs`, etc.
- Specific to file/module being changed

### Subject Line
- Imperative mood: "add", not "adds" or "added"
- Lowercase first letter
- No period at end
- Maximum 50 characters
- Use as continuation of "If applied, this commit will..."

### Body
- Explain what and why, not how
- Wrap at 72 characters
- Separate from subject with blank line
- Reference issues: "Closes #123"

### Examples
```
feat(engine): add user shortcut table support

Implement ShortcutTable struct to store user-defined abbreviations
with priority matching. Allows users to define custom transforms
like "hv" → "không" (no space).

Closes #45
```

```
fix(transform): handle ư vowel in compound patterns

The ư vowel was not correctly recognized in sequences like "ưu" and
"ươ" due to missing pattern in validation. Add explicit check for
horn modifier on u vowel.

Fixes #78
```

## Documentation Standards

### Code Comments
- **Module-level**: `//!` with purpose and usage examples
- **Function-level**: `///` with Args, Returns, Safety (if applicable)
- **Inline**: `//` for non-obvious logic (skip obvious comments)

### Examples in Docs
```rust
/// Process keystroke and transform Vietnamese text.
///
/// # Arguments
/// * `key` - macOS virtual keycode (0-127)
/// * `caps` - true if Shift or CapsLock pressed
///
/// # Returns
/// Pointer to Result struct with action, backspace count, output chars.
/// Caller must free with `ime_free()`.
///
/// # Example
/// ```c
/// ImeResult* r = ime_key(keys::A, false, false);
/// if (r && r->action == 1) {
///     // Send r->backspace deletes, then r->chars
/// }
/// ime_free(r);
/// ```
#[no_mangle]
pub extern "C" fn ime_key(key: u16, caps: bool, ctrl: bool) -> *mut Result { }
```

## Version Numbering

- **Semantic Versioning**: MAJOR.MINOR.PATCH (e.g., 1.6.0)
- **MAJOR**: Breaking changes (rare)
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes only
- **Release**: Tag with `v` prefix (e.g., `v1.7.4`)
- **Current Version**: v1.7.4 (Windows production-ready)

## Pull Request Guidelines

- **Title**: Follow commit message format
- **Description**: Reference related issues
- **Changes**: One logical change per PR (no mega-PRs)
- **Tests**: All new code must have tests
- **CI**: Must pass format, clippy, and test suite
- **Review**: At least one approval before merge

---

**Last Updated**: 2025-12-31
**Enforced By**: GitHub Actions CI (`ci.yml`)
**Test Coverage**: 700+ integration tests across 15 test files in `core/tests/`
**Platforms**: Windows 10/11 (production)
**Repository**: https://github.com/miken90/gonhanh.org
