package core

import "testing"

func TestSmartPaste_IsMojibake(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "mojibake Vietnamese text CP850",
			input:    "\u00adƒÄ╣ FKey - Bß╗Ö g├Á tiß║┐ng Viß╗çt",
			expected: true,
		},
		{
			name:     "clean English text",
			input:    "Hello World",
			expected: false,
		},
		{
			name:     "clean Vietnamese text",
			input:    "🎹 FKey - Bộ gõ tiếng Việt",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "partial mojibake pattern",
			input:    "Some text with ß╗ in it",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMojibake(tt.input)
			if result != tt.expected {
				t.Errorf("IsMojibake(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSmartPaste_FixMojibake(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedOutput  string
		expectedChanged bool
	}{
		{
			name:            "fix Vietnamese mojibake CP850",
			input:           "\u00adƒÄ╣ FKey - Bß╗Ö g├Á tiß║┐ng Viß╗çt",
			expectedOutput:  "🎹 FKey - Bộ gõ tiếng Việt",
			expectedChanged: true,
		},
		{
			name:            "English text unchanged",
			input:           "Hello World! This is a test.",
			expectedOutput:  "Hello World! This is a test.",
			expectedChanged: false,
		},
		{
			name:            "already correct Vietnamese unchanged",
			input:           "🎹 FKey - Bộ gõ tiếng Việt",
			expectedOutput:  "🎹 FKey - Bộ gõ tiếng Việt",
			expectedChanged: false,
		},
		{
			name:            "empty string",
			input:           "",
			expectedOutput:  "",
			expectedChanged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, changed := FixMojibake(tt.input)
			if result != tt.expectedOutput {
				t.Errorf("FixMojibake(%q) = %q, want %q", tt.input, result, tt.expectedOutput)
			}
			if changed != tt.expectedChanged {
				t.Errorf("FixMojibake(%q) changed = %v, want %v", tt.input, changed, tt.expectedChanged)
			}
		})
	}
}

func TestSmartPaste_ContainsVietnamese(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Vietnamese with extended block chars",
			input:    "Bộ gõ tiếng Việt",
			expected: true,
		},
		{
			name:     "Vietnamese with basic diacritics",
			input:    "Đà Nẵng",
			expected: true,
		},
		{
			name:     "English only",
			input:    "Hello World",
			expected: false,
		},
		{
			name:     "numbers and symbols",
			input:    "12345!@#$%",
			expected: false,
		},
		{
			name:     "emoji only",
			input:    "🎹🎵🎶",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsVietnamese(tt.input)
			if result != tt.expected {
				t.Errorf("containsVietnamese(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
