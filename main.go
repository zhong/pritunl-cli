package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zhong/pritunl-cli/cmd"
)

func main() {
	// Parse global flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Pritunl CLI - Command-line management tool for Pritunl VPN
Version: 1.0.0

Usage:
  pritunl <command> [options]

Commands:
  status              Get Pritunl server status
  server              Manage VPN servers
  org                 Manage organizations
  user                Manage users
  routes              Manage server routes (batch operations supported)
  config              Manage CLI configuration

Global Options:
  -help               Show this help message
  -token string       API token (or env PRITUNL_API_TOKEN)
  -secret string      API secret (or env PRITUNL_API_SECRET)
  -base string        Base URL (or env PRITUNL_BASE_URL)
  -insecure           Skip TLS verification

Examples:
  pritunl status
  pritunl server list
  pritunl routes batch-add <server-id> --file routes.json
  pritunl config init

For more information, visit: https://github.com/zhong/pritunl-cli
`)
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	cmdName := os.Args[1]

	// Route to subcommand
	switch cmdName {
	case "status":
		if err := cmd.StatusCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "server":
		os.Args = append([]string{filepath.Base(os.Args[0])}, os.Args[2:]...)
		if err := cmd.ServerCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "org":
		os.Args = append([]string{filepath.Base(os.Args[0])}, os.Args[2:]...)
		if err := cmd.OrgCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "user":
		os.Args = append([]string{filepath.Base(os.Args[0])}, os.Args[2:]...)
		if err := cmd.UserCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "routes":
		os.Args = append([]string{filepath.Base(os.Args[0])}, os.Args[2:]...)
		if err := cmd.RoutesCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "config":
		os.Args = append([]string{filepath.Base(os.Args[0])}, os.Args[2:]...)
		if err := cmd.ConfigCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "-help", "--help", "-h", "help":
		flag.Usage()

	case "version", "-v", "--version":
		fmt.Println("pritunl version 1.0.0")

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmdName)
		flag.Usage()
		os.Exit(1)
	}
}
