package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration.
type Config struct {
	BaseURL         string `yaml:"base_url"`
	APIToken        string `yaml:"api_token"`
	APISecret       string `yaml:"api_secret"`
	Insecure        bool   `yaml:"insecure,omitempty"`
	OutputFormat    string `yaml:"output_format,omitempty"`
	DefaultServerID string `yaml:"default_server_id,omitempty"`
}

const (
	configDir  = ".pritunl"
	configFile = "config.yaml"
)

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configFile), nil
}

// Load loads configuration from file.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s. Run 'pritunl config init' to create one", path)
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{
		Insecure:     true,
		OutputFormat: "table",
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Validate required fields
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base_url is required in config")
	}
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("api_token is required in config")
	}
	if cfg.APISecret == "" {
		return nil, fmt.Errorf("api_secret is required in config")
	}

	return cfg, nil
}

// LoadOrEnv loads config from file or environment variables.
func LoadOrEnv() (*Config, error) {
	// Try environment variables first
	cfg := &Config{
		BaseURL:      os.Getenv("PRITUNL_BASE_URL"),
		APIToken:     os.Getenv("PRITUNL_API_TOKEN"),
		APISecret:    os.Getenv("PRITUNL_API_SECRET"),
		Insecure:     os.Getenv("PRITUNL_INSECURE") == "true",
		OutputFormat: os.Getenv("PRITUNL_OUTPUT_FORMAT"),
	}

	// If env vars not set, try config file
	if cfg.BaseURL == "" || cfg.APIToken == "" {
		fileCfg, err := Load()
		if err == nil {
			return fileCfg, nil
		}
		// If both env and file fail, return error
		if cfg.BaseURL == "" || cfg.APIToken == "" {
			return nil, fmt.Errorf("config not found. Set PRITUNL_BASE_URL, PRITUNL_API_TOKEN, PRITUNL_API_SECRET env vars or run 'pritunl config init'")
		}
	}

	// Set defaults
	if cfg.OutputFormat == "" {
		cfg.OutputFormat = "table"
	}

	return cfg, nil
}

// Save saves configuration to file.
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}
