package process

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProcessConfig represents the configuration for an external tool execution.
type ProcessConfig struct {
	Name        string            `yaml:"name" json:"name"`
	Command     string            `yaml:"command" json:"command"`
	Args        []string          `yaml:"args" json:"args"`
	Environment map[string]string `yaml:"env" json:"env"`
	Description string            `yaml:"description" json:"description"`
}

// ConfigFile represents the structure of tools.yaml
type ConfigFile struct {
	Tools []ProcessConfig `yaml:"tools" json:"tools"`
}

// LoadTools reads a configuration file (YAML or JSON) and returns a map of tool names to configs.
func LoadTools(path string) (map[string]ProcessConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty map if file doesn't exist, treating it as "no tools configured"
			// unless the user explicitly requested a file that is missing (handled by caller?)
			// For now, if default path is missing, we return empty.
			return map[string]ProcessConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read tools config: %w", err)
	}

	var cfg ConfigFile
	ext := strings.ToLower(filepath.Ext(path))

	if ext == ".json" {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse tools.json: %w", err)
		}
	} else {
		// Default to YAML
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse tools.yaml: %w", err)
		}
	}

	toolMap := make(map[string]ProcessConfig)
	for _, tool := range cfg.Tools {
		if tool.Name == "" {
			continue
		}
		toolMap[tool.Name] = tool
	}

	return toolMap, nil
}
