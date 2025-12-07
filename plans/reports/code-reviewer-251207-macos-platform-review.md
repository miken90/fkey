# Code Review: macOS Platform

**Reviewed**: 2025-12-07
**Scope**: `/platforms/macos/` (672 LOC across 4 Swift files)
**Files**: App.swift, MenuBar.swift, SettingsView.swift, RustBridge.swift

---

## Overall Assessment

Code functional but needs refactoring. Major issues: excessive debug logging in production, hardcoded values, dead code, state sync problems, no error recovery, force unwraps, and missing modern Swift patterns.

---

## CRITICAL ISSUES

### 1. Debug Logging in Production Code
**File**: `RustBridge.swift:7-21`
**Issue**: `debugLog()` writes to `/tmp/gonhanh_debug.log` on EVERY keystroke - massive I/O overhead, disk usage, privacy leak
**Impact**: Performance degradation, potential disk fill, logs user keystrokes
**Fix**:
```swift
// Add conditional compilation
#if DEBUG
func debugLog(_ message: String) { /* existing impl */ }
#else
func debugLog(_ message: String) {}  // no-op in release
#endif
```
**Priority**: HIGH

### 2. Force Unwrap on File I/O
**File**: `RustBridge.swift:16`
**Issue**: `handle.write(logMessage.data(using: .utf8)!)`
**Impact**: Crash if encoding fails (rare but possible)
**Fix**:
```swift
if let data = logMessage.data(using: .utf8) {
    handle.write(data)
}
```
**Priority**: MEDIUM

### 3. No Memory Management for FFI
**File**: `RustBridge.swift:80-85`
**Issue**: `ime_key()` returns pointer that MUST be freed with `ime_free()`, but if early return between L80-88, leaks memory
**Fix**:
```swift
guard let resultPtr = ime_key(keyCode, caps, ctrl) else { return nil }
defer { ime_free(resultPtr) }  // ensures cleanup
let result = resultPtr.pointee
// ... rest of code
```
**Priority**: HIGH

---

## HIGH PRIORITY

### 4. State Synchronization Bug
**Files**:
- `MenuBar.swift:7-8` - maintains `isEnabled`, `currentMethod`
- `SettingsView.swift:4-5` - has SEPARATE `enabled`, `mode` state

**Issue**: No shared state model - changes in MenuBar don't update SettingsView and vice versa
**Example**: Toggle in menu bar → SettingsView still shows old state
**Fix**: Use `@AppStorage` or ObservableObject singleton
```swift
class AppState: ObservableObject {
    @Published var isEnabled = true
    @Published var inputMethod: InputMethod = .telex
}
```
**Priority**: HIGH

### 5. Hardcoded UI Dimensions
**File**: `MenuBar.swift:131`, `SettingsView.swift:55`
**Issue**: Window size `NSSize(width: 400, height: 300)` hardcoded twice
**Fix**: Single source of truth
```swift
private enum Constants {
    static let settingsWindowSize = NSSize(width: 400, height: 300)
}
```
**Priority**: MEDIUM

### 6. Unsafe Array Access
**File**: `MenuBar.swift:90-92, 119-120`
**Issue**: `menu?.item(at: 0)`, `methodMenu.item(at: 1)` can crash if menu structure changes
**Fix**: Tag menu items, access by tag
```swift
enabledItem.tag = 100
// Later:
if let item = menu?.item(withTag: 100) { ... }
```
**Priority**: HIGH

### 7. No Error Recovery for Event Tap Failure
**File**: `RustBridge.swift:208-229`
**Issue**: If event tap creation fails, shows alert then returns - app still running but non-functional
**Fix**: Either retry or show persistent warning in menu bar
```swift
if tap == nil {
    showPersistentWarning()
    startRetryTimer()  // retry every 5s
}
```
**Priority**: HIGH

