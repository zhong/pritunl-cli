package routes

import (
	"fmt"
	"net"
	"strings"

	pritunl "github.com/zhong/pritunl-go-sdk"
)

// ValidationError represents a validation error for a route.
type ValidationError struct {
	Index   int
	Network string
	Message string
}

// Validator validates routes.
type Validator struct {
	errors []ValidationError
}

// NewValidator creates a new validator.
func NewValidator() *Validator {
	return &Validator{errors: []ValidationError{}}
}

// ValidateNetwork validates a single network CIDR.
func (v *Validator) ValidateNetwork(network string) bool {
	_, _, err := net.ParseCIDR(network)
	return err == nil
}

// ValidateRoutes validates a slice of routes.
func (v *Validator) ValidateRoutes(routes []RouteData) []ValidationError {
	v.errors = nil
	seen := make(map[string]bool)

	for i, r := range routes {
		// Check required fields
		if strings.TrimSpace(r.Network) == "" {
			v.errors = append(v.errors, ValidationError{
				Index:   i,
				Network: r.Network,
				Message: "network is required",
			})
			continue
		}

		// Validate CIDR format
		if !v.ValidateNetwork(r.Network) {
			v.errors = append(v.errors, ValidationError{
				Index:   i,
				Network: r.Network,
				Message: fmt.Sprintf("invalid CIDR format: %s", r.Network),
			})
			continue
		}

		// Check for duplicates
		if seen[r.Network] {
			v.errors = append(v.errors, ValidationError{
				Index:   i,
				Network: r.Network,
				Message: "duplicate network (already appears in this batch)",
			})
			continue
		}
		seen[r.Network] = true

		// Validate NAT configuration
		if r.NAT && r.NATInterface == "" {
			v.errors = append(v.errors, ValidationError{
				Index:   i,
				Network: r.Network,
				Message: "nat_interface is required when nat=true",
			})
		}

		// Validate metric - allow 0 (will be set to default 100) or 1-32767
		if r.Metric != 0 && (r.Metric < 1 || r.Metric > 32767) {
			v.errors = append(v.errors, ValidationError{
				Index:   i,
				Network: r.Network,
				Message: fmt.Sprintf("metric must be between 1 and 32767, got %d", r.Metric),
			})
		}
	}

	return v.errors
}

// HasErrors returns true if there are validation errors.
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors.
func (v *Validator) Errors() []ValidationError {
	return v.errors
}

// ErrorsStr returns validation errors as a formatted string.
func (v *Validator) ErrorsStr() string {
	if len(v.errors) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Validation errors:\n")
	for _, e := range v.errors {
		sb.WriteString(fmt.Sprintf("  Row %d (%s): %s\n", e.Index+1, e.Network, e.Message))
	}
	return sb.String()
}

// CheckConflicts checks for conflicts with existing routes.
func CheckConflicts(newRoutes []RouteData, existingRoutes []pritunl.ServerRoute) ([]RouteData, []RouteData) {
	existingMap := make(map[string]bool)
	for _, r := range existingRoutes {
		existingMap[r.Network] = true
	}

	var duplicates, toAdd []RouteData
	for _, r := range newRoutes {
		if existingMap[r.Network] {
			duplicates = append(duplicates, r)
		} else {
			toAdd = append(toAdd, r)
		}
	}

	return toAdd, duplicates
}

// Deduplicate removes duplicate routes from the slice.
func Deduplicate(routes []RouteData) []RouteData {
	seen := make(map[string]bool)
	var result []RouteData

	for _, r := range routes {
		if !seen[r.Network] {
			result = append(result, r)
			seen[r.Network] = true
		}
	}

	return result
}
