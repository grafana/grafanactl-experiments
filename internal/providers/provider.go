package providers

import "github.com/spf13/cobra"

// Provider defines the interface for a grafanactl provider.
// Providers extend grafanactl with commands for managing Grafana Cloud
// product resources (e.g., SLO, Synthetic Monitoring, OnCall).
type Provider interface {
	// Name returns the unique identifier for this provider.
	Name() string

	// ShortDesc returns a one-line description of the provider.
	ShortDesc() string

	// Commands returns the Cobra commands contributed by this provider.
	Commands() []*cobra.Command

	// Validate checks that the given provider configuration is valid.
	Validate(cfg map[string]string) error

	// ConfigKeys returns the configuration keys used by this provider.
	ConfigKeys() []string
}
