use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;

#[repr(C)]
#[derive(Debug, Serialize, Deserialize)]
pub struct Config {
    pub enabled: bool,
    pub mode: u8, // 0 = Telex, 1 = VNI
}

impl Default for Config {
    fn default() -> Self {
        Self {
            enabled: true,
            mode: 0, // Telex by default
        }
    }
}

impl Config {
    /// Load configuration from file
    pub fn load() -> Self {
        let path = Self::config_path();

        if let Ok(content) = fs::read_to_string(&path) {
            if let Ok(config) = toml::from_str(&content) {
                return config;
            }
        }

        Self::default()
    }

    /// Save configuration to file
    pub fn save(&self) -> Result<(), Box<dyn std::error::Error>> {
        let path = Self::config_path();

        // Ensure config directory exists
        if let Some(parent) = path.parent() {
            fs::create_dir_all(parent)?;
        }

        let content = toml::to_string_pretty(self)?;
        fs::write(path, content)?;
        Ok(())
    }

    /// Get configuration file path
    fn config_path() -> PathBuf {
        #[cfg(target_os = "macos")]
        {
            let mut path = dirs::home_dir().unwrap_or_else(|| PathBuf::from("."));
            path.push("Library");
            path.push("Application Support");
            path.push("GoNhanh");
            path.push("config.toml");
            path
        }

        #[cfg(target_os = "windows")]
        {
            let mut path = dirs::config_dir().unwrap_or_else(|| PathBuf::from("."));
            path.push("GoNhanh");
            path.push("config.toml");
            path
        }

        #[cfg(not(any(target_os = "macos", target_os = "windows")))]
        {
            let mut path = dirs::config_dir().unwrap_or_else(|| PathBuf::from("."));
            path.push("gonhanh");
            path.push("config.toml");
            path
        }
    }
}

// Add dirs dependency
#[allow(dead_code)]
mod dirs {
    use std::path::PathBuf;

    pub fn home_dir() -> Option<PathBuf> {
        std::env::var_os("HOME").map(PathBuf::from)
    }

    pub fn config_dir() -> Option<PathBuf> {
        #[cfg(target_os = "macos")]
        {
            home_dir().map(|h| h.join("Library/Application Support"))
        }

        #[cfg(not(target_os = "macos"))]
        {
            std::env::var_os("XDG_CONFIG_HOME")
                .map(PathBuf::from)
                .or_else(|| home_dir().map(|h| h.join(".config")))
        }
    }
}
