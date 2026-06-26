package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	pritunl "github.com/example/pritunl-go-sdk"
	"github.com/example/pritunl-cli/pkg/output"
	"github.com/example/pritunl-cli/pkg/routes"
)

// RoutesCmd handles route subcommands.
func RoutesCmd() error {
	if len(os.Args) < 2 {
		fmt.Println(`Usage: pritunl routes <subcommand> [options]

Subcommands:
  list <server-id>              List routes for a server
  add <server-id>               Add a single route
  batch-add <server-id>         Add routes from file (JSON or CSV)
  validate                      Validate route file
  export <server-id>            Export server routes to file
  delete <server-id>            Delete a route

Options:
  -file string      JSON or CSV file (for batch-add, validate, export)
  -csv string       CSV file (shorthand for -file with format detection)
  -network string   Network CIDR (for add, delete)
  -comment string   Route comment (for add)
  -metric int       Route metric (for add, default: 100)
  -nat              Enable NAT
  -output format    Output format (table, json, yaml)
  -token string     API token
  -secret string    API secret
  -base string      Base URL
  -insecure         Skip TLS verification
  -skip-confirm     Skip confirmation prompt

Examples:
  pritunl routes list 5f1234567890abcdef000000
  pritunl routes batch-add 5f1234567890abcdef000000 -file routes.json
  pritunl routes batch-add 5f1234567890abcdef000000 -csv routes.csv
  pritunl routes add 5f1234567890abcdef000000 -network 10.0.0.0/24 -comment "Office"
  pritunl routes export 5f1234567890abcdef000000 -file backup.json
`)
		return nil
	}

	subCmd := os.Args[1]
	fs := flag.NewFlagSet("routes", flag.ContinueOnError)

	file := fs.String("file", "", "JSON or CSV file")
	csv := fs.String("csv", "", "CSV file (shorthand)")
	network := fs.String("network", "", "Network CIDR")
	comment := fs.String("comment", "", "Route comment")
	metric := fs.Int("metric", 100, "Route metric")
	nat := fs.Bool("nat", false, "Enable NAT")
	outputFormat := fs.String("output", "table", "Output format")
	token := fs.String("token", "", "API token")
	secret := fs.String("secret", "", "API secret")
	base := fs.String("base", "", "Base URL")
	insecure := fs.Bool("insecure", false, "Skip TLS verification")
	skipConfirm := fs.Bool("skip-confirm", false, "Skip confirmation")

	os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	fs.Parse(os.Args[1:])

	cfg, err := loadConfigWithOverrides(*token, *secret, *base, *insecure)
	if err != nil {
		return err
	}

	client := pritunl.NewClient(cfg.BaseURL, cfg.APIToken, cfg.APISecret, cfg.Insecure)
	formatter := output.NewFormatter(*outputFormat)

	args := fs.Args()

	switch subCmd {
	case "list":
		if len(args) == 0 {
			return fmt.Errorf("server ID required")
		}
		return routesList(client, formatter, args[0], *outputFormat)

	case "add":
		if len(args) == 0 {
			return fmt.Errorf("server ID required")
		}
		if *network == "" {
			return fmt.Errorf("-network flag required")
		}
		return routesAdd(client, formatter, args[0], *network, *comment, *metric, *nat)

	case "batch-add":
		if len(args) == 0 {
			return fmt.Errorf("server ID required")
		}
		if *file == "" && *csv == "" {
			return fmt.Errorf("-file or -csv flag required")
		}
		if *csv != "" {
			*file = *csv
		}
		return routesBatchAdd(client, formatter, args[0], *file, *skipConfirm)

	case "validate":
		if *file == "" && *csv == "" {
			return fmt.Errorf("-file or -csv flag required")
		}
		if *csv != "" {
			*file = *csv
		}
		return routesValidate(formatter, *file)

	case "export":
		if len(args) == 0 {
			return fmt.Errorf("server ID required")
		}
		if *file == "" {
			return fmt.Errorf("-file flag required")
		}
		return routesExport(client, formatter, args[0], *file)

	case "delete":
		if len(args) == 0 {
			return fmt.Errorf("server ID required")
		}
		if *network == "" {
			return fmt.Errorf("-network flag required")
		}
		return routesDelete(client, formatter, args[0], *network)

	default:
		return fmt.Errorf("unknown routes subcommand: %s", subCmd)
	}
}

func routesList(client *pritunl.Client, formatter *output.Formatter, serverID, outputFmt string) error {
	server, err := client.GetServer(serverID)
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}

	if outputFmt == "table" {
		if len(server.Routes) == 0 {
			fmt.Println("(no routes)")
			return nil
		}

		table := &output.Table{
			Headers: []string{"Network", "NAT", "Net Gateway"},
			Rows:    make([][]string, len(server.Routes)),
		}
		for i, r := range server.Routes {
			table.Rows[i] = []string{
				r.Network,
				fmt.Sprintf("%v", r.NAT),
				fmt.Sprintf("%v", r.NetGateway),
			}
		}
		return formatter.OutputTable(table)
	}

	return formatter.Output(server.Routes)
}