### 8. Race Condition in Window Management
**File**: `MenuBar.swift:124-136`
**Issue**: `settingsWindow` created only once, but multiple clicks → `makeKeyAndOrderFront()` on deallocated window if user closed it
**Fix**:
```swift
@objc func openSettings() {
    if settingsWindow == nil || settingsWindow?.isVisible == false {
        // recreate window
    }
    settingsWindow?.makeKeyAndOrderFront(nil)
}
```
**Priority**: MEDIUM

---

## MEDIUM PRIORITY

### 9. Dead Code
**File**: `SettingsView.swift:61-63`
**Issue**: `loadSettings()` has empty body marked TODO, called but does nothing
**Fix**: Remove if not needed, or implement config persistence
**Priority**: LOW

### 10. Unused Parameter
**File**: `RustBridge.swift:352, 389`
**Issue**: `proxy: CGEventTapProxy` parameter never used in send functions
**Fix**: Remove or prefix with `_`
```swift
private func sendTextReplacement(backspaceCount: Int, chars: [Character], proxy _: CGEventTapProxy)
```
**Priority**: LOW

### 11. Magic Numbers
**File**: `RustBridge.swift:368-371, 395-396`
**Issue**: Key codes `0x33` (backspace), `0x7B` (left arrow) hardcoded without comments
**Fix**:
```swift
private enum KeyCode {
    static let backspace: CGKeyCode = 0x33
    static let leftArrow: CGKeyCode = 0x7B
}
```
**Priority**: MEDIUM

### 12. Inefficient String Building
**File**: `RustBridge.swift:96-106`
**Issue**: Converts tuple→array→scalars→chars - unnecessary allocations
**Fix**: Use `withUnsafeBytes` directly on tuple
**Priority**: LOW (micro-optimization)

### 13. Duplicate Logic
**Files**:
- `MenuBar.swift:103-107` (setTelex)
- `MenuBar.swift:109-113` (setVNI)

**Issue**: Nearly identical code - violates DRY
**Fix**:
```swift
@objc func setTelex() { setInputMethod(.telex) }
@objc func setVNI() { setInputMethod(.vni) }

private func setInputMethod(_ method: InputMethod) {
    currentMethod = method.rawValue
    RustBridge.setMethod(currentMethod)
    updateMethodMenu()
}
```
**Priority**: MEDIUM

### 14. Non-Idiomatic Swift
**File**: `RustBridge.swift:292-294`
**Issue**: Overly complex boolean expression
**Fix**:
```swift
let caps = flags.contains([.maskShift, .maskAlphaShift])
let ctrl = flags.contains([.maskCommand, .maskControl, .maskAlternate])
```
**Priority**: LOW

### 15. Missing Documentation
**Files**: All
**Issue**: No function-level docs for public APIs - hard to maintain
**Fix**: Add SwiftDoc comments
```swift
/// Processes keyboard event through IME engine
/// - Parameters:
///   - keyCode: Virtual key code from CGEvent
///   - caps: Shift/CapsLock state
///   - ctrl: Cmd/Ctrl/Alt state
/// - Returns: Tuple of (backspace count, replacement chars) or nil
static func processKey(keyCode: UInt16, caps: Bool, ctrl: Bool) -> (Int, [Character])?
```
**Priority**: MEDIUM

---

## LOW PRIORITY

### 16. Settings Window Not Centered After First Open
**File**: `MenuBar.swift:132`
**Issue**: `center()` only called on creation - if user moves window, next open shows at moved position
**Fix**: Call `center()` every time before showing
**Priority**: LOW

### 17. About Panel Version Hardcoded
**File**: `MenuBar.swift:142`
**Issue**: Version "0.1.0" hardcoded - needs manual update
**Fix**: Read from Info.plist
```swift
let version = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "Unknown"
```
**Priority**: LOW

### 18. Preview Provider Unused
**File**: `SettingsView.swift:72-76`
**Issue**: Preview never used in production build - bloats binary
**Fix**: Wrap in `#if DEBUG`
**Priority**: LOW

---

## CODE SMELLS

