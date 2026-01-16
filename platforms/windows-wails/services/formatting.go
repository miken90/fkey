package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// FormatProfile defines formatting templates for different styles
type FormatProfile struct {
	Bold          string `json:"bold"`
	Italic        string `json:"italic"`
	Underline     string `json:"underline"`
	Code          string `json:"code"`
	Strikethrough string `json:"strikethrough"`
	Link          string `json:"link"`
}

// AppConfig holds per-app formatting configuration
type AppConfig struct {
	Profile        string            `json:"profile"`                  // "markdown", "html", or "disabled"
	ExcludeHotkeys []string          `json:"excludeHotkeys,omitempty"` // ["strikethrough", "link"] etc.
	CustomHotkeys  map[string]string `json:"customHotkeys,omitempty"`  // {"strikethrough": "Ctrl+Alt+S"} override default
}

// FormattingConfig holds all formatting configuration
type FormattingConfig struct {
	Enabled        bool                     `json:"enabled"`
	DefaultProfile string                   `json:"defaultProfile"` // "disabled", "markdown", "html"
	Hotkeys        map[string]string        `json:"hotkeys"`        // "bold" -> "Ctrl+B"
	Profiles       map[string]FormatProfile `json:"profiles"`
	Apps           map[string]AppConfig     `json:"apps"` // "discord" -> AppConfig{...}
}

// FormattingService manages formatting configuration
type FormattingService struct {
	config     *FormattingConfig
	configPath string
}

// NewFormattingService creates a new formatting service
func NewFormattingService() *FormattingService {
	return &FormattingService{
		config:     DefaultFormattingConfig(),
		configPath: getConfigPath(),
	}
}

// DefaultFormattingConfig returns the default configuration
func DefaultFormattingConfig() *FormattingConfig {
	return &FormattingConfig{
		Enabled:        true,
		DefaultProfile: "disabled",
		Hotkeys: map[string]string{
			"bold":          "Ctrl+B",
			"italic":        "Ctrl+I",
			"underline":     "Ctrl+U",
			"code":          "Ctrl+`",
			"strikethrough": "Ctrl+Alt+S",
			"link":          "Ctrl+K",
		},
		Profiles: map[string]FormatProfile{
			"markdown": {
				Bold:          "**{text}**",
				Italic:        "*{text}*",
				Underline:     "__{text}__",
				Code:          "`{text}`",
				Strikethrough: "~~{text}~~",
				Link:          "[{text}](url)",
			},
			"html": {
				Bold:          "<b>{text}</b>",
				Italic:        "<i>{text}</i>",
				Underline:     "<u>{text}</u>",
				Code:          "<code>{text}</code>",
				Strikethrough: "<s>{text}</s>",
				Link:          "<a href=\"url\">{text}</a>",
			},
		},
		Apps: map[string]AppConfig{
			"discord":       {Profile: "markdown"},
			"discordcanary": {Profile: "markdown"},
			"discordptb":    {Profile: "markdown"},
			"slack":         {Profile: "markdown"},
			"telegram":      {Profile: "markdown"},
			"notion":        {Profile: "markdown"},
			"obsidian":      {Profile: "markdown"},
			"chrome":        {Profile: "markdown"},
			"firefox":       {Profile: "markdown"},
			"msedge":        {Profile: "markdown"},
		},
	}
}

// getConfigPath returns the path to formatting.json
func getConfigPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return "formatting.json"
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "settings", "formatting.json")
}

// Config returns the current configuration
func (s *FormattingService) Config() *FormattingConfig {
	return s.config
}

// Load reads configuration from file, falls back to defaults if missing
func (s *FormattingService) Load() error {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		// File doesn't exist, use defaults
		s.config = DefaultFormattingConfig()
		return nil
	}

	// First, try to unmarshal into a raw map to detect format
	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		// Invalid JSON, use defaults
		s.config = DefaultFormattingConfig()
		return nil
	}

	config := &FormattingConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		// Invalid JSON, use defaults
		s.config = DefaultFormattingConfig()
		return nil
	}

	// Handle backwards compatibility for apps field
	// Old format: {"apps": {"chrome": "markdown"}}
	// New format: {"apps": {"chrome": {"profile": "markdown", "excludeHotkeys": [], "customHotkeys": {}}}}
	if rawApps, ok := rawConfig["apps"].(map[string]interface{}); ok {
		config.Apps = make(map[string]AppConfig)
		for k, v := range rawApps {
			switch appVal := v.(type) {
			case string:
				// Old format: string profile
				config.Apps[k] = AppConfig{Profile: appVal}
			case map[string]interface{}:
				// New format: object with profile, excludeHotkeys, customHotkeys
				appConfig := AppConfig{}
				if profile, ok := appVal["profile"].(string); ok {
					appConfig.Profile = profile
				}
				if excludeHotkeys, ok := appVal["excludeHotkeys"].([]interface{}); ok {
					for _, h := range excludeHotkeys {
						if hStr, ok := h.(string); ok {
							appConfig.ExcludeHotkeys = append(appConfig.ExcludeHotkeys, hStr)
						}
					}
				}
				if customHotkeys, ok := appVal["customHotkeys"].(map[string]interface{}); ok {
					appConfig.CustomHotkeys = make(map[string]string)
					for hk, hv := range customHotkeys {
						if hvStr, ok := hv.(string); ok {
							appConfig.CustomHotkeys[hk] = hvStr
						}
					}
				}
				config.Apps[k] = appConfig
			}
		}
	}

	// Merge with defaults to ensure all fields exist
	defaults := DefaultFormattingConfig()
	if config.Hotkeys == nil {
		config.Hotkeys = defaults.Hotkeys
	}
	if config.Profiles == nil {
		config.Profiles = defaults.Profiles
	}
	if config.Apps == nil {
		config.Apps = defaults.Apps
	}

	s.config = config
	return nil
}

