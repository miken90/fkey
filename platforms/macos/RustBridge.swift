import Foundation
import Carbon

// MARK: - FFI Result Struct (must match Rust #[repr(C)])

struct ImeResult {
    var chars: (UInt32, UInt32, UInt32, UInt32, UInt32, UInt32, UInt32, UInt32,
                UInt32, UInt32, UInt32, UInt32, UInt32, UInt32, UInt32, UInt32,
                UInt32, UInt32, UInt32, UInt32, UInt32, UInt32, UInt32, UInt32,
                UInt32, UInt32, UInt32, UInt32, UInt32, UInt32, UInt32, UInt32)
    var action: UInt8      // 0=None, 1=Send, 2=Restore
    var backspace: UInt8
    var count: UInt8
    var _pad: UInt8
}

// MARK: - C Function Declarations

@_silgen_name("ime_init")
func ime_init()

@_silgen_name("ime_key")
func ime_key(_ key: UInt16, _ caps: Bool, _ ctrl: Bool) -> UnsafeMutablePointer<ImeResult>?

@_silgen_name("ime_method")
func ime_method(_ method: UInt8)

@_silgen_name("ime_enabled")
func ime_enabled(_ enabled: Bool)

@_silgen_name("ime_modern")
func ime_modern(_ modern: Bool)

@_silgen_name("ime_clear")
func ime_clear()

@_silgen_name("ime_free")
func ime_free(_ result: UnsafeMutablePointer<ImeResult>?)

// MARK: - RustBridge

class RustBridge {
    static var isInitialized = false

    /// Initialize engine (call once at app start)
    static func initialize() {
        guard !isInitialized else { return }
        ime_init()
        isInitialized = true
        print("[RustBridge] Engine initialized")
    }

    /// Process key event
    /// Returns: (backspaceCount, newChars) or nil if no action needed
    static func processKey(keyCode: UInt16, caps: Bool, ctrl: Bool) -> (Int, [Character])? {
        guard isInitialized else {
            print("[RustBridge] Engine not initialized!")
            return nil
        }

        guard let resultPtr = ime_key(keyCode, caps, ctrl) else {
            return nil
        }

        let result = resultPtr.pointee
        ime_free(resultPtr)

        // Action: 0=None, 1=Send, 2=Restore
        guard result.action == 1 else {
            return nil
        }

        let backspace = Int(result.backspace)
        var chars: [Character] = []

        // Extract chars from tuple
        let charArray = withUnsafePointer(to: result.chars) { ptr in
            ptr.withMemoryRebound(to: UInt32.self, capacity: 32) { bound in
                Array(UnsafeBufferPointer(start: bound, count: Int(result.count)))
            }
        }

        for code in charArray {
            if let scalar = Unicode.Scalar(code) {
                chars.append(Character(scalar))
            }
        }

        return (backspace, chars)
    }

    /// Set input method (0=Telex, 1=VNI)
    static func setMethod(_ method: Int) {
        ime_method(UInt8(method))
        print("[RustBridge] Method set to: \(method == 0 ? "Telex" : "VNI")")
    }

    /// Enable/disable engine
    static func setEnabled(_ enabled: Bool) {
        ime_enabled(enabled)
        print("[RustBridge] Engine enabled: \(enabled)")
    }

    /// Set modern orthography (true=oà, false=òa)
    static func setModern(_ modern: Bool) {
        ime_modern(modern)
    }

    /// Clear buffer (new session, e.g., on mouse click)
    static func clearBuffer() {
        ime_clear()
    }
}

// MARK: - Keyboard Hook Manager

class KeyboardHookManager {
    static let shared = KeyboardHookManager()

    private var eventTap: CFMachPort?
    private var runLoopSource: CFRunLoopSource?
    private var isRunning = false

    private init() {}

    func start() {
        guard !isRunning else { return }

        // Initialize Rust engine
        RustBridge.initialize()

        // Create event tap
        let eventMask = (1 << CGEventType.keyDown.rawValue)

        guard let tap = CGEvent.tapCreate(
            tap: .cgSessionEventTap,
            place: .headInsertEventTap,
            options: .defaultTap,
            eventsOfInterest: CGEventMask(eventMask),
            callback: keyboardCallback,
            userInfo: nil
        ) else {
            print("[KeyboardHook] Failed to create event tap. Check Accessibility permissions.")
            return
        }

        eventTap = tap
        runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, tap, 0)

        if let source = runLoopSource {
            CFRunLoopAddSource(CFRunLoopGetCurrent(), source, .commonModes)
            CGEvent.tapEnable(tap: tap, enable: true)
            isRunning = true
            print("[KeyboardHook] Started")
        }
    }

    func stop() {
        guard isRunning else { return }

        if let tap = eventTap {
            CGEvent.tapEnable(tap: tap, enable: false)
        }

        if let source = runLoopSource {
            CFRunLoopRemoveSource(CFRunLoopGetCurrent(), source, .commonModes)
        }

        eventTap = nil
        runLoopSource = nil
        isRunning = false
        print("[KeyboardHook] Stopped")
    }
}

// MARK: - Keyboard Callback

private func keyboardCallback(
    proxy: CGEventTapProxy,
    type: CGEventType,
    event: CGEvent,
    refcon: UnsafeMutableRawPointer?
) -> Unmanaged<CGEvent>? {

    // Only handle key down
    guard type == .keyDown else {
        return Unmanaged.passRetained(event)
    }

    let keyCode = UInt16(event.getIntegerValueField(.keyboardEventKeycode))
    let flags = event.flags

    let caps = flags.contains(.maskShift) || flags.contains(.maskAlphaShift)
    let ctrl = flags.contains(.maskCommand) || flags.contains(.maskControl) ||
               flags.contains(.maskAlternate)

    // Process key through Rust engine
    if let (backspace, chars) = RustBridge.processKey(keyCode: keyCode, caps: caps, ctrl: ctrl) {
        print("[KeyboardHook] Processing: backspace=\(backspace), chars=\(chars)")

        // Send backspaces
        for _ in 0..<backspace {
            sendBackspace(proxy: proxy)
        }

        // Send new characters
        sendCharacters(chars, proxy: proxy)

        // Consume original event
        return nil
    }

    // Pass through
    return Unmanaged.passUnretained(event)
}

// MARK: - Send Keys

private func sendBackspace(proxy: CGEventTapProxy) {
    let source = CGEventSource(stateID: .privateState)

    if let down = CGEvent(keyboardEventSource: source, virtualKey: 0x33, keyDown: true),
       let up = CGEvent(keyboardEventSource: source, virtualKey: 0x33, keyDown: false) {
        down.post(tap: .cgSessionEventTap)
        up.post(tap: .cgSessionEventTap)
    }
}

private func sendCharacters(_ chars: [Character], proxy: CGEventTapProxy) {
    let source = CGEventSource(stateID: .privateState)

    let string = String(chars)
    let utf16 = Array(string.utf16)

    if let down = CGEvent(keyboardEventSource: source, virtualKey: 0, keyDown: true),
       let up = CGEvent(keyboardEventSource: source, virtualKey: 0, keyDown: false) {

        down.keyboardSetUnicodeString(stringLength: utf16.count, unicodeString: utf16)
        up.keyboardSetUnicodeString(stringLength: utf16.count, unicodeString: utf16)

        down.post(tap: .cgSessionEventTap)
        up.post(tap: .cgSessionEventTap)
    }
}
