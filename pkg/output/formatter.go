package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Formatter formats and outputs data.
type Formatter struct {
	Format string // "table", "json", "yaml"
	Writer io.Writer
}

// NewFormatter creates a new formatter.
func NewFormatter(format string) *Formatter {
	if format == "" {
		format = "table"
	}
	return &Formatter{
		Format: strings.ToLower(format),
		Writer: os.Stdout,
	}
}

// Table outputs data as a formatted table.
type Table struct {
	Headers []string
	Rows    [][]string
}

// Output outputs the data in the configured format.
func (f *Formatter) Output(data interface{}) error {
	switch f.Format {
	case "json":
		return f.outputJSON(data)
	case "yaml":
		return f.outputYAML(data)
	case "table":
		return f.outputTable(data)
	default:
		return fmt.Errorf("unsupported output format: %s", f.Format)
	}
}

// OutputTable outputs a table.
func (f *Formatter) OutputTable(table *Table) error {
	if f.Format != "table" {
		// For non-table formats, convert to slice of maps
		rows := make([]map[string]string, len(table.Rows))
		for i, row := range table.Rows {
			m := make(map[string]string)
			for j, header := range table.Headers {
				if j < len(row) {
					m[header] = row[j]
				}
			}
			rows[i] = m
		}
		return f.Output(rows)
	}

	// Simple table output without external dependencies
	if len(table.Rows) == 0 {
		fmt.Fprintln(f.Writer, "(no data)")
		return nil
	}

	// Calculate column widths
	widths := make([]int, len(table.Headers))
	for i, h := range table.Headers {
		widths[i] = len(h)
	}
	for _, row := range table.Rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	for i, h := range table.Headers {
		fmt.Fprint(f.Writer, h)
		if i < len(table.Headers)-1 {
			fmt.Fprint(f.Writer, strings.Repeat(" ", widths[i]-len(h)+2))
		}
	}
	fmt.Fprintln(f.Writer)

	// Print separator
	for i, w := range widths {
		fmt.Fprint(f.Writer, strings.Repeat("-", w))
		if i < len(widths)-1 {
			fmt.Fprint(f.Writer, "  ")
		}
	}
	fmt.Fprintln(f.Writer)

	// Print rows
	for _, row := range table.Rows {
		for i, cell := range row {
			fmt.Fprint(f.Writer, cell)
			if i < len(table.Headers)-1 {
				fmt.Fprint(f.Writer, strings.Repeat(" ", widths[i]-len(cell)+2))
			}
		}
		fmt.Fprintln(f.Writer)
	}

	return nil
}

func (f *Formatter) outputJSON(data interface{}) error {
	enc := json.NewEncoder(f.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (f *Formatter) outputYAML(data interface{}) error {
	enc := yaml.NewEncoder(f.Writer)
	defer enc.Close()
	return enc.Encode(data)
}

func (f *Formatter) outputTable(data interface{}) error {
	// Generic table output for slices
	switch v := data.(type) {
	case []map[string]interface{}:
		if len(v) == 0 {
			fmt.Fprintln(f.Writer, "(no data)")
			return nil
		}

		// Get headers from first row
		var headers []string
		for k := range v[0] {
			headers = append(headers, k)
		}

		// Sort headers for consistency
		for i := 0; i < len(headers); i++ {
			for j := i + 1; j < len(headers); j++ {
				if headers[i] > headers[j] {
					headers[i], headers[j] = headers[j], headers[i]
				}
			}
		}

		table := &Table{
			Headers: headers,
			Rows:    make([][]string, len(v)),
		}

		for i, row := range v {
			cells := make([]string, len(headers))
			for j, h := range headers {
				if val, ok := row[h]; ok {
					cells[j] = fmt.Sprintf("%v", val)
				}
			}
			table.Rows[i] = cells
		}

		return f.OutputTable(table)
	default:
		// Default: output as JSON for non-table types
		return f.outputJSON(data)
	}
}

// PrintSuccess prints a success message.
func (f *Formatter) PrintSuccess(msg string) {
	fmt.Fprintf(f.Writer, "✅ %s\n", msg)
}

// PrintError prints an error message.
func (f *Formatter) PrintError(msg string) {
	fmt.Fprintf(f.Writer, "❌ %s\n", msg)
}

// PrintWarning prints a warning message.
func (f *Formatter) PrintWarning(msg string) {
	fmt.Fprintf(f.Writer, "⚠️  %s\n", msg)
}

// PrintInfo prints an info message.
func (f *Formatter) PrintInfo(msg string) {
	fmt.Fprintf(f.Writer, "ℹ️  %s\n", msg)
}
