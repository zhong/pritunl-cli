package cmd

import (
	"os"

	"github.com/zhong/pritunl-cli/pkg/config"
)

func loadConfigWithOverrides(token, secret, baseURL string, insecure bool) (*config.Config, error) {
	// Try to load config file first
	cfg, err := config.LoadOrEnv()
	if err != nil {
		// If no config, create minimal config from env/flags
		cfg = &config.Config{
			BaseURL:      os.Getenv("PRITUNL_BASE_URL"),
			APIToken:     os.Getenv("PRITUNL_API_TOKEN"),
			APISecret:    os.Getenv("PRITUNL_API_SECRET"),
			Insecure:     os.Getenv("PRITUNL_INSECURE") == "true",
			OutputFormat: "table",
		}
	}

	// Apply command-line overrides
	if token != "" {
		cfg.APIToken = token
	}
	if secret != "" {
		cfg.APISecret = secret
	}
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if insecure {
		cfg.Insecure = insecure
	}

	// Validate
	if cfg.BaseURL == "" {
		return nil, ErrMissingBaseURL
	}
	if cfg.APIToken == "" {
		return nil, ErrMissingToken
	}
	if cfg.APISecret == "" {
		return nil, ErrMissingSecret
	}

	return cfg, nil
}
