package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator"
	"gopkg.in/yaml.v3"
)

type LogLevel string

type LogFormat string

const (
	Debug LogLevel = "debug" // Detailed debug information
	Info  LogLevel = "info"  // General operational information
	Warn  LogLevel = "warn"  // Warning messages for potentially harmful situations
	Error LogLevel = "error" // Error messages for serious problems

	Text LogFormat = "text"

	// Default configuration values
	DefaultVersion     = "1.0.0"   // Initial version number
	DefaultTimeFormat  = "RFC3339" // Standard time format (Not in use right now)
	DefaultRefreshRate = 5         // Dashboard refresh rate in seconds
)

type Project struct {
	Name        string `yaml:"name" validate:"required"`
	Description string `yaml:"description"`
}

type Database struct {
	Path string `yaml:"path" validate:"required"`
}

type Logging struct {
	Level      LogLevel  `yaml:"level" validate:"required,oneof=error info debug warn"`
	Format     LogFormat `yaml:"format" validate:"required"`
	TimeFormat string    `yaml:"timeFormat" validate:"required"`
}

type Dashboard struct {
	Port        string `yaml:"port" validate:"required"`
	Theme       string `yaml:"theme" validate:"required"`
	RefreshRate int    `yaml:"refreshRate" validate:"required"`
}

type JotlConfig struct {
	Version   string    `yaml:"version" json:"version"`
	Project   Project   `yaml:"project" json:"project"`
	Database  Database  `yaml:"database" json:"database"`
	Logging   Logging   `yaml:"logging" json:"logging"`
	Dashboard Dashboard `yaml:"dashboard" json:"dashboard"`
}

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
			Port:        "8080",
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

// LoadConfig reads and parses user YAML configuration file.
// This may be called every time you run `... dev` to verify user configs
func LoadConfig(startDir string) (*JotlConfig, error) {

	for {
		configFilePath := filepath.Join(startDir, "jotl", "config.yaml")

		if _, err := os.Stat(configFilePath); err == nil {

			data, err := os.ReadFile(configFilePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}

			config := &JotlConfig{}
			if err := yaml.Unmarshal(data, config); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}

			validate := validator.New()
			if err := validate.Struct(config); err != nil {
				return nil, fmt.Errorf("config validation failed: %w", err)
			}

			return config, nil
		}

		parentDir := filepath.Dir(startDir)
		if parentDir == startDir { // Reached root directory
			break
		}
		startDir = parentDir
	}

	return nil, fmt.Errorf("config file not found in any parent directory")
}

func (c *JotlConfig) SetLogLevel(level string) error {
	switch LogLevel(level) {
	case Debug, Info, Warn, Error:
		c.Logging.Level = LogLevel(level)
		return nil
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
}

func (c *JotlConfig) SetProjectName(name string) {
	c.Project.Name = name
}

func (c *JotlConfig) SetDescription(desc string) {
	c.Project.Description = desc
}
