package cmd

import (
	"flag"
	"fmt"
	"os"

	pritunl "github.com/example/pritunl-go-sdk"
	"github.com/example/pritunl-cli/pkg/output"
)

// StatusCmd handles the status command.
func StatusCmd() error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("Usage: pritunl status [options]")
		fmt.Println("\nGet Pritunl server status")
		fmt.Println("\nOptions:")
		fs.PrintDefaults()
	}

	outputFormat := fs.String("output", "table", "Output format: table, json, yaml")
	token := fs.String("token", "", "API token")
	secret := fs.String("secret", "", "API secret")
	base := fs.String("base", "", "Base URL")
	insecure := fs.Bool("insecure", false, "Skip TLS verification")

	fs.Parse(os.Args[1:])

	// Load config
	cfg, err := loadConfigWithOverrides(*token, *secret, *base, *insecure)
	if err != nil {
		return err
	}

	// Create client
	client := pritunl.NewClient(cfg.BaseURL, cfg.APIToken, cfg.APISecret, cfg.Insecure)

	// Get status
	status, err := client.GetStatus()
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}

	// Format output
	formatter := output.NewFormatter(*outputFormat)

	switch *outputFormat {
	case "table":
		table := &output.Table{
			Headers: []string{"Key", "Value"},
			Rows: [][]string{
				{"Version", status.ServerVersion},
				{"Organizations", fmt.Sprintf("%d", status.OrgCount)},
				{"Users", fmt.Sprintf("%d", status.UserCount)},
				{"Users Online", fmt.Sprintf("%d", status.UsersOnline)},
				{"Servers", fmt.Sprintf("%d", status.ServerCount)},
				{"Servers Online", fmt.Sprintf("%d", status.ServersOnline)},
				{"Hosts", fmt.Sprintf("%d", status.HostCount)},
				{"Hosts Online", fmt.Sprintf("%d", status.HostsOnline)},
				{"Current Host", status.CurrentHost},
				{"Public IP", status.PublicIP},
			},
		}
		return formatter.OutputTable(table)
	default:
		return formatter.Output(status)
	}
}