// Save writes configuration to file
func (s *FormattingService) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.configPath, data, 0644)
}

// GetProfileForApp returns the profile name for a given process name
func (s *FormattingService) GetProfileForApp(processName string) string {
	if !s.config.Enabled {
		return "disabled"
	}

	// Normalize process name (lowercase, remove .exe)
	name := strings.ToLower(processName)
	name = strings.TrimSuffix(name, ".exe")

	if appConfig, ok := s.config.Apps[name]; ok {
		return appConfig.Profile
	}

	return s.config.DefaultProfile
}

// IsHotkeyExcluded checks if a hotkey is excluded for a given app
func (s *FormattingService) IsHotkeyExcluded(processName, formatType string) bool {
	// Normalize process name (lowercase, remove .exe)
	name := strings.ToLower(processName)
	name = strings.TrimSuffix(name, ".exe")

	if appConfig, ok := s.config.Apps[name]; ok {
		for _, excluded := range appConfig.ExcludeHotkeys {
			if excluded == formatType {
				return true
			}
		}
	}
	return false
}

// GetCustomHotkey returns the custom hotkey for a format type in a specific app
// Returns empty string if no custom hotkey is set (use default)
func (s *FormattingService) GetCustomHotkey(processName, formatType string) string {
	name := strings.ToLower(processName)
	name = strings.TrimSuffix(name, ".exe")

	if appConfig, ok := s.config.Apps[name]; ok {
		if appConfig.CustomHotkeys != nil {
			if hotkey, ok := appConfig.CustomHotkeys[formatType]; ok {
				return hotkey
			}
		}
	}
	return ""
}

// GetAppConfig returns the full config for an app
func (s *FormattingService) GetAppConfig(processName string) *AppConfig {
	name := strings.ToLower(processName)
	name = strings.TrimSuffix(name, ".exe")

	if appConfig, ok := s.config.Apps[name]; ok {
		return &appConfig
	}
	return nil
}

// SetAppCustomHotkey sets a custom hotkey for an app
func (s *FormattingService) SetAppCustomHotkey(appName, formatType, hotkey string) {
	name := strings.ToLower(appName)
	if existing, ok := s.config.Apps[name]; ok {
		if existing.CustomHotkeys == nil {
			existing.CustomHotkeys = make(map[string]string)
		}
		if hotkey == "" {
			delete(existing.CustomHotkeys, formatType)
		} else {
			existing.CustomHotkeys[formatType] = hotkey
		}
		s.config.Apps[name] = existing
	}
}

// Format applies formatting template to text
func (s *FormattingService) Format(formatType, text, profileName string) string {
	profile, ok := s.config.Profiles[profileName]
	if !ok {
		return text
	}

	var template string
	switch formatType {
	case "bold":
		template = profile.Bold
	case "italic":
		template = profile.Italic
	case "underline":
		template = profile.Underline
	case "code":
		template = profile.Code
	case "strikethrough":
		template = profile.Strikethrough
	case "link":
		template = profile.Link
	default:
		return text
	}

	if template == "" {
		return text
	}

	return strings.Replace(template, "{text}", text, 1)
}

// SetEnabled enables or disables formatting
func (s *FormattingService) SetEnabled(enabled bool) {
	s.config.Enabled = enabled
}

// IsEnabled returns whether formatting is enabled
func (s *FormattingService) IsEnabled() bool {
	return s.config.Enabled
}

// SetAppProfile sets the profile for a specific app (preserves ExcludeHotkeys)
func (s *FormattingService) SetAppProfile(appName, profile string) {
	name := strings.ToLower(appName)
	if existing, ok := s.config.Apps[name]; ok {
		existing.Profile = profile
		s.config.Apps[name] = existing
	} else {
		s.config.Apps[name] = AppConfig{Profile: profile, ExcludeHotkeys: []string{}}
	}
}

// RemoveAppProfile removes an app from the configuration
func (s *FormattingService) RemoveAppProfile(appName string) {
	delete(s.config.Apps, strings.ToLower(appName))
}

