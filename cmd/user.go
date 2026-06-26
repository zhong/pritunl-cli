package cmd

import (
	"flag"
	"fmt"
	"os"

	pritunl "github.com/zhong/pritunl-go-sdk"
	"github.com/zhong/pritunl-cli/pkg/output"
)

// UserCmd handles user subcommands.
func UserCmd() error {
	if len(os.Args) < 2 {
		fmt.Println(`Usage: pritunl user <subcommand> [options]

Subcommands:
  list <org-id>     List users in organization
  get <org> <user>  Get a specific user

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
	fs := flag.NewFlagSet("user", flag.ContinueOnError)

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
		args := fs.Args()
		if len(args) == 0 {
			return fmt.Errorf("organization ID required")
		}

		users, err := client.ListUsers(args[0])
		if err != nil {
			return fmt.Errorf("list users: %w", err)
		}

		if *outputFormat == "table" {
			table := &output.Table{
				Headers: []string{"ID", "Name", "Email", "Groups", "Type"},
				Rows:    make([][]string, len(users)),
			}
			for i, u := range users {
				table.Rows[i] = []string{
					u.ID,
					u.Name,
					u.Email,
					fmt.Sprintf("%d", len(u.Groups)),
					u.Type,
				}
			}
			return formatter.OutputTable(table)
		}
		return formatter.Output(users)

	case "get":
		args := fs.Args()
		if len(args) < 2 {
			return fmt.Errorf("organization ID and user ID required")
		}

		user, err := client.GetUser(args[0], args[1])
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}

		return formatter.Output(user)

	default:
		return fmt.Errorf("unknown user subcommand: %s", subCmd)
	}
}
