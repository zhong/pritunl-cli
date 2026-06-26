package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/example/pritunl-cli/pkg/config"
	"github.com/example/pritunl-cli/pkg/output"
)

// ConfigCmd handles configuration subcommands.
func ConfigCmd() error {
	if len(os.Args) < 2 {
		fmt.Println(`Usage: pritunl config <subcommand>

Subcommands:
  init              Initialize configuration interactively
  show              Display current configuration
  set <key> <val>   Set a configuration value
  clear             Clear configuration

Examples:
  pritunl config init
  pritunl config show
  pritunl config set base_url https://pritunl.example.com
`)
		return nil
	}

	subCmd := os.Args[1]

	switch subCmd {
	case "init":
		return configInit()
	case "show":
		return configShow()
	case "set":
		if len(os.Args) < 4 {
			return fmt.Errorf("key and value required")
		}
		return configSet(os.Args[2], os.Args[3])
	case "clear":
		return configClear()
	default:
		return fmt.Errorf("unknown config subcommand: %s", subCmd)
	}
}

func configInit() error {
	reader := bufio.NewReader(os.Stdin)
	formatter := output.NewFormatter("table")

	cfg := &config.Config{
		Insecure:     true,
		OutputFormat: "table",
	}

	fmt.Println("Pritunl CLI Configuration")
	fmt.Println("=========================================")

	// Base URL
	fmt.Print("Base URL (e.g., https://pritunl.example.com): ")
	input, _ := reader.ReadString('\n')
	cfg.BaseURL = strings.TrimSpace(input)

	// API Token
	fmt.Print("API Token: ")
	input, _ = reader.ReadString('\n')
	cfg.APIToken = strings.TrimSpace(input)

	// API Secret
	fmt.Print("API Secret: ")
	input, _ = reader.ReadString('\n')
	cfg.APISecret = strings.TrimSpace(input)

	// Insecure
	fmt.Print("Skip TLS verification? (y/n, default: y): ")
	input, _ = reader.ReadString('\n')
	cfg.Insecure = strings.ToLower(strings.TrimSpace(input)) != "n"

	// Output Format
	fmt.Print("Output format (table/json/yaml, default: table): ")
	input, _ = reader.ReadString('\n')
	if f := strings.TrimSpace(input); f != "" {
		cfg.OutputFormat = f
	}

	// Save
	if err := cfg.Save(); err != nil {
		formatter.PrintError(fmt.Sprintf("Failed to save config: %v", err))
		return err
	}

	path, _ := config.ConfigPath()
	formatter.PrintSuccess(fmt.Sprintf("Configuration saved to %s", path))
	return nil
}

func configShow() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	formatter := output.NewFormatter("table")
	table := &output.Table{
		Headers: []string{"Setting", "Value"},
		Rows: [][]string{
			{"Base URL", cfg.BaseURL},
			{"API Token", maskSecret(cfg.APIToken)},
			{"API Secret", maskSecret(cfg.APISecret)},
			{"Insecure", fmt.Sprintf("%v", cfg.Insecure)},
			{"Output Format", cfg.OutputFormat},
			{"Default Server ID", cfg.DefaultServerID},
		},
	}

	return formatter.OutputTable(table)
}

func configSet(key, value string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	switch key {
	case "base_url":
		cfg.BaseURL = value
	case "api_token":
		cfg.APIToken = value
	case "api_secret":
		cfg.APISecret = value
	case "insecure":
		cfg.Insecure = strings.ToLower(value) == "true"
	case "output_format":
		cfg.OutputFormat = value
	case "default_server_id":
		cfg.DefaultServerID = value
	default:
		return fmt.Errorf("unknown setting: %s", key)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	formatter := output.NewFormatter("table")
	formatter.PrintSuccess(fmt.Sprintf("Set %s = %s", key, value))
	return nil
}

func configClear() error {
	path, err := config.ConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear config: %w", err)
	}

	formatter := output.NewFormatter("table")
	formatter.PrintSuccess("Configuration cleared")
	return nil
}

func maskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "..." + s[len(s)-4:]
}
