package routes

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	pritunl "github.com/example/pritunl-go-sdk"
)

// RouteData represents a route to be added.
type RouteData struct {
	Network         string `json:"network"`
	Comment         string `json:"comment,omitempty"`
	Metric          int    `json:"metric,omitempty"`
	NAT             bool   `json:"nat,omitempty"`
	NATInterface    string `json:"nat_interface,omitempty"`
	NATNetmap       string `json:"nat_netmap,omitempty"`
	Advertise       bool   `json:"advertise,omitempty"`
	AdvertiseResource string `json:"advertise_resource,omitempty"`
	NetGateway      bool   `json:"net_gateway,omitempty"`
}

// LoadFromJSON loads routes from a JSON file.
func LoadFromJSON(filename string) ([]RouteData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var routes []RouteData
	if err := json.Unmarshal(data, &routes); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	return routes, nil
}

// LoadFromCSV loads routes from a CSV file.
// Expected header: network,comment,metric,nat,nat_interface,advertise,net_gateway
func LoadFromCSV(filename string) ([]RouteData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV must have header and at least one data row")
	}

	// Parse header
	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Validate required fields
	if _, ok := headerMap["network"]; !ok {
		return nil, fmt.Errorf("CSV must have 'network' column")
	}

	var routes []RouteData
	for i, record := range records[1:] {
		route := RouteData{Metric: 100}

		// Parse network
		if idx, ok := headerMap["network"]; ok && idx < len(record) {
			route.Network = strings.TrimSpace(record[idx])
		}

		// Parse comment
		if idx, ok := headerMap["comment"]; ok && idx < len(record) {
			route.Comment = strings.TrimSpace(record[idx])
		}

		// Parse metric
		if idx, ok := headerMap["metric"]; ok && idx < len(record) {
			if v := strings.TrimSpace(record[idx]); v != "" {
				fmt.Sscanf(v, "%d", &route.Metric)
			}
		}

		// Parse boolean fields
		if idx, ok := headerMap["nat"]; ok && idx < len(record) {
			route.NAT = strings.ToLower(strings.TrimSpace(record[idx])) == "true"
		}
		if idx, ok := headerMap["advertise"]; ok && idx < len(record) {
			route.Advertise = strings.ToLower(strings.TrimSpace(record[idx])) == "true"
		}
		if idx, ok := headerMap["net_gateway"]; ok && idx < len(record) {
			route.NetGateway = strings.ToLower(strings.TrimSpace(record[idx])) == "true"
		}

		// Parse interface/netmap
		if idx, ok := headerMap["nat_interface"]; ok && idx < len(record) {
			route.NATInterface = strings.TrimSpace(record[idx])
		}
		if idx, ok := headerMap["nat_netmap"]; ok && idx < len(record) {
			route.NATNetmap = strings.TrimSpace(record[idx])
		}

		if route.Network == "" {
			return nil, fmt.Errorf("row %d: network is required", i+2)
		}

		routes = append(routes, route)
	}

	return routes, nil
}

// SaveToJSON saves routes to a JSON file.
func SaveToJSON(routes []RouteData, filename string) error {
	data, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// SaveToCSV saves routes to a CSV file.
func SaveToCSV(routes []RouteData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"network", "comment", "metric", "nat", "nat_interface", "advertise", "net_gateway"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Write rows
	for _, r := range routes {
		row := []string{
			r.Network,
			r.Comment,
			fmt.Sprintf("%d", r.Metric),
			fmt.Sprintf("%v", r.NAT),
			r.NATInterface,
			fmt.Sprintf("%v", r.Advertise),
			fmt.Sprintf("%v", r.NetGateway),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}

	return nil
}

// ToSDKRoute converts RouteData to SDK ServerRoute.
func (r RouteData) ToSDKRoute() pritunl.ServerRoute {
	return pritunl.ServerRoute{
		Network:    r.Network,
		NAT:        r.NAT,
		NetGateway: r.NetGateway,
	}
}

// LoadFromReader loads routes from an io.Reader (JSON or CSV based on format).
func LoadFromReader(reader io.Reader, format string) ([]RouteData, error) {
	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read data: %w", err)
	}

	format = strings.ToLower(format)
	switch format {
	case "json":
		var routes []RouteData
		if err := json.Unmarshal(data, &routes); err != nil {
			return nil, fmt.Errorf("parse JSON: %w", err)
		}
		return routes, nil

	case "csv":
		reader := csv.NewReader(strings.NewReader(string(data)))
		records, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("parse CSV: %w", err)
		}

		// Write to temp file and use LoadFromCSV
		tmpfile, err := os.CreateTemp("", "pritunl-*.csv")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmpfile.Name())

		for _, record := range records {
			fmt.Fprintln(tmpfile, strings.Join(record, ","))
		}
		tmpfile.Close()

		return LoadFromCSV(tmpfile.Name())

	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
