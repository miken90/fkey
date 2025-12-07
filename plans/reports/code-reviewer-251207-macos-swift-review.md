# Code Review: GoNhanh macOS Swift Platform

**Date**: 2025-12-07
**Reviewer**: Code Review Agent
**Scope**: Swift macOS platform (App.swift, MenuBar.swift, SettingsView.swift, RustBridge.swift)

---

## Scope

**Files reviewed**:
- `platforms/macos/App.swift` (29 lines)
- `platforms/macos/MenuBar.swift` (152 lines)
- `platforms/macos/SettingsView.swift` (77 lines)
- `platforms/macos/RustBridge.swift` (430 lines)

**Total lines**: ~688 lines
**Review focus**: Recent changes to RustBridge.swift (40 modified lines), code quality, architecture, FFI safety, error handling, maintainability
**Recent changes**: Improved memory safety (defer for ime_free), added DEBUG-only logging, extracted KeyCode enum, removed unused proxy parameter

---

## Overall Assessment

**Code is production-ready with strong fundamentals**. Swift code is clean, idiomatic, and shows solid understanding of macOS platform APIs. FFI bridge demonstrates excellent memory safety practices. Architecture is well-separated with clear responsibilities. Recent refactorings show attention to safety and code quality.

**Key strengths**: Memory safety, FFI design, app lifecycle management, platform-specific workarounds
**Key gaps**: Settings persistence, error recovery, testability, state synchronization

---

## DX Scores by Category

### 1. Code Quality & Idioms: **8/10**

**Strengths**:
- Idiomatic Swift with proper use of optional chaining, guards, defer
- Clean separation with MARK comments
- Proper use of SwiftUI property wrappers (@State, @NSApplicationDelegateAdaptor)
- Follows Swift API design guidelines (clear naming, parameter labels)
- Recent fix: `defer { ime_free(resultPtr) }` prevents use-after-free (RustBridge.swift:90)

**Issues**:
- MenuBar.swift:90 - Magic index `item(at: 0)` fragile, use `item(withTitle:)` instead
- MenuBar.swift:119-120 - Magic indices for method menu, should query by title
- SettingsView.swift:43,68 - Uses deprecated `NSApp.keyWindow`, prefer window reference passing
- RustBridge.swift:103-107 - Complex pointer rebinding could be helper function

**Quick wins**:
```swift
// MenuBar.swift:90 - Replace magic index
if let item = statusItem.menu?.item(withTitle: "Bật GoNhanh") {
    item.state = isEnabled ? .on : .off
}

// MenuBar.swift:115 - Use tags instead of indices
telexItem.tag = 0
vniItem.tag = 1
// Then update via: methodMenu.items.first { $0.tag == currentMethod }?.state = .on
```

---

### 2. Architecture: **8.5/10**

**Strengths**:
- Clear layering: App → MenuBar → RustBridge → FFI → Rust Core
- Single responsibility: MenuBar (UI), RustBridge (FFI), KeyboardHookManager (event capture)
- Singleton pattern appropriate for KeyboardHookManager.shared
- Static class for RustBridge avoids lifecycle issues
- App.swift:19 - Correct use of `.accessory` policy for menu bar app
- MenuBar.swift:7-8 - State is private, encapsulated

