package config

import (
	"os"
)

// These variables can be set via -ldflags at build time.
var (
	TailscaleApiToken string
	TailscaleTailnet  string
	ServiceDir        string
)

// GetTailscaleApiToken returns the API token, preferring the compiled value, then env.
func GetTailscaleApiToken() string {
	if TailscaleApiToken != "" {
		return TailscaleApiToken
	}
	return os.Getenv("TAILSCALE_API_TOKEN")
}

// GetTailscaleTailnet returns the tailnet, preferring the compiled value, then env.
func GetTailscaleTailnet() string {
	if TailscaleTailnet != "" {
		return TailscaleTailnet
	}
	return os.Getenv("TAILSCALE_TAILNET")
}

// GetServiceDir returns the service dir, preferring the compiled value, then env.
func GetServiceDir() string {
	if ServiceDir != "" {
		return ServiceDir
	}
	return os.Getenv("SERVICE_DIR")
}
