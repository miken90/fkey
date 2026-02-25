# Research Report: Discord Formatting Lag/Freeze Issues

**Date:** 2026-01-18  
**Issue:** Lag, đơ, mất chữ khi dùng text formatting với Discord

---

## Executive Summary

Vấn đề lag/đơ/mất chữ khi dùng tính năng định dạng văn bản với Discord có 5 nguyên nhân chính:

1. **Clipboard race conditions** - OpenClipboard fail khi app khác đang giữ
2. **Fixed delays quá ngắn** - 50ms không đủ cho Electron apps
3. **Overlapping operations** - Goroutine restore chồng chéo
4. **Modifier key handling** - Không xử lý L/R variants
5. **Empty clipboard not restored** - Logic bỏ qua restore khi clipboard trống

---

## Root Cause Analysis

### 1. Clipboard Race Conditions

**Problem:** `OpenClipboard(0)` fails với error `0x800401D0` khi app khác đang giữ clipboard.

**Current code (clipboard.go:44-46):**
```go
ret, _, _ := procOpenClipboard.Call(0)
if ret == 0 {
    return "", ErrOpenClipboard  // Fails immediately!
}
```

**Solution:** Retry loop với backoff:
```go
func OpenClipboardRetry(maxRetries int, delay time.Duration) error {
    for i := 0; i < maxRetries; i++ {
        ret, _, _ := procOpenClipboard.Call(0)
        if ret != 0 {
            return nil
        }
        time.Sleep(delay)
    }
    return ErrOpenClipboard
}
```

### 2. Fixed Delays Không Đủ

**Problem:** 50ms delay sau Ctrl+C không đủ cho Discord (Electron app).

**Current code (format_handler.go:12):**
```go
ClipboardCopyDelay = 50  // Too short for Electron!
```

**Solution:** Dùng `GetClipboardSequenceNumber()` để poll thay vì fixed delay:
```go
// Windows API
var procGetClipboardSequenceNumber = user32.NewProc("GetClipboardSequenceNumber")

func WaitClipboardChange(oldSeq uint32, timeout time.Duration) bool {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        newSeq, _, _ := procGetClipboardSequenceNumber.Call()
        if uint32(newSeq) != oldSeq {
            return true
        }
        time.Sleep(10 * time.Millisecond)
    }
    return false
}
```

### 3. Overlapping Operations

**Problem:** Goroutine restore clipboard có thể chồng chéo với operation tiếp theo.

**Current code (format_handler.go:283-288):**
```go
go func() {
    time.Sleep(ClipboardRestoreWait * time.Millisecond)
    if originalClipboard != "" {
        SetClipboardText(originalClipboard)  // Race with next operation!
    }
}()
```

**Solution:** Serialize all operations với mutex:
```go
var formatOpMu sync.Mutex

func (h *FormatHandler) HandleFormatHotkey(formatType, profile string) {
    formatOpMu.Lock()
    defer formatOpMu.Unlock()
    // ... all operations synchronously
}
```

### 4. Modifier Key Handling

**Problem:** Chỉ release VK_CONTROL, không xử lý VK_LCONTROL/VK_RCONTROL.

**Current code (format_handler.go:295-304):**
```go
if isKeyDown(VK_CONTROL) {
    // Only releases VK_CONTROL
}
```

**Solution:** Release cả left và right variants:
```go
const (
    VK_LCONTROL = 0xA2
    VK_RCONTROL = 0xA3
    VK_LSHIFT   = 0xA0
    VK_RSHIFT   = 0xA1
    VK_LMENU    = 0xA4
    VK_RMENU    = 0xA5
)

func ReleaseAllModifiers() {
    keysToRelease := []uint16{
        VK_CONTROL, VK_LCONTROL, VK_RCONTROL,
        VK_SHIFT, VK_LSHIFT, VK_RSHIFT,
        VK_MENU, VK_LMENU, VK_RMENU,
    }
    // Release all that are down
}
```

### 5. Empty Clipboard Not Restored

**Problem:** Nếu clipboard ban đầu trống, không restore.

**Current code (format_handler.go:285-287):**
```go
if originalClipboard != "" {  // Bug: never restores empty clipboard!
    SetClipboardText(originalClipboard)
}
```

**Solution:** 
```go
// Always restore, even if empty
if originalClipboard == "" {
    ClearClipboard()
} else {
    SetClipboardText(originalClipboard)
}
```

---

## Recommended Implementation Plan

### Phase 1: Critical Fixes (High Impact, Low Risk)

1. **Add clipboard retry loop** - `OpenClipboardRetry(5, 20ms)`
2. **Use clipboard sequence polling** - Replace fixed 50ms with sequence-based wait
3. **Add serialization mutex** - Prevent overlapping operations

### Phase 2: Robustness Improvements

4. **Fix empty clipboard restore** - Add `ClearClipboard()` function
5. **Handle L/R modifiers** - Release all variants
6. **Add compare-and-swap restore** - Only restore if clipboard unchanged

### Phase 3: Discord-Specific Tuning

7. **Per-app timing config** - Add to `FormattingConfig`:
   ```go
   type AppConfig struct {
       Profile          string
       CopyTimeout      int  // ms, default 600
       PasteSettleDelay int  // ms, default 100
   }
   ```

8. **Discord defaults:**
   - CopyTimeout: 800ms
   - PasteSettleDelay: 200ms

---

## Code Changes Summary

| File | Change | Impact |
|------|--------|--------|
| `clipboard.go` | Add retry loop, sequence polling | High |
| `format_handler.go` | Add mutex, fix restore logic | High |
| `format_handler.go` | Release L/R modifiers | Medium |
| `services/formatting.go` | Add per-app timing config | Medium |
| `settings/formatting.json` | Discord timing overrides | Low |

---

## Alternative Approaches (Not Recommended Yet)

1. **UI Automation API** - More reliable but limited hotkey support
2. **Direct Unicode injection** - Bypass clipboard but needs selection length
3. **TSF (Text Services Framework)** - Complex, app-specific

**Recommendation:** Implement Phase 1-2 first. Only explore alternatives if issues persist.

---

## Testing Strategy

1. Run formatting với Discord liên tục 20+ lần
2. Kiểm tra clipboard sau mỗi operation
3. Test với rapid consecutive hotkeys
4. Test với clipboard ban đầu trống
5. Test với text selection dài

---

## References

- [Windows Clipboard Retry Pattern](https://learn.microsoft.com/en-us/answers/questions/1695747)
- [SendInput Best Practices](https://stackoverflow.com/questions/14184493)
- [GetClipboardSequenceNumber](https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getclipboardsequencenumber)