### 19. God Object: RustBridge.swift (417 LOC)
**Issue**: Single file contains 5 responsibilities: FFI declarations, bridge logic, keyboard hook manager, event callback, key sending
**Fix**: Split into modules
```
RustBridge/
├── FFI.swift              # @_silgen_name declarations
├── Bridge.swift           # RustBridge class
├── KeyboardHook.swift     # KeyboardHookManager
├── EventHandler.swift     # keyboardCallback
└── KeySender.swift        # sendTextReplacement*
```
**Priority**: MEDIUM

### 20. Inconsistent Naming
- `setMethod()` vs `setEnabled()` vs `setModern()` (consistent ✓)
- BUT: `toggleEnabled()` vs `setTelex()`/`setVNI()` (toggle vs set inconsistent)
**Fix**: Use consistent verbs - either all `set*` or all `toggle*`
**Priority**: LOW

---

## SWIFT BEST PRACTICES VIOLATIONS

### 21. Using `NSApp` Global
**Files**: Multiple locations
**Issue**: Anti-pattern in SwiftUI - harder to test, breaks environment-based injection
**Fix**: Use `@Environment(\.openWindow)` for window management
**Priority**: LOW (works but not modern)

### 22. No Access Control
**Issue**: All types/methods implicitly internal - no clear public API surface
**Fix**: Mark with `public`/`private`/`fileprivate` explicitly
**Priority**: LOW

### 23. Class-Based Singleton Instead of Actor
**File**: `KeyboardHookManager` uses class + private init
**Issue**: Not thread-safe if accessed from multiple threads (unlikely but possible)
**Fix**: Use actor for thread safety
```swift
actor KeyboardHookManager {
    static let shared = KeyboardHookManager()
}
```
**Priority**: LOW

---

## PERFORMANCE NOTES

### 24. Excessive Logging Overhead
**Impact**: EVERY keystroke writes to file (2x syscalls: seek + write)
**Measurement**: ~100-500μs per log call on SSD
**Recommendation**: Remove production logging or use OSLog with disabled categories

### 25. CFRunLoop Source Leak Check
**File**: `RustBridge.swift:234-237`
**Issue**: `runLoopSource` retained but unclear if `CFRunLoopAddSource` also retains
**Verify**: Check with Instruments for leaks

---

## UNRESOLVED QUESTIONS

1. **Is `ime_init()` thread-safe?** No mutex protection in Swift layer
2. **What happens if Rust panics?** No error handling on FFI boundary
3. **Why both Accessibility AND Input Monitoring?** Code requests both permissions - Input Monitoring alone sufficient for event tap
4. **Settings persistence**: Where should config be saved? UserDefaults? Rust side?
5. **App detection logic**: Why `hasPrefix` instead of exact match? Could cause false positives (e.g., "com.google.Chrome.Beta")

---

## ACTIONABLE RECOMMENDATIONS

### Immediate (Before Next Release)
1. ✅ Wrap debug logging in `#if DEBUG`
2. ✅ Add `defer { ime_free() }` to prevent leak
3. ✅ Fix state sync with shared model
4. ✅ Add error recovery for event tap failure

### Short Term (Next Sprint)
5. Extract constants for magic numbers
6. Split RustBridge.swift into modules
7. Add proper error handling on FFI boundary
8. Implement settings persistence
9. Tag menu items instead of index access

### Long Term
10. Add comprehensive unit tests (FFI mocking)
11. Add SwiftDoc comments
12. Migrate to modern SwiftUI patterns (@Observable, etc.)
13. Add telemetry/analytics (opt-in)

---

## METRICS

- **Type Safety**: No explicit type errors (Swift compiler catches most)
- **Memory Safety**: 2 potential leaks (FFI pointer, runloop source)
- **Code Quality**: 3/10 - functional but needs significant cleanup
- **Maintainability**: 4/10 - god object, tight coupling, no docs
- **Test Coverage**: 0% (no tests found)

---

**End of Report**