func routesAdd(client *pritunl.Client, formatter *output.Formatter, serverID, network, comment string, metric int, nat bool) error {
	route := pritunl.ServerRoute{
		Network:    network,
		NAT:        nat,
		NetGateway: false,
	}

	if err := client.AddRoute(serverID, route); err != nil {
		return fmt.Errorf("add route: %w", err)
	}

	formatter.PrintSuccess(fmt.Sprintf("Added route %s", network))
	return nil
}

func routesBatchAdd(client *pritunl.Client, formatter *output.Formatter, serverID, filename string, skipConfirm bool) error {
	// Detect format from filename
	var routeData []routes.RouteData
	var err error

	if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		routeData, err = routes.LoadFromCSV(filename)
	} else {
		routeData, err = routes.LoadFromJSON(filename)
	}

	if err != nil {
		return fmt.Errorf("load routes: %w", err)
	}

	formatter.PrintInfo(fmt.Sprintf("Loaded %d routes from %s", len(routeData), filename))

	// Validate
	validator := routes.NewValidator()
	validator.ValidateRoutes(routeData)
	if validator.HasErrors() {
		fmt.Println(validator.ErrorsStr())
		return fmt.Errorf("validation failed")
	}

	// Show preview
	if len(routeData) > 0 {
		fmt.Println("\nPreview of routes to add:")
		table := &output.Table{
			Headers: []string{"Network", "Comment", "Metric", "NAT"},
			Rows:    make([][]string, len(routeData)),
		}
		for i, r := range routeData {
			table.Rows[i] = []string{
				r.Network,
				r.Comment,
				fmt.Sprintf("%d", r.Metric),
				fmt.Sprintf("%v", r.NAT),
			}
		}
		formatter.OutputTable(table)
	}

	// Ask for confirmation
	if !skipConfirm {
		fmt.Printf("\nContinue adding %d routes? (y/n): ", len(routeData))
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Perform batch add
	result, batchErr := routes.BatchAdd(client, routes.BatchAddOptions{
		ServerID: serverID,
		Routes:   routeData,
	})

	if batchErr != nil {
		formatter.PrintError(fmt.Sprintf("Batch add failed: %v", batchErr))
		if len(result.Errors) > 0 {
			fmt.Println("Errors:")
			for _, e := range result.Errors {
				fmt.Printf("  - %s\n", e)
			}
		}
		return batchErr
	}

	// Print summary
	fmt.Println("\n" + result.Summary())
	if len(result.ValidationErrors) > 0 {
		fmt.Println("Validation errors:")
		for _, ve := range result.ValidationErrors {
			fmt.Printf("  Row %d: %s\n", ve.Index+1, ve.Message)
		}
	}

	return nil
}

func routesValidate(formatter *output.Formatter, filename string) error {
	var routeData []routes.RouteData
	var err error

	if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		routeData, err = routes.LoadFromCSV(filename)
	} else {
		routeData, err = routes.LoadFromJSON(filename)
	}

	if err != nil {
		return fmt.Errorf("load routes: %w", err)
	}

	validator := routes.NewValidator()
	validator.ValidateRoutes(routeData)

	formatter.PrintInfo(fmt.Sprintf("Validating %d routes", len(routeData)))

	if !validator.HasErrors() {
		formatter.PrintSuccess("All routes are valid")
		return nil
	}

	fmt.Println(validator.ErrorsStr())
	return fmt.Errorf("validation failed with %d errors", len(validator.Errors()))
}

func routesExport(client *pritunl.Client, formatter *output.Formatter, serverID, filename string) error {
	server, err := client.GetServer(serverID)
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}

	// Convert to RouteData
	routeData := make([]routes.RouteData, len(server.Routes))
	for i, r := range server.Routes {
		routeData[i] = routes.RouteData{
			Network:    r.Network,
			NAT:        r.NAT,
			NetGateway: r.NetGateway,
		}
	}

	var saveErr error
	if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		saveErr = routes.SaveToCSV(routeData, filename)
	} else {
		saveErr = routes.SaveToJSON(routeData, filename)
	}

	if saveErr != nil {
		return fmt.Errorf("save routes: %w", saveErr)
	}

	formatter.PrintSuccess(fmt.Sprintf("Exported %d routes to %s", len(routeData), filename))
	return nil
}

func routesDelete(client *pritunl.Client, formatter *output.Formatter, serverID, network string) error {
	if err := client.DeleteRoute(serverID, network); err != nil {
		return fmt.Errorf("delete route: %w", err)
	}

	formatter.PrintSuccess(fmt.Sprintf("Deleted route %s", network))
	return nil
}
