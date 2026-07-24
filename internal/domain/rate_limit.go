package domain

// RateLimitInfo describes the rate limit status for an API endpoint.
type RateLimitInfo struct {
	Endpoint   string `json:"endpoint"`
	Limit      int    `json:"limit"`       // max requests per window
	Remaining  int    `json:"remaining"`   // remaining requests
	WindowSec  int    `json:"window_sec"`  // window size in seconds
	ResetAt    int64  `json:"reset_at"`    // unix timestamp when window resets
}