// ToMap converts the configuration to a map for frontend
func (s *FormattingService) ToMap() map[string]interface{} {
	appsMap := make(map[string]interface{})
	for k, v := range s.config.Apps {
		appMap := map[string]interface{}{
			"profile": v.Profile,
		}
		if len(v.ExcludeHotkeys) > 0 {
			appMap["excludeHotkeys"] = v.ExcludeHotkeys
		}
		if len(v.CustomHotkeys) > 0 {
			appMap["customHotkeys"] = v.CustomHotkeys
		}
		appsMap[k] = appMap
	}
	return map[string]interface{}{
		"enabled":        s.config.Enabled,
		"defaultProfile": s.config.DefaultProfile,
		"hotkeys":        s.config.Hotkeys,
		"apps":           appsMap,
	}
}

// FromMap updates the configuration from a map
func (s *FormattingService) FromMap(config map[string]interface{}) {
	if enabled, ok := config["enabled"].(bool); ok {
		s.config.Enabled = enabled
	}
	if defaultProfile, ok := config["defaultProfile"].(string); ok {
		s.config.DefaultProfile = defaultProfile
	}
	if hotkeys, ok := config["hotkeys"].(map[string]interface{}); ok {
		s.config.Hotkeys = make(map[string]string)
		for k, v := range hotkeys {
			if vStr, ok := v.(string); ok {
				s.config.Hotkeys[k] = vStr
			}
		}
	}
	if apps, ok := config["apps"].(map[string]interface{}); ok {
		s.config.Apps = make(map[string]AppConfig)
		for k, v := range apps {
			switch appVal := v.(type) {
			case string:
				// Old format: string profile
				s.config.Apps[k] = AppConfig{Profile: appVal}
			case map[string]interface{}:
				// New format: object with profile, excludeHotkeys, customHotkeys
				appConfig := AppConfig{}
				if profile, ok := appVal["profile"].(string); ok {
					appConfig.Profile = profile
				}
				if excludeHotkeys, ok := appVal["excludeHotkeys"].([]interface{}); ok {
					for _, h := range excludeHotkeys {
						if hStr, ok := h.(string); ok {
							appConfig.ExcludeHotkeys = append(appConfig.ExcludeHotkeys, hStr)
						}
					}
				}
				if customHotkeys, ok := appVal["customHotkeys"].(map[string]interface{}); ok {
					appConfig.CustomHotkeys = make(map[string]string)
					for hk, hv := range customHotkeys {
						if hvStr, ok := hv.(string); ok {
							appConfig.CustomHotkeys[hk] = hvStr
						}
					}
				}
				s.config.Apps[k] = appConfig
			}
		}
	}
}

// GetGlobalHotkey returns the global hotkey for a format type
// Returns empty string if using default
func (s *FormattingService) GetGlobalHotkey(formatType string) string {
	if s.config.Hotkeys != nil {
		if hotkey, ok := s.config.Hotkeys[formatType]; ok {
			return hotkey
		}
	}
	return ""
}

// ParseHotkeyString parses "Ctrl+Alt+S" format into keyCode and modifiers
func ParseHotkeyString(hotkeyStr string) (keyCode uint16, ctrl, alt, shift bool) {
	parts := strings.Split(hotkeyStr, "+")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch strings.ToLower(part) {
		case "ctrl":
			ctrl = true
		case "alt":
			alt = true
		case "shift":
			shift = true
		default:
			// This is the key
			keyCode = keyNameToVK(part)
		}
	}
	return
}

// keyNameToVK converts key name to virtual key code
func keyNameToVK(name string) uint16 {
	name = strings.ToUpper(strings.TrimSpace(name))
	
	// Single letter A-Z
	if len(name) == 1 && name[0] >= 'A' && name[0] <= 'Z' {
		return uint16(name[0])
	}
	
	// Single digit 0-9
	if len(name) == 1 && name[0] >= '0' && name[0] <= '9' {
		return uint16(name[0])
	}
	
	// Special keys
	switch name {
	case "`", "~":
		return 0xC0 // VK_OEM_3
	case "SPACE":
		return 0x20
	case "ENTER", "RETURN":
		return 0x0D
	case "TAB":
		return 0x09
	case "ESCAPE", "ESC":
		return 0x1B
	case "BACKSPACE":
		return 0x08
	case "DELETE", "DEL":
		return 0x2E
	case "INSERT", "INS":
		return 0x2D
	case "HOME":
		return 0x24
	case "END":
		return 0x23
	case "PAGEUP", "PGUP":
		return 0x21
	case "PAGEDOWN", "PGDN":
		return 0x22
	case "UP":
		return 0x26
	case "DOWN":
		return 0x28
	case "LEFT":
		return 0x25
	case "RIGHT":
		return 0x27
	case "F1":
		return 0x70
	case "F2":
		return 0x71
	case "F3":
		return 0x72
	case "F4":
		return 0x73
	case "F5":
		return 0x74
	case "F6":
		return 0x75
	case "F7":
		return 0x76
	case "F8":
		return 0x77
	case "F9":
		return 0x78
	case "F10":
		return 0x79
	case "F11":
		return 0x7A
	case "F12":
		return 0x7B
	}
	
	return 0
}
