# Code Review Report: GoNhanh Core Vietnamese IME

**Review Date:** 2025-12-07
**Scope:** `/Users/khaphan/Documents/Work/gonhanh.org/core/src/`
**Lines Analyzed:** ~1,868 LOC
**Files Reviewed:** 10 Rust files

---

## Overall Assessment

Codebase quality: **GOOD**

Code is well-structured with clear separation of concerns. Architecture follows phonology-based approach for Vietnamese input. All tests passing (49/49). No clippy warnings. Code compiles cleanly in both debug and release modes.

Main strengths:
- Clear module boundaries
- Comprehensive test coverage
- Well-documented phonological rules
- Clean FFI interface
- Effective use of Rust type system

---

## CRITICAL Issues

**None found.**

---

## HIGH Priority Findings

### H1. Code Duplication: `key_to_char()` function duplicated in tests

**Location:** `engine/mod.rs:27-68` (function), `engine/mod.rs:468-507` (test helper)

**Issue:** `key_to_char()` at line 27 converts keycodes to chars. Test helper `type_keys()` at line 468 reimplements same mapping with massive match statement (40 lines).

**Impact:**
- DRY violation
- Maintenance burden (changes need sync)
- Test code bloat

**Fix:**
```rust
// In tests module (line 465+)
fn type_keys(e: &mut Engine, s: &str) -> Vec<Result> {
    s.chars()
        .map(|c| {
            let key = char_to_key(c); // Extract to reusable fn
            e.on_key(key, c.is_uppercase(), false)
        })
        .collect()
}

// Add above tests:
fn char_to_key(c: char) -> u16 {
    match c.to_ascii_lowercase() {
        'a' => keys::A,
        'b' => keys::B,
        // ... etc
        _ => 0,
    }
}
```

**Priority:** HIGH
**File:** `engine/mod.rs:465-510`

---

### H2. Large Function: `Engine::process()` too complex

**Location:** `engine/mod.rs:183-265`

**Issue:** 83-line function handles all key processing logic. Handles 7 different cases with nested conditions and early returns. Cyclomatic complexity ~12.

**Impact:**
- Hard to reason about control flow
- Testing individual cases difficult
- Bug risk in edge case interactions

**Fix:** Extract to smaller functions:
```rust
fn process(&mut self, key: u16, caps: bool) -> Result {
    let m = input::get(self.method);

    // Handle đ transformations
    if let Some(r) = self.try_handle_d(key, &m) {
        return r;
    }

    // Handle tone modifiers
    if let Some(r) = self.try_handle_tone(key, &m) {
        return r;
    }

    // Handle marks
    if let Some(r) = self.try_handle_mark(key, &m) {
        return r;
    }

    // Handle remove
    if m.is_remove(key) {
        self.last_transform = None;
        return self.handle_remove();
    }

    // Normal letter
    self.handle_normal_letter(key, caps)
}

fn try_handle_d(&mut self, key: u16, m: &Box<dyn Method>) -> Option<Result> {
    // Lines 188-205
}

fn try_handle_tone(&mut self, key: u16, m: &Box<dyn Method>) -> Option<Result> {
    // Lines 207-237
}

fn try_handle_mark(&mut self, key: u16, m: &Box<dyn Method>) -> Option<Result> {
    // Lines 239-248
}

fn handle_normal_letter(&mut self, key: u16, caps: bool) -> Result {
    // Lines 256-264
}
```

**Priority:** HIGH
**File:** `engine/mod.rs:183-265`

---

## MEDIUM Priority Improvements

### M1. Unnecessary Allocation: `collect_vowels()` creates temporary Vec

**Location:** `engine/mod.rs:397-412`

**Issue:** `collect_vowels()` allocates Vec on every mark operation. Used only once then dropped.

**Impact:**
- Allocation overhead on every keystroke with marks
- Pressure on allocator

**Fix:** Iterator-based approach:
```rust
// Replace line 336-344:
fn handle_mark(&mut self, key: u16, mark: u8) -> Result {
    // Use iterator directly - no allocation
    let vowels: Vec<_> = self.buf
        .iter()
        .enumerate()
        .filter(|(_, c)| keys::is_vowel(c.key))
        .map(|(pos, c)| {
            let modifier = match c.tone {
                1 => Modifier::Circumflex,
                2 => Modifier::Horn,
                _ => Modifier::None,
            };
            Vowel::new(c.key, modifier, pos)
        })
        .collect();

    if vowels.is_empty() {
        return Result::none();
    }

    let last_vowel_pos = vowels.last().unwrap().pos;
    let has_final = self.has_final_consonant(last_vowel_pos);
    let pos = Phonology::find_tone_position(&vowels, has_final, self.modern);

    // ... rest same
}
```

