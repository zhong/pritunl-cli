package cmd

import "errors"

var (
	ErrMissingBaseURL = errors.New("base URL is required. Set PRITUNL_BASE_URL or run 'pritunl config init'")
	ErrMissingToken   = errors.New("API token is required. Set PRITUNL_API_TOKEN or run 'pritunl config init'")
	ErrMissingSecret  = errors.New("API secret is required. Set PRITUNL_API_SECRET or run 'pritunl config init'")
)
