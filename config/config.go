package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/anuramat/modagent/junior"
	"github.com/anuramat/modagent/logworm"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Tools map[string]ToolConfig `yaml:"tools"`
}

type ToolConfig struct {
	Description Description `yaml:"description"`
}

type Description struct {
	Text *string `yaml:"text,omitempty"`
	Path *string `yaml:"path,omitempty"`
}

const (
	configDirName  = "modagent"
	configFileName = "config.yaml"
)

var validToolNames = []string{"junior-r", "junior-rwx", "logworm"}

func LoadConfig() (*Config, error) {
	configPath := filepath.Join(xdg.ConfigHome, configDirName, configFileName)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{Tools: make(map[string]ToolConfig)}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.Tools == nil {
		cfg.Tools = make(map[string]ToolConfig)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	validTools := make(map[string]bool)
	for _, tool := range validToolNames {
		validTools[tool] = true
	}

	for toolName, toolConfig := range cfg.Tools {
		if !validTools[toolName] {
			return fmt.Errorf("unknown tool name in config: %s (valid tools: %v)", toolName, validToolNames)
		}

		// Validate mutually exclusive text/path
		desc := toolConfig.Description
		if desc.Text != nil && desc.Path != nil {
			return fmt.Errorf("tool %s: text and path are mutually exclusive", toolName)
		}

		// Validate file exists if path is specified
		if desc.Path != nil {
			configDir := filepath.Join(xdg.ConfigHome, configDirName)
			filePath := *desc.Path
			if !filepath.IsAbs(filePath) {
				filePath = filepath.Join(configDir, filePath)
			}
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return fmt.Errorf("tool %s: description file not found: %s", toolName, *desc.Path)
			}
		}
	}
	return nil
}

func GenerateDefaultConfig() error {
	configDir := filepath.Join(xdg.ConfigHome, configDirName)
	configPath := filepath.Join(configDir, configFileName)

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	juniorRDesc := junior.Description + " (read-only mode)"
	juniorRWXDesc := junior.Description + " (full access mode)"
	logwormDesc := logworm.Description

	defaultConfig := Config{
		Tools: map[string]ToolConfig{
			"junior-r": {
				Description: Description{
					Text: &juniorRDesc,
				},
			},
			"junior-rwx": {
				Description: Description{
					Text: &juniorRWXDesc,
				},
			},
			"logworm": {
				Description: Description{
					Text: &logwormDesc,
				},
			},
		},
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Default config generated at: %s\n", configPath)
	return nil
}

func (c *Config) GetToolDescription(toolName, defaultDesc string) string {
	toolConfig, exists := c.Tools[toolName]
	if !exists {
		return defaultDesc
	}

	desc := toolConfig.Description

	// Handle text description
	if desc.Text != nil {
		return *desc.Text
	}

	// Handle path description
	if desc.Path != nil {
		configDir := filepath.Join(xdg.ConfigHome, configDirName)
		filePath := *desc.Path
		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(configDir, filePath)
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			// Return default if file can't be read (validation should have caught this)
			return defaultDesc
		}
		return string(content)
	}

	// Neither text nor path specified, return default
	return defaultDesc
}