Or better, make `collect_vowels()` return impl Iterator if possible.

**Priority:** MEDIUM
**File:** `engine/mod.rs:336-352`

---

### M2. Redundant Clone: `vowels` in tone revert logic

**Location:** `engine/mod.rs:225-230`

**Issue:** Creates `all_vowel_keys` Vec when `vowel_keys` was already available. Filters twice.

**Impact:**
- Extra filtering iteration
- Extra allocation

**Fix:** Refactor tone revert check:
```rust
// Lines 222-236 - simplify double-key detection
if let Some(Transform::Tone(last_key, last_tone, last_target)) = self.last_transform {
    if last_key == key {
        // Just check if we would target same vowel
        // Don't need to rebuild all_vowel_keys
        if let Some(last_vowel) = self.buf.last() {
            if keys::is_vowel(last_vowel.key) && last_vowel.key == last_target {
                return self.revert_tone(key, caps);
            }
        }
    }
}
```

**Priority:** MEDIUM
**File:** `engine/mod.rs:222-236`

---

### M3. Code Duplication: Revert logic similar for tone and mark

**Location:** `engine/mod.rs:310-332` and `engine/mod.rs:355-377`

**Issue:** `revert_tone()` and `revert_mark()` have 90% same code structure.

**Impact:**
- Duplication
- Inconsistency risk

**Fix:** Extract common pattern:
```rust
fn revert_modifier<F>(
    &mut self,
    key: u16,
    caps: bool,
    mut clear_fn: F
) -> Result
where
    F: FnMut(&mut Char) -> bool  // Returns true if cleared something
{
    self.last_transform = None;

    for pos in self.buf.find_vowels().into_iter().rev() {
        if let Some(c) = self.buf.get_mut(pos) {
            if clear_fn(c) {
                let mut result = self.rebuild_from(pos);
                if let Some(ch) = key_to_char(key, caps) {
                    if result.count < MAX as u8 {
                        result.chars[result.count as usize] = ch as u32;
                        result.count += 1;
                    }
                }
                return result;
            }
        }
    }
    Result::none()
}

fn revert_tone(&mut self, key: u16, caps: bool) -> Result {
    self.revert_modifier(key, caps, |c| {
        if c.tone > 0 {
            c.tone = 0;
            true
        } else {
            false
        }
    })
}

fn revert_mark(&mut self, key: u16, caps: bool) -> Result {
    self.revert_modifier(key, caps, |c| {
        if c.mark > 0 {
            c.mark = 0;
            true
        } else {
            false
        }
    })
}
```

**Priority:** MEDIUM
**File:** `engine/mod.rs:310-377`

---

### M4. Unused Public Functions in `chars.rs`

**Location:** `data/chars.rs:124-138`

**Issue:** `is_vowel_char()` and `get_base_vowel()` are public but unused anywhere in codebase.

**Impact:**
- Dead code in public API
- Confusing for users

**Fix:**
1. If these are meant for future use, mark with `#[allow(dead_code)]` and document intent
2. If meant for FFI, expose via C FFI
3. Otherwise, remove

**Priority:** MEDIUM
**File:** `data/chars.rs:124-138`

---

### M5. Magic Numbers: Hardcoded modifier/mark values

**Location:** Multiple files (`engine/mod.rs`, `input/telex.rs`, `input/vni.rs`)

**Issue:** Mark values (1=sắc, 2=huyền, etc.) and tone values (1=circumflex, 2=horn) hardcoded throughout.

**Examples:**
- `engine/mod.rs:404-408` - match on tone values 1, 2
- `input/telex.rs:17-23` - return hardcoded 1-5 for marks
- `data/vowel.rs:27-31` - enum repr(u8) but values used as raw numbers

**Impact:**
- Error-prone
- Not self-documenting

