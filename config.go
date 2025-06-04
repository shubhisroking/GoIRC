package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	IRC      IRCConfig `json:"irc"`
	UI       UIConfig  `json:"ui"`
	Logging  LogConfig `json:"logging"`
	FilePath string    `json:"-"` // Don't serialize the file path
}

// IRCConfig contains IRC-related configuration
type IRCConfig struct {
	Server   string   `json:"server"`
	Port     int      `json:"port,omitempty"`
	Nick     string   `json:"nick"`
	Username string   `json:"username,omitempty"`
	RealName string   `json:"realname,omitempty"`
	Channels []string `json:"channels"`
	UseSSL   bool     `json:"use_ssl"`
	Password string   `json:"password,omitempty"`
	QuitMsg  string   `json:"quit_message,omitempty"`
}

// UIConfig contains UI-related configuration
type UIConfig struct {
	ShowSidebar  bool `json:"show_sidebar"`
	SidebarWidth int  `json:"sidebar_width"`
	Theme        struct {
		Primary   string `json:"primary"`
		Secondary string `json:"secondary"`
		Accent    string `json:"accent"`
	} `json:"theme"`
}

// LogConfig contains logging configuration
type LogConfig struct {
	Enabled   bool   `json:"enabled"`
	MaxSizeKB int    `json:"max_size_kb"`
	LogPath   string `json:"log_path"`
	DebugMode bool   `json:"debug_mode"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "goirc")

	return &Config{
		IRC: IRCConfig{
			Server:   "irc.libera.chat",
			Port:     6697,
			Nick:     "goirc-user",
			Username: "goirc",
			RealName: "GoIRC Client",
			Channels: []string{"#goirc-test"},
			UseSSL:   true,
			QuitMsg:  "Goodbye from GoIRC!",
		},
		UI: UIConfig{
			ShowSidebar:  true,
			SidebarWidth: 30,
			Theme: struct {
				Primary   string `json:"primary"`
				Secondary string `json:"secondary"`
				Accent    string `json:"accent"`
			}{
				Primary:   "#7C3AED",
				Secondary: "#A855F7",
				Accent:    "#EC4899",
			},
		},
		Logging: LogConfig{
			Enabled:   false, // Disabled by default to prevent log spam
			MaxSizeKB: 512,
			LogPath:   filepath.Join(configDir, "logs"),
			DebugMode: false,
		},
		FilePath: filepath.Join(configDir, "config.json"),
	}
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "goirc")
	return configDir, nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Also create logs directory
	logsDir := filepath.Join(configDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	return nil
}

// LoadConfig loads configuration from the config file or creates a default one
func LoadConfig() (*Config, error) {
	if err := EnsureConfigDir(); err != nil {
		return nil, err
	}

	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.json")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		config := DefaultConfig()
		config.FilePath = configPath
		if err := config.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return config, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	config.FilePath = configPath

	// Validate and fill missing fields with defaults
	defaultConfig := DefaultConfig()
	if config.IRC.Server == "" {
		config.IRC = defaultConfig.IRC
	}
	if config.Logging.MaxSizeKB == 0 {
		config.Logging.MaxSizeKB = defaultConfig.Logging.MaxSizeKB
	}
	if config.Logging.LogPath == "" {
		config.Logging.LogPath = defaultConfig.Logging.LogPath
	}

	return &config, nil
}

// Save saves the configuration to the config file
func (c *Config) Save() error {
	if c.FilePath == "" {
		configDir, err := GetConfigDir()
		if err != nil {
			return err
		}
		c.FilePath = filepath.Join(configDir, "config.json")
	}

	// Ensure the directory exists
	dir := filepath.Dir(c.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.FilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetLogFilePath returns the current log file path
func (c *Config) GetLogFilePath() string {
	if c.Logging.LogPath == "" {
		configDir, _ := GetConfigDir()
		c.Logging.LogPath = filepath.Join(configDir, "logs")
	}
	return filepath.Join(c.Logging.LogPath, "irc.log")
}

// GetDebugLogFilePath returns the debug log file path
func (c *Config) GetDebugLogFilePath() string {
	if c.Logging.LogPath == "" {
		configDir, _ := GetConfigDir()
		c.Logging.LogPath = filepath.Join(configDir, "logs")
	}
	return filepath.Join(c.Logging.LogPath, "debug.log")
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.IRC.Server == "" {
		return fmt.Errorf("IRC server cannot be empty")
	}
	if c.IRC.Nick == "" {
		return fmt.Errorf("IRC nick cannot be empty")
	}
	if c.Logging.MaxSizeKB <= 0 {
		return fmt.Errorf("log max size must be positive")
	}
	return nil
}
