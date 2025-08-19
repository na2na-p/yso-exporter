package version

import "github.com/na2na/yso-exporter/internal/config"

//go:generate mockgen -source=resolver.go -destination=mocks/mock_resolver.go -package=mocks

// AddressResolver defines the interface for resolving memory addresses
// based on game version configuration.
// Purpose: Provide address resolution capabilities for different game versions
type AddressResolver interface {
	// ResolveAddress resolves a memory address for the given metric name
	// Returns the resolved address configuration or error if not found
	ResolveAddress(metricName string) (*config.AddressConfig, error)
	
	// GetProcessName returns the process name for the current version
	GetProcessName() string
	
	// GetVersionName returns the version name being resolved
	GetVersionName() string
}