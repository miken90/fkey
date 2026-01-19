package core

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// InjectionMethod determines how text is injected
type InjectionMethod int

const (
	MethodUInput  InjectionMethod = iota // Direct uinput (fastest, no process spawn)
	MethodXdotool                        // xdotool fallback (reliable Unicode)
)

// TextSender handles Unicode text injection on Linux
// Uses uinput for speed, falls back to xdotool for reliability
type TextSender struct {
	method       InjectionMethod
	uinputDevice *UInputDevice
	useXdotool   bool
}

// NewTextSender creates a text sender with best available method
func NewTextSender() *TextSender {
	ts := &TextSender{}

	// Try uinput first (fastest)
	if IsUInputAvailable() {
		device, err := GetUInputDevice()
		if err == nil {
			ts.uinputDevice = device
			ts.method = MethodUInput
			log.Println("Using uinput for text injection (fastest)")
			return ts
		}
		log.Printf("uinput init failed: %v, falling back to xdotool", err)
	}

	// Fall back to xdotool
	_, err := exec.LookPath("xdotool")
	if err == nil {
		ts.useXdotool = true
		ts.method = MethodXdotool
		log.Println("Using xdotool for text injection (fallback)")
		return ts
	}

	log.Println("WARNING: No text injection method available (need uinput access or xdotool)")
	return ts
}

// SendText sends replacement text with backspaces
func (s *TextSender) SendText(text string, backspaces int) error {
	switch s.method {
	case MethodUInput:
		return s.sendWithUInput(text, backspaces)
	case MethodXdotool:
		return s.sendWithXdotool(text, backspaces)
	default:
		log.Printf("No injection method available for: %s", text)
		return nil
	}
}

func (s *TextSender) sendWithUInput(text string, backspaces int) error {
	log.Printf("[DEBUG TextSender] uinput: backspaces=%d, text=%q", backspaces, text)

	if s.uinputDevice == nil {
		// Try to reinitialize
		device, err := GetUInputDevice()
		if err != nil {
			log.Printf("uinput unavailable, falling back to xdotool: %v", err)
			return s.sendWithXdotool(text, backspaces)
		}
		s.uinputDevice = device
	}

	err := s.uinputDevice.SendText(text, backspaces)
	if err != nil {
		log.Printf("uinput failed, falling back to xdotool: %v", err)
		return s.sendWithXdotool(text, backspaces)
	}

	return nil
}

func (s *TextSender) sendWithXdotool(text string, backspaces int) error {
	log.Printf("[DEBUG TextSender] xdotool: backspaces=%d, text=%q", backspaces, text)

	// Small delay to ensure original key has been processed
	time.Sleep(10 * time.Millisecond)

	// Send backspaces - use multiple key arguments in one command
	if backspaces > 0 {
		args := []string{"key"}
		for i := 0; i < backspaces; i++ {
			args = append(args, "BackSpace")
		}
		cmd := exec.Command("xdotool", args...)
		if err := cmd.Run(); err != nil {
			log.Printf("xdotool backspace error: %v", err)
		}
	}

	// Send text using type (handles Unicode)
	if text != "" {
		log.Printf("[DEBUG TextSender] Typing text: %q", text)
		cmd := exec.Command("xdotool", "type", "--delay", "0", "--", text)
		if err := cmd.Run(); err != nil {
			log.Printf("[DEBUG TextSender] xdotool type error: %v", err)
			return err
		}
		log.Printf("[DEBUG TextSender] Text sent successfully")
	}

	return nil
}

// SendBackspaces sends only backspace keys
func (s *TextSender) SendBackspaces(count int) error {
	if count <= 0 {
		return nil
	}

	switch s.method {
	case MethodUInput:
		if s.uinputDevice != nil {
			return s.uinputDevice.SendBackspaces(count)
		}
		fallthrough
	case MethodXdotool:
		if s.useXdotool {
			bsKeys := strings.Repeat("BackSpace ", count)
			cmd := exec.Command("xdotool", "key", "--clearmodifiers", strings.TrimSpace(bsKeys))
			return cmd.Run()
		}
	}

	return nil
}

// SendUnicode sends a single Unicode codepoint
func (s *TextSender) SendUnicode(codepoint rune) error {
	return s.SendText(string(codepoint), 0)
}

// GetMethod returns the current injection method
func (s *TextSender) GetMethod() InjectionMethod {
	return s.method
}

// GetMethodName returns a human-readable name for the current method
func (s *TextSender) GetMethodName() string {
	switch s.method {
	case MethodUInput:
		return "uinput"
	case MethodXdotool:
		return "xdotool"
	default:
		return "none"
	}
}

// ForceMethod forces a specific injection method (for testing)
func (s *TextSender) ForceMethod(method InjectionMethod) {
	s.method = method
	log.Printf("Forced injection method to: %s", s.GetMethodName())
}

// Close cleans up resources
func (s *TextSender) Close() {
	// uinput device is managed globally, don't close here
}

// sendWithXTest is a placeholder for XTest extension (not implemented)
func (s *TextSender) sendWithXTest(text string, backspaces int) error {
	log.Printf("XTest injection not fully implemented for: %s", text)
	return nil
}

// unused but kept for API compatibility
func (s *TextSender) sendUnicodeLegacy(codepoint rune) error {
	if s.useXdotool {
		cmd := exec.Command("xdotool", "type", "--clearmodifiers", "--", string(codepoint))
		return cmd.Run()
	}

	// Fallback: Try Ctrl+Shift+U + hex code (works in GTK apps)
	hex := strconv.FormatInt(int64(codepoint), 16)
	log.Printf("Trying Ctrl+Shift+U method for: U+%s", hex)
	return nil
}