**Issues**:
- MenuBar.swift:7-8 - State duplicated between MenuBar and RustBridge (isEnabled, currentMethod)
- SettingsView.swift:4-5 - State isolated, doesn't sync with MenuBar's state
- No state synchronization mechanism (SettingsView changes don't update MenuBar icon)
- RustBridge.swift:69 - `isInitialized` flag prevents re-init but not idempotent cleanup

**Recommendations**:
```swift
// Create shared state manager (ObservableObject)
class AppState: ObservableObject {
    @Published var isEnabled: Bool = true { didSet { RustBridge.setEnabled(isEnabled) } }
    @Published var inputMethod: InputMethod = .telex { didSet { RustBridge.setMethod(inputMethod.rawValue) } }
}

// Inject into MenuBar and SettingsView via @EnvironmentObject
```

---

### 3. FFI Bridge Safety: **9/10**

**Strengths**:
- RustBridge.swift:32-41 - FFI struct matches Rust `#[repr(C)]` exactly (verified padding)
- RustBridge.swift:45-64 - Proper use of `@_silgen_name` for C ABI
- RustBridge.swift:90 - **Excellent**: Uses `defer { ime_free(resultPtr) }` for guaranteed cleanup
- RustBridge.swift:103-107 - Safe pointer rebinding with `withMemoryRebound`
- RustBridge.swift:82-84 - Guards against uninitialized engine
- RustBridge.swift:110-112 - Safe Unicode scalar validation before Character creation

**Issues**:
- RustBridge.swift:49 - Returns raw pointer, no lifetime guarantees documented
- No error handling if Rust panics (undefined behavior in FFI)
- RustBridge.swift:109-113 - Silent failure if Unicode scalar is invalid (should log)

**Recommendations**:
```swift
// RustBridge.swift:109 - Log invalid scalars
for code in charArray {
    guard let scalar = Unicode.Scalar(code) else {
        debugLog("[RustBridge] Invalid Unicode scalar: 0x\(String(code, radix: 16))")
        continue
    }
    chars.append(Character(scalar))
}

// Document lifetime contract
/// Returns: Owned pointer that MUST be freed via ime_free()
/// Safety: Caller responsible for calling ime_free() exactly once
@_silgen_name("ime_key")
func ime_key(_ key: UInt16, _ caps: Bool, _ ctrl: Bool) -> UnsafeMutablePointer<ImeResult>?
```

---

### 4. Error Handling: **6/10**

**Strengths**:
- RustBridge.swift:155-167 - Graceful permission check with user prompt
- RustBridge.swift:283-289 - Handles event tap disable/timeout, auto-recovers
- RustBridge.swift:215-236 - Fallback tap creation with user guidance
- Logging in DEBUG mode (RustBridge.swift:7-28)

**Issues**:
- App.swift:26 - Silent failure if KeyboardHookManager.stop() errors
- MenuBar.swift:125-136 - Settings window creation doesn't handle hostingController failures
- RustBridge.swift:216 - No retry logic or telemetry if all event tap methods fail
- RustBridge.swift:376-428 - CGEvent creation failures silently ignored (should log)
- No user-facing error alerts for critical failures (e.g., FFI init failure)
- SettingsView.swift:61-63 - TODO for loadSettings, no fallback defaults

**Critical gaps**:
1. **No error recovery for RustBridge.initialize() failure** - app continues silently
2. **Event posting failures invisible** - user thinks IME is working but text doesn't appear
3. **No crash reporting or telemetry** - debugging production issues impossible

**Recommendations**:
```swift
// RustBridge.swift:72 - Add init error handling
static func initialize() throws {
    guard !isInitialized else { return }
    ime_init()  // If this panics, Swift can't catch it - need Rust-side error return
    isInitialized = true
    debugLog("[RustBridge] Engine initialized")
}

// RustBridge.swift:380 - Log event failures
if let down = CGEvent(keyboardEventSource: source, virtualKey: KeyCode.backspace, keyDown: true) {
    down.post(tap: .cgSessionEventTap)
} else {
    debugLog("[RustBridge] Failed to create backspace event")
}

// Show alert for critical failures
static func showCriticalError(_ message: String) {
    DispatchQueue.main.async {
        let alert = NSAlert()
        alert.messageText = "Lỗi GoNhanh"
        alert.informativeText = message
        alert.alertStyle = .critical
        alert.runModal()
    }
}
```

---

### 5. Maintainability: **7.5/10**

**Strengths**:
- RustBridge.swift:354-359 - Extracted KeyCode enum (recent refactor, good!)
- Clear MARK comments separate sections
- RustBridge.swift:324-352 - Platform-specific workarounds documented inline
- Function names are descriptive (sendTextReplacementWithSelection vs WithBackspace)
- Recent removal of unused proxy parameter shows active maintenance

**Issues**:
- MenuBar.swift:141-145 - Version hardcoded ("0.1.0"), should read from Info.plist/config
- RustBridge.swift:334-343 - App bundle ID list is hardcoded, should be configurable
- SettingsView.swift:30-31 - Hardcoded Telex/VNI descriptions, not DRY with MenuBar
- No dependency injection (all singletons/static classes = hard to test)
- Magic numbers: RustBridge.swift:32-36 (32 chars tuple), should be const

**Hard to add features**:
1. **New input method** - Requires changes in 3+ places (MenuBar, SettingsView, RustBridge)
2. **Custom app workarounds** - Must edit RustBridge source, can't load from config
3. **Settings persistence** - No storage layer, unclear where to add

**Quick wins**:
```swift
// Centralize configuration
struct AppConfig {
    static let version: String = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "Unknown"
    static let maxCharsPerEvent = 32
    static let autocompleteApps = [
        "com.google.Chrome",
        "com.microsoft.edgemac",
        // ... load from plist or UserDefaults
    ]
}

// Enum for input methods (shared)
enum InputMethod: Int, CaseIterable {
    case telex = 0
    case vni = 1

    var displayName: String {
        switch self {
        case .telex: return "Telex (aw, ow, w, s, f, r, x, j)"
        case .vni: return "VNI (a8, o9, 1-5)"
        }
    }
}
```

---

### 6. Testability: **5/10**

**Strengths**:
- Pure functions exist: needsSelectionWorkaround() could be unit tested
- FFI is separated from UI logic
- RustBridge methods are stateless (static)

**Critical issues**:
- **No tests found** in reviewed files
- Singletons/statics prevent dependency injection
- CGEvent creation uses system APIs (not mockable)
- No protocols/interfaces - tight coupling to concrete types
- MenuBar couples UI to business logic (toggleEnabled modifies RustBridge directly)
- SettingsView.swift:72-76 - Preview provider exists but doesn't test behavior

**Cannot test**:
1. Keyboard event processing (uses real CGEvent tap)
2. Menu bar UI updates (tightly coupled to NSStatusItem)
3. Settings persistence (not implemented, but will be hard to test when added)
4. Error recovery paths (no injection points for failures)

**Recommendations**:
```swift
// Introduce protocols for testing
protocol ImeEngine {
    func processKey(keyCode: UInt16, caps: Bool, ctrl: Bool) -> (Int, [Character])?
    func setMethod(_ method: Int)
    func setEnabled(_ enabled: Bool)
}

class RustBridge: ImeEngine {
    // Existing implementation
}

class MockImeEngine: ImeEngine {
    var shouldReturn: (Int, [Character])? = nil
    func processKey(keyCode: UInt16, caps: Bool, ctrl: Bool) -> (Int, [Character])? {
        return shouldReturn
    }
}

// MenuBar accepts engine via init
class MenuBarController {
    private let engine: ImeEngine
    init(engine: ImeEngine = RustBridge.self) { ... }
}
```

---

## Specific Issues Found

### Critical (Security/Crashes)
**None found** - Recent `defer` fix addressed main memory safety concern.

### High Priority

1. **State Synchronization Broken** (MenuBar.swift:7-8, SettingsView.swift:4-5)
   - Changing settings in SettingsView doesn't update MenuBar icon
   - MenuBar state changes don't update SettingsView if reopened
   - **Impact**: User confusion, settings appear to not save
   - **Fix**: Introduce shared ObservableObject state manager

2. **No Settings Persistence** (SettingsView.swift:61-63)
   - TODO comment confirms not implemented
   - Settings reset on app restart
   - **Impact**: Poor UX, violates "Smart Defaults" philosophy
   - **Fix**: UserDefaults storage + load on init
   ```swift
   func loadSettings() {
       enabled = UserDefaults.standard.bool(forKey: "ime.enabled")
       let method = UserDefaults.standard.integer(forKey: "ime.method")
       mode = InputMethod(rawValue: method) ?? .telex
   }
   ```

3. **Event Posting Failures Silent** (RustBridge.swift:376-428)
   - If CGEvent creation/posting fails, no logging or recovery
   - **Impact**: User types but no text appears, debuggability near-zero
   - **Fix**: Add DEBUG logging for all nil checks

### Medium Priority

4. **Magic Indices in Menu Updates** (MenuBar.swift:90, 119-120)
   - Fragile if menu structure changes
   - **Impact**: Menu state updates break if items reordered
   - **Fix**: Use tags or title-based lookup

5. **Hardcoded Version String** (MenuBar.swift:141)
   - Not synced with actual bundle version
   - **Impact**: About dialog shows wrong version
   - **Fix**: Read from `Bundle.main.infoDictionary["CFBundleShortVersionVersion"]`

6. **No Error Handling for FFI Init** (RustBridge.swift:72-77)
   - If Rust panics during ime_init(), Swift can't catch it
   - **Impact**: Undefined behavior, possible crash
   - **Fix**: Rust should return error code, Swift checks it

7. **App Bundle List Hardcoded** (RustBridge.swift:334-343)
   - Adding new app requires source edit
   - **Impact**: Poor extensibility, harder for users to customize
   - **Fix**: Load from UserDefaults or plist

### Low Priority

8. **Deprecated NSApp.keyWindow** (SettingsView.swift:43, 68)
   - Works but deprecated in macOS 13+
   - **Impact**: Future compatibility risk
   - **Fix**: Pass window reference via environment or callback

9. **Complex Pointer Rebinding** (RustBridge.swift:103-107)
   - Correct but hard to verify at a glance
   - **Impact**: Maintenance burden
   - **Fix**: Extract to well-documented helper function

10. **Invalid Unicode Scalar Silent Failure** (RustBridge.swift:109-113)
    - Skips invalid scalars without logging
    - **Impact**: Debuggability for corrupted FFI data
    - **Fix**: Add debugLog for skipped scalars

---

## Positive Observations

1. **Excellent FFI Memory Safety** - defer pattern in RustBridge.swift:90 is textbook-correct
2. **Platform Workarounds Well-Documented** - Chrome/Excel autocomplete fix is clearly explained (RustBridge.swift:309-313, 363-372)
3. **Permission Handling UX** - Proactive accessibility prompt with clear instructions (RustBridge.swift:219-234)
4. **Event Tap Fallback Strategy** - Tries 3 different tap types for maximum compatibility (RustBridge.swift:182-213)
5. **Clean Code Structure** - MARK comments, private visibility, clear naming throughout
6. **Recent Improvements Show Care** - DEBUG-only logging, KeyCode enum, defer cleanup all demonstrate active quality focus

---

## Recommended Actions (Prioritized by Impact)

### Immediate (1-2 hours)

1. **Add Settings Persistence** (SettingsView.swift:61-69)
   ```swift
   func loadSettings() {
       enabled = UserDefaults.standard.bool(forKey: "gonhanh.enabled")
       let method = UserDefaults.standard.integer(forKey: "gonhanh.method")
       mode = InputMode(rawValue: method) ?? .telex
   }
   func saveSettings() {
       UserDefaults.standard.set(enabled, forKey: "gonhanh.enabled")
       UserDefaults.standard.set(mode.rawValue, forKey: "gonhanh.method")
       RustBridge.setEnabled(enabled)
       RustBridge.setMethod(mode.rawValue)
       NSApp.keyWindow?.close()
   }
   ```

2. **Fix Magic Indices** (MenuBar.swift:90, 119-120)
   - Use `.item(withTitle:)` or tag-based lookup

3. **Add Event Failure Logging** (RustBridge.swift:380, 392, 407, 421)
   ```swift
   if let down = CGEvent(...) {
       down.post(...)
   } else {
       debugLog("[Send] Failed to create event: backspace")
   }
   ```

4. **Fix Hardcoded Version** (MenuBar.swift:141)
   ```swift
   .applicationVersion: Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "Unknown"
   ```

### Short-term (1 week)

5. **Implement Shared State Manager**
   - Create `AppState: ObservableObject`
   - Inject into MenuBar and SettingsView via `@EnvironmentObject`
   - Synchronizes enabled/method state across UI and engine

6. **Make App Bundle List Configurable**
   - Move to UserDefaults with default list
   - Add UI in SettingsView to add/remove apps

7. **Extract Config Constants**
   - Create `AppConfig` struct for version, max chars, bundle IDs
   - Replace magic numbers throughout

8. **Add FFI Error Handling**
   - Modify Rust FFI to return error codes
   - Check in Swift and show user-facing alerts

### Medium-term (1 month)

9. **Improve Testability**
   - Introduce `ImeEngine` protocol
   - Extract pure business logic from MenuBar
   - Add unit tests for core functions

10. **Add Telemetry/Crash Reporting** (Optional)
    - Track FFI failures, event tap errors
    - Help debug production issues

11. **Replace Deprecated APIs**
    - Fix `NSApp.keyWindow` usage
    - Audit for other deprecated macOS APIs

---

## Metrics

- **Type Safety**: Excellent (Swift's type system enforced, FFI carefully managed)
- **Memory Safety**: Excellent (defer cleanup, safe pointer handling)
- **Test Coverage**: 0% (no tests found)
- **TODOs**: 1 (SettingsView.swift:62 - loadSettings)
- **Deprecated APIs**: 2 (NSApp.keyWindow usage)
- **Magic Numbers**: ~5 (menu indices, keycode 0x33/0x7B now fixed)

---

## Unresolved Questions

1. **FFI Panic Handling**: What happens if Rust panics in ime_init/ime_key? Can we catch it in Swift?
2. **Modern Orthography**: When is RustBridge.setModern() called? No UI for this setting.
3. **Clear Buffer on Focus Change**: Should clearBuffer() be called automatically when switching apps?
4. **Multiple Keyboards**: How does this handle users with multiple physical keyboards (different layouts)?
5. **Testing Strategy**: Should we use XCTest, or integration tests via UI automation?

---

## Final Verdict

**Production-ready with known gaps**. Code demonstrates strong engineering: excellent FFI safety, thoughtful platform workarounds, clean architecture. Recent improvements show active maintenance. Main gaps are settings persistence (UX critical) and state synchronization (causes confusion). No blocking bugs found.

**Recommended before public release**:
1. Settings persistence (immediate)
2. State synchronization (short-term)
3. Add basic error logging (immediate)

**Overall DX Score: 7.5/10** (would be 8.5/10 with settings persistence + state sync)
