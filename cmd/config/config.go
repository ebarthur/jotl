package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LogLevel represents the severity level of log messages
type LogLevel string

// LogFormat represents the output format for log messages
type LogFormat string

const (
	Debug LogLevel = "debug" // Detailed debug information
	Info  LogLevel = "info"  // General operational information
	Warn  LogLevel = "warn"  // Warning messages for potentially harmful situations
	Error LogLevel = "error" // Error messages for serious problems

	// Log Formats define how log messages are structured
	Text LogFormat = "text" // Human-readable text format

	// Default configuration values
	DefaultVersion     = "1.0.0"   // Initial version number
	DefaultTimeFormat  = "RFC3339" // Standard time format
	DefaultRefreshRate = 5         // Dashboard refresh rate in seconds
)

// Project contains basic project identification and description
type Project struct {
	Name        string `yaml:"name" json:"name"`               // Project name
	Description string `yaml:"description" json:"description"` // Project description
}

// Database contains database connection configuration
type Database struct {
	Path string `yaml:"path" json:"path"` // Database connection string or file path
}

// Logging contains log handling configuration
type Logging struct {
	Level      LogLevel  `yaml:"level" json:"level"`           // Minimum log level to record
	Format     LogFormat `yaml:"format" json:"format"`         // Output format for logs
	TimeFormat string    `yaml:"timeFormat" json:"timeFormat"` // Time format string for log entries
}

// Dashboard contains web interface configuration
type Dashboard struct {
	Port        int    `yaml:"port" json:"port"`               // HTTP port for dashboard
	Theme       string `yaml:"theme" json:"theme"`             // UI theme (system/light/dark)
	RefreshRate int    `yaml:"refreshRate" json:"refreshRate"` // Data refresh interval in seconds
}

// JotlConfig is the root configuration structure containing all settings
type JotlConfig struct {
	Version   string    `yaml:"version" json:"version"`     // Configuration version
	Project   Project   `yaml:"project" json:"project"`     // Project settings
	Database  Database  `yaml:"database" json:"database"`   // Database settings
	Logging   Logging   `yaml:"logging" json:"logging"`     // Logging settings
	Dashboard Dashboard `yaml:"dashboard" json:"dashboard"` // Dashboard settings
}

// NewConfig creates a new configuration with default values.
func NewConfig(name, loglevel, dbPath string) *JotlConfig {
	return &JotlConfig{
		Version: DefaultVersion,
		Project: Project{
			Name: name,
		},
		Database: Database{
			Path: dbPath,
		},
		Logging: Logging{
			Level:      LogLevel(loglevel),
			Format:     Text,
			TimeFormat: DefaultTimeFormat,
		},
		Dashboard: Dashboard{
			Port:        8080,
			Theme:       "system",
			RefreshRate: DefaultRefreshRate,
		},
	}
}

// SaveConfig saves the configuration to a YAML file.
// It creates parent directories if they don't exist.

func (c *JotlConfig) SaveConfig(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfig reads and parses a YAML configuration file.
// This may be called every time you run `... dev` to verify user configs
func LoadConfig(path string) (*JotlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &JotlConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// SetLogLevel updates the logging level if valid.
func (c *JotlConfig) SetLogLevel(level string) error {
	switch LogLevel(level) {
	case Debug, Info, Warn, Error:
		c.Logging.Level = LogLevel(level)
		return nil
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
}

// SetProjectName updates the project name.
func (c *JotlConfig) SetProjectName(name string) {
	c.Project.Name = name
}

// SetDescription updates the project description.
func (c *JotlConfig) SetDescription(desc string) {
	c.Project.Description = desc
}
