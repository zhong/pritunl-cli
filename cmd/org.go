package cmd

import (
	"flag"
	"fmt"
	"os"

	pritunl "github.com/example/pritunl-go-sdk"
	"github.com/example/pritunl-cli/pkg/output"
)

// OrgCmd handles organization subcommands.
func OrgCmd() error {
	if len(os.Args) < 2 {
		fmt.Println(`Usage: pritunl org <subcommand> [options]

Subcommands:
  list              List all organizations
  get <id>          Get a specific organization

Options:
  -output format    Output format (table, json, yaml)
  -token string     API token
  -secret string    API secret
  -base string      Base URL
  -insecure         Skip TLS verification
`)
		return nil
	}

	subCmd := os.Args[1]
	fs := flag.NewFlagSet("org", flag.ContinueOnError)

	outputFormat := fs.String("output", "table", "Output format")
	token := fs.String("token", "", "API token")
	secret := fs.String("secret", "", "API secret")
	base := fs.String("base", "", "Base URL")
	insecure := fs.Bool("insecure", false, "Skip TLS verification")

	os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	fs.Parse(os.Args[1:])

	cfg, err := loadConfigWithOverrides(*token, *secret, *base, *insecure)
	if err != nil {
		return err
	}

	client := pritunl.NewClient(cfg.BaseURL, cfg.APIToken, cfg.APISecret, cfg.Insecure)
	formatter := output.NewFormatter(*outputFormat)

	switch subCmd {
	case "list":
		orgs, err := client.ListOrganizations()
		if err != nil {
			return fmt.Errorf("list organizations: %w", err)
		}

		if *outputFormat == "table" {
			table := &output.Table{
				Headers: []string{"ID", "Name", "Users"},
				Rows:    make([][]string, len(orgs)),
			}
			for i, o := range orgs {
				table.Rows[i] = []string{o.ID, o.Name, fmt.Sprintf("%d", o.UserCount)}
			}
			return formatter.OutputTable(table)
		}
		return formatter.Output(orgs)

	case "get":
		args := fs.Args()
		if len(args) == 0 {
			return fmt.Errorf("organization ID required")
		}

		org, err := client.GetOrganization(args[0])
		if err != nil {
			return fmt.Errorf("get organization: %w", err)
		}

		return formatter.Output(org)

	default:
		return fmt.Errorf("unknown org subcommand: %s", subCmd)
	}
}
