package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds FKey Linux settings
type Config struct {
	Enabled     bool   `toml:"enabled"`
	InputMethod int    `toml:"input_method"` // 0=Telex, 1=VNI
	ModernTone  bool   `toml:"modern_tone"`
	EscRestore  bool   `toml:"esc_restore"`
	AutoStart   bool   `toml:"auto_start"`

	// Hotkey settings
	ToggleHotkey string `toml:"toggle_hotkey"` // e.g., "Ctrl+Space"
}

// Default returns default configuration
func Default() *Config {
	return &Config{
		Enabled:      true,
		InputMethod:  0, // Telex
		ModernTone:   true,
		EscRestore:   true,
		AutoStart:    false,
		ToggleHotkey: "Ctrl+Space",
	}
}

// ConfigPath returns XDG-compliant config file path
func ConfigPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "fkey", "config.toml")
}

// Load reads config from file
func Load() (*Config, error) {
	path := ConfigPath()

	cfg := Default()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Config doesn't exist, create default
		if err := Save(cfg); err != nil {
			return cfg, err
		}
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes config to file
func Save(cfg *Config) error {
	path := ConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(cfg)
}
