package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	pritunl "github.com/zhong/pritunl-go-sdk"
)

// DebugCmd provides debugging utilities
func DebugCmd() error {
	if len(os.Args) < 2 {
		fmt.Println(`Usage: pritunl debug <subcommand> [options]

Subcommands:
  server <id>       Show raw server JSON response
  routes <id>       Show raw routes from server response

Options:
  -token string     API token
  -secret string    API secret
  -base string      Base URL
  -insecure         Skip TLS verification
`)
		return nil
	}

	subCmd := os.Args[1]
	fs := flag.NewFlagSet("debug", flag.ContinueOnError)

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
	args := fs.Args()

	switch subCmd {
	case "server":
		if len(args) == 0 {
			return fmt.Errorf("server ID required")
		}
		return debugServer(client, args[0])
	case "routes":
		if len(args) == 0 {
			return fmt.Errorf("server ID required")
		}
		return debugRoutes(client, args[0])
	default:
		return fmt.Errorf("unknown debug subcommand: %s", subCmd)
	}
}

func debugServer(client *pritunl.Client, serverID string) error {
	server, err := client.GetServer(serverID)
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}

	jsonData, _ := json.MarshalIndent(server, "", "  ")
	fmt.Println(string(jsonData))
	return nil
}

func debugRoutes(client *pritunl.Client, serverID string) error {
	server, err := client.GetServer(serverID)
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}

	fmt.Printf("Routes Count: %d\n", len(server.Routes))
	fmt.Printf("Routes Field Present: %v\n", server.Routes != nil)
	
	jsonData, _ := json.MarshalIndent(server.Routes, "", "  ")
	fmt.Println(string(jsonData))
	return nil
}
