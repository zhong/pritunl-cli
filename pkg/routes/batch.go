package routes

import (
	"fmt"

	pritunl "github.com/zhong/pritunl-go-sdk"
)

// BatchAddOptions contains options for batch adding routes.
type BatchAddOptions struct {
	ServerID      string
	Routes        []RouteData
	SkipValidation bool
	SkipConfirm   bool
}

// BatchAdd performs batch addition of routes with validation and conflict detection.
func BatchAdd(client *pritunl.Client, opts BatchAddOptions) (*BatchResult, error) {
	result := &BatchResult{
		ServerID: opts.ServerID,
	}

	// Validate routes if not skipped
	if !opts.SkipValidation {
		validator := NewValidator()
		errors := validator.ValidateRoutes(opts.Routes)
		if validator.HasErrors() {
			result.ValidationErrors = errors
			return result, fmt.Errorf("validation failed with %d errors", len(errors))
		}
	}

	// Deduplicate routes
	opts.Routes = Deduplicate(opts.Routes)
	result.InputCount = len(opts.Routes)

	// Get current routes
	server, err := client.GetServer(opts.ServerID)
	if err != nil {
		return result, fmt.Errorf("get server: %w", err)
	}
	result.CurrentCount = len(server.Routes)

	// Check for conflicts
	toAdd, duplicates := CheckConflicts(opts.Routes, server.Routes)
	result.Duplicates = len(duplicates)
	result.ToAdd = toAdd

	// If all routes are duplicates, return early
	if len(toAdd) == 0 {
		result.SkipReason = "All routes already exist on server"
		return result, nil
	}

	// Add routes
	addedCount := 0
	var addErrors []string

	for _, route := range toAdd {
		if err := client.AddRoute(opts.ServerID, route.ToSDKRoute()); err != nil {
			addErrors = append(addErrors, fmt.Sprintf("%s: %v", route.Network, err))
		} else {
			addedCount++
		}
	}

	result.AddedCount = addedCount
	result.Errors = addErrors

	if len(addErrors) > 0 {
		return result, fmt.Errorf("added %d routes with %d errors", addedCount, len(addErrors))
	}

	return result, nil
}

// BatchResult contains the result of a batch add operation.
type BatchResult struct {
	ServerID           string
	InputCount         int
	CurrentCount       int
	ToAdd              []RouteData
	Duplicates         int
	AddedCount         int
	Errors             []string
	ValidationErrors   []ValidationError
	SkipReason         string
}

// Summary returns a summary of the batch operation.
func (r *BatchResult) Summary() string {
	if r.SkipReason != "" {
		return fmt.Sprintf("Skipped: %s", r.SkipReason)
	}

	if len(r.Errors) > 0 {
		return fmt.Sprintf("Added %d/%d routes. %d errors. %d duplicates.",
			r.AddedCount, len(r.ToAdd), len(r.Errors), r.Duplicates)
	}

	return fmt.Sprintf("✅ Successfully added %d routes (duplicates: %d, total on server: %d)",
		r.AddedCount, r.Duplicates, r.CurrentCount+r.AddedCount)
}