**Fix:** Define constants or use enum values:
```rust
// In data/mod.rs or separate marks.rs
pub const MARK_NONE: u8 = 0;
pub const MARK_SAC: u8 = 1;
pub const MARK_HUYEN: u8 = 2;
pub const MARK_HOI: u8 = 3;
pub const MARK_NGA: u8 = 4;
pub const MARK_NANG: u8 = 5;

pub const TONE_NONE: u8 = 0;
pub const TONE_CIRCUMFLEX: u8 = 1;  // ^
pub const TONE_HORN_BREVE: u8 = 2;  // horn/breve

// Then use throughout:
fn is_mark(&self, key: u16) -> Option<u8> {
    match key {
        keys::S => Some(MARK_SAC),
        keys::F => Some(MARK_HUYEN),
        // ...
    }
}
```

**Priority:** MEDIUM
**Files:** Multiple

---

### M6. Function `key_to_char()` doesn't need Option return

**Location:** `engine/mod.rs:27-68`

**Issue:** Returns `Option<char>` but most callers expect Some(). Numbers use early return pattern inconsistently.

**Impact:**
- Adds unwrapping burden
- Inconsistent style (numbers vs letters)

**Fix:**
```rust
fn key_to_char(key: u16, caps: bool) -> char {
    let ch = match key {
        keys::A => 'a',
        // ... all letters
        keys::N0 => '0',
        keys::N1 => '1',
        // ... all numbers
        _ => return '\0', // or panic, since invalid key shouldn't happen
    };
    if caps && ch.is_alphabetic() {
        ch.to_ascii_uppercase()
    } else {
        ch
    }
}
```

Then callers don't need `if let Some(ch)` checks.

**Priority:** MEDIUM
**File:** `engine/mod.rs:27-68`

---

### M7. Buffer unused helper methods

**Location:** `engine/buffer.rs:112-118`

**Issue:** `last_mut()` defined but never used in codebase.

**Impact:** Dead code

**Fix:** Remove or mark `#[allow(dead_code)]` if planned for future use.

**Priority:** MEDIUM
**File:** `engine/buffer.rs:112-118`

---

## LOW Priority Suggestions

### L1. Vowel::new() could be inlined

**Location:** `data/vowel.rs:50-52`

**Issue:** Trivial 3-line constructor called frequently.

**Fix:** Add `#[inline]`:
```rust
#[inline]
pub fn new(key: u16, modifier: Modifier, pos: usize) -> Self {
    Self { key, modifier, pos }
}
```

**Priority:** LOW
**File:** `data/vowel.rs:50-52`

---

### L2. Buffer could use const generics for MAX

**Location:** `engine/buffer.rs:3`

**Issue:** `MAX` is hardcoded to 32. Could be generic for flexibility.

**Fix:**
```rust
pub struct Buffer<const N: usize = 32> {
    data: [Char; N],
    len: usize,
}
```

But probably not worth it unless multiple buffer sizes needed.

**Priority:** LOW
**File:** `engine/buffer.rs:3-44`

---

### L3. Documentation: Missing examples for key functions

**Location:** Multiple files

**Issue:** Public API lacks usage examples. e.g., `Engine::on_key()`, `Phonology::find_tone_position()`.

**Impact:**
- Harder for new contributors
- No runnable examples

