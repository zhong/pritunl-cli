package cmd

import (
	"flag"
	"fmt"
	"os"

	pritunl "github.com/zhong/pritunl-go-sdk"
	"github.com/zhong/pritunl-cli/pkg/output"
)

// ServerCmd handles server subcommands.
func ServerCmd() error {
	if len(os.Args) < 2 {
		fmt.Println(`Usage: pritunl server <subcommand> [options]

Subcommands:
  list              List all servers
  get <id>          Get a specific server
  create            Create a new server
  delete <id>       Delete a server
  start <id>        Start a server
  stop <id>         Stop a server
  restart <id>      Restart a server

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
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	outputFormat := fs.String("output", "table", "Output format")
	token := fs.String("token", "", "API token")
	secret := fs.String("secret", "", "API secret")
	base := fs.String("base", "", "Base URL")
	insecure := fs.Bool("insecure", false, "Skip TLS verification")

	// Remove subcommand from args for flag parsing
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
		servers, err := client.ListServers()
		if err != nil {
			return fmt.Errorf("list servers: %w", err)
		}

		if *outputFormat == "table" {
			table := &output.Table{
				Headers: []string{"ID", "Name", "Status", "Network", "Organizations"},
				Rows:    make([][]string, len(servers)),
			}
			for i, s := range servers {
				table.Rows[i] = []string{
					s.ID,
					s.Name,
					s.Status,
					s.Network,
					fmt.Sprintf("%d", len(s.Organizations)),
				}
			}
			return formatter.OutputTable(table)
		}
		return formatter.Output(servers)

	case "get":
		args := fs.Args()
		if len(args) == 0 {
			return fmt.Errorf("server ID required")
		}

		server, err := client.GetServer(args[0])
		if err != nil {
			return fmt.Errorf("get server: %w", err)
		}

		return formatter.Output(server)

	case "start", "stop", "restart":
		args := fs.Args()
		if len(args) == 0 {
			return fmt.Errorf("server ID required")
		}

		if err := client.ServerOperation(args[0], subCmd); err != nil {
			return fmt.Errorf("%s server: %w", subCmd, err)
		}

		formatter.PrintSuccess(fmt.Sprintf("Server %sd", subCmd))
		return nil

	default:
		return fmt.Errorf("unknown server subcommand: %s", subCmd)
	}
}
