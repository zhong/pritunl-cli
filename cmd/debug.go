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
  server <id>         Show raw server JSON response
  routes <id>         Show raw routes from server response
  test-delete <id> <network>  Test different methods to delete a route

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
	case "test-delete":
		if len(args) < 2 {
			return fmt.Errorf("server ID and network required")
		}
		return debugTestDelete(client, args[0], args[1])
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

	if len(server.Routes) > 0 {
		fmt.Println("\nRoutes Details:")
		for i, r := range server.Routes {
			fmt.Printf("[%d] Network: %s, NAT: %v, NetGateway: %v\n", i, r.Network, r.NAT, r.NetGateway)
		}
	}

	jsonData, _ := json.MarshalIndent(server.Routes, "", "  ")
	fmt.Println(string(jsonData))
	return nil
}

func debugTestDelete(client *pritunl.Client, serverID, network string) error {
	fmt.Printf("Testing route deletion for: %s from server: %s\n\n", network, serverID)

	// Verify route exists
	routes, err := client.GetServerRoutes(serverID)
	if err != nil {
		return fmt.Errorf("get routes: %w", err)
	}

	found := false
	for _, r := range routes {
		if r.Network == network {
			found = true
			fmt.Printf("✓ Route found: %s (NAT: %v, NetGateway: %v)\n\n", r.Network, r.NAT, r.NetGateway)
			break
		}
	}

	if !found {
		return fmt.Errorf("route %s not found", network)
	}

	// Try the standard DELETE method
	fmt.Println("Testing DELETE method:")
	err = client.DeleteRoute(serverID, network)
	if err == nil {
		fmt.Println("✓ SUCCESS: Route deleted successfully!")
		return nil
	}

	fmt.Printf("✗ Failed: %v\n\n", err)

	// If DELETE failed, we need to investigate alternative methods
	fmt.Println("Investigating alternative deletion methods...")
	fmt.Println("\nPossible reasons for 404:")
	fmt.Println("1. API endpoint format is different")
	fmt.Println("2. URL encoding is incorrect")
	fmt.Println("3. Pritunl requires a different HTTP method")
	fmt.Println("4. Pritunl requires the route to be updated via server PUT endpoint")
	fmt.Println("5. This version of Pritunl doesn't support route deletion")

	return fmt.Errorf("delete failed, investigate further")
}