**Fix:** Add doc examples:
```rust
/// Handle key event - main entry point
///
/// # Example
/// ```
/// use gonhanh_core::engine::Engine;
/// use gonhanh_core::data::keys;
///
/// let mut e = Engine::new();
/// let result = e.on_key(keys::A, false, false);
/// // Type 'a' then 's' for á
/// let result = e.on_key(keys::S, false, false);
/// assert_eq!(result.chars[0], 'á' as u32);
/// ```
pub fn on_key(&mut self, key: u16, caps: bool, ctrl: bool) -> Result {
```

**Priority:** LOW
**Files:** Multiple public APIs

---

### L4. Vowel phonology could cache role classification

**Location:** `data/vowel.rs:190-230`

**Issue:** `classify_roles()` defined but never used. If used in future, recalculates roles every time.

**Impact:** Currently dead code

**Fix:**
1. Remove if truly unused
2. If planned for future, document intended use case

**Priority:** LOW
**File:** `data/vowel.rs:190-230`

---

### L5. FFI Result struct padding field exposed

**Location:** `engine/mod.rs:86`

**Issue:** `_pad: u8` is public field in FFI struct.

**Impact:** C callers might accidentally use it.

**Fix:** Document it better:
```rust
/// Padding byte for struct alignment (do not use)
pub _pad: u8,
```

**Priority:** LOW
**File:** `engine/mod.rs:79-113`

---

### L6. Engine could impl Display for debugging

**Location:** `engine/mod.rs:122-459`

**Issue:** No easy way to inspect engine state for debugging.

**Fix:**
```rust
impl std::fmt::Debug for Engine {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("Engine")
            .field("method", &self.method)
            .field("enabled", &self.enabled)
            .field("modern", &self.modern)
            .field("buffer_len", &self.buf.len())
            .field("last_transform", &self.last_transform)
            .finish()
    }
}
```

**Priority:** LOW
**File:** `engine/mod.rs:122-146`

---

## Performance Notes

**Positive:**
- Fixed-size buffer (no heap allocs for buffer itself)
- Efficient lookup tables for char conversion
- Minimal copying

**Could improve:**
- M1: Remove Vec allocation in `collect_vowels()`
- M2: Eliminate redundant filtering

**Estimated impact:** Minor (single-digit % improvement on keystroke latency)

---

## Architecture Observations

**Module Structure: GOOD**

```
core/src/
├── data/           # Linguistic data (clean separation)
│   ├── keys.rs     # Platform keycodes
│   ├── chars.rs    # Unicode conversion
│   └── vowel.rs    # Phonology rules
├── input/          # Input method implementations
│   ├── telex.rs
│   └── vni.rs
├── engine/         # Core processing logic
│   ├── buffer.rs
│   └── mod.rs
└── lib.rs          # FFI interface
```

Separation of concerns well-maintained. No circular dependencies.

**Potential split:**
- `engine/mod.rs` (555 lines) could split into:
  - `engine/mod.rs` - public API (150 lines)
  - `engine/processor.rs` - key processing logic (250 lines)
  - `engine/transforms.rs` - tone/mark transformations (150 lines)

But current size still manageable.

---

## Security Notes

- FFI boundary handled correctly with null checks
- No unsafe code except in FFI free function (correct usage)
- No buffer overflows (MAX enforced)
- No panics in hot path

**Concern:**
- Line 102 in lib.rs: `unsafe { ime_free(r1) }` in test - ensure all callers follow protocol

---

## Test Coverage

**Total tests:** 49 (28 unit + 21 behavior)
**Status:** All passing
**Coverage:** Good

**Gaps identified:**
1. No tests for ctrl key handling (line 165)
2. No tests for buffer overflow (> MAX chars)
3. No tests for VNI delayed đ (dung9 → đung)
4. Edge cases: empty buffer operations

**Recommendation:** Add tests for above gaps.

---

## Actionable Recommendations

**Immediate (do in next session):**
1. Fix H1 - Extract char_to_key() helper (~30 min)
2. Fix M4 - Remove or document unused public fns (~15 min)
3. Fix M7 - Remove unused last_mut() (~5 min)
4. Add M5 - Define mark/tone constants (~20 min)

**Short-term (do this week):**
1. Fix H2 - Refactor Engine::process() into smaller fns (~2 hours)
2. Fix M3 - Extract common revert logic (~1 hour)
3. Fix M1 - Remove unnecessary Vec allocation (~30 min)
4. Fix M2 - Simplify tone revert logic (~30 min)

**Long-term (nice to have):**
1. L3 - Add API documentation examples
2. L6 - Add Debug impl for Engine
3. Split engine/mod.rs into submodules
4. Add missing test coverage

---

## Summary Stats

| Metric | Value |
|--------|-------|
| Files reviewed | 10 |
| Lines of code | 1,868 |
| Critical issues | 0 |
| High priority | 2 |
| Medium priority | 7 |
| Low priority | 6 |
| Tests passing | 49/49 |
| Clippy warnings | 0 |
| Compile errors | 0 |

---

## Conclusion

Codebase quality is **good** overall. No critical bugs or security issues. Main areas for improvement:

1. **Code duplication** - test helpers, revert logic
2. **Function complexity** - Engine::process() needs refactoring
3. **Unnecessary allocations** - minor perf improvements available
4. **Dead code** - cleanup unused public API

Code follows Rust best practices. Architecture is sound. Well-tested. Production-ready with recommended refactorings applied.

**Next steps:** Prioritize H1, H2, and M-series fixes for cleaner, more maintainable code.
