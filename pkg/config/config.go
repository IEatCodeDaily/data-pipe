package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the pipeline configuration
type Config struct {
	Pipeline PipelineConfig `json:"pipeline"`
	Source   SourceConfig   `json:"source"`
	Sink     SinkConfig     `json:"sink"`
}

// PipelineConfig contains pipeline-level settings
type PipelineConfig struct {
	Name string `json:"name"`
}

// SourceConfig contains source configuration
type SourceConfig struct {
	Type     string                 `json:"type"` // mongodb, convex, etc.
	Settings map[string]interface{} `json:"settings"`
}

// SinkConfig contains sink configuration
type SinkConfig struct {
	Type     string                 `json:"type"` // postgresql, clickhouse, etc.
	Settings map[string]interface{} `json:"settings"`
}

// LoadFromFile loads configuration from a JSON file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetString safely retrieves a string from settings
func (s SourceConfig) GetString(key string) string {
	if val, ok := s.Settings[key].(string); ok {
		return val
	}
	return ""
}

// GetString safely retrieves a string from settings
func (s SinkConfig) GetString(key string) string {
	if val, ok := s.Settings[key].(string); ok {
		return val
	}
	return ""
}
