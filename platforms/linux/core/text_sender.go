package core

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
)

// TextSender handles Unicode text injection on Linux
// Uses xdotool as fallback for reliable Unicode support
type TextSender struct {
	useXdotool bool
}

// NewTextSender creates a text sender
func NewTextSender() *TextSender {
	// Check if xdotool is available
	_, err := exec.LookPath("xdotool")
	useXdotool := err == nil

	if useXdotool {
		log.Println("Using xdotool for text injection")
	} else {
		log.Println("xdotool not found, using XTest (limited Unicode)")
	}

	return &TextSender{useXdotool: useXdotool}
}

// SendText sends replacement text with backspaces
func (s *TextSender) SendText(text string, backspaces int) error {
	if s.useXdotool {
		return s.sendWithXdotool(text, backspaces)
	}
	return s.sendWithXTest(text, backspaces)
}

func (s *TextSender) sendWithXdotool(text string, backspaces int) error {
	// Send backspaces
	if backspaces > 0 {
		bsKeys := strings.Repeat("BackSpace ", backspaces)
		cmd := exec.Command("xdotool", "key", "--clearmodifiers", strings.TrimSpace(bsKeys))
		if err := cmd.Run(); err != nil {
			log.Printf("xdotool backspace error: %v", err)
		}
	}

	// Send text using type (handles Unicode)
	if text != "" {
		cmd := exec.Command("xdotool", "type", "--clearmodifiers", "--", text)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (s *TextSender) sendWithXTest(text string, backspaces int) error {
	// Fallback: use XTest for basic keys only
	// This won't work well for Vietnamese characters
	log.Printf("XTest injection not fully implemented for: %s", text)
	return nil
}

// SendBackspaces sends only backspace keys
func (s *TextSender) SendBackspaces(count int) error {
	if count <= 0 {
		return nil
	}

	if s.useXdotool {
		bsKeys := strings.Repeat("BackSpace ", count)
		cmd := exec.Command("xdotool", "key", "--clearmodifiers", strings.TrimSpace(bsKeys))
		return cmd.Run()
	}

	return nil
}

// SendUnicode sends a single Unicode codepoint
func (s *TextSender) SendUnicode(codepoint rune) error {
	if s.useXdotool {
		// xdotool type handles Unicode
		cmd := exec.Command("xdotool", "type", "--clearmodifiers", "--", string(codepoint))
		return cmd.Run()
	}

	// Fallback: Try Ctrl+Shift+U + hex code (works in GTK apps)
	hex := strconv.FormatInt(int64(codepoint), 16)
	log.Printf("Trying Ctrl+Shift+U method for: U+%s", hex)

	return nil
}
