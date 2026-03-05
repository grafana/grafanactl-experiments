package alert

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/grafana/grafanactl/internal/config"
	"github.com/grafana/grafanactl/internal/providers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// AlertProvider manages Grafana alerting resources.
type AlertProvider struct{}

// Name returns the unique identifier for this provider.
func (p *AlertProvider) Name() string { return "alert" }

// ShortDesc returns a one-line description of the provider.
func (p *AlertProvider) ShortDesc() string { return "Manage Grafana alerting resources." }

// Commands returns the Cobra commands contributed by this provider.
func (p *AlertProvider) Commands() []*cobra.Command {
	loader := &configLoader{}

	alertCmd := &cobra.Command{
		Use:   "alert",
		Short: p.ShortDesc(),
	}

	loader.bindFlags(alertCmd.PersistentFlags())

	alertCmd.AddCommand(rulesCommands(loader))
	alertCmd.AddCommand(groupsCommands(loader))

	return []*cobra.Command{alertCmd}
}

// Validate checks that the given provider configuration is valid.
func (p *AlertProvider) Validate(cfg map[string]string) error {
	return nil
}

// ConfigKeys returns the configuration keys used by this provider.
func (p *AlertProvider) ConfigKeys() []providers.ConfigKey {
	return nil
}

// configLoader loads REST config for alert commands.
type configLoader struct {
	configFile string
	ctxName    string
}

func (l *configLoader) bindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&l.configFile, "config", "", "Path to the configuration file to use")
	flags.StringVar(&l.ctxName, "context", "", "Name of the context to use")
}

// LoadRESTConfig loads the REST config from the config file.
func (l *configLoader) LoadRESTConfig(ctx context.Context) (config.NamespacedRESTConfig, error) {
	source := l.configSource()

	overrides := []config.Override{
		func(cfg *config.Config) error {
			if cfg.CurrentContext == "" {
				cfg.CurrentContext = config.DefaultContextName
			}

			if !cfg.HasContext(cfg.CurrentContext) {
				cfg.SetContext(cfg.CurrentContext, true, config.Context{})
			}

			curCtx := cfg.Contexts[cfg.CurrentContext]
			if curCtx.Grafana == nil {
				curCtx.Grafana = &config.GrafanaConfig{}
			}

			if err := env.Parse(curCtx); err != nil {
				return err
			}

			const providerEnvPrefix = "GRAFANA_PROVIDER_"
			for _, envVar := range os.Environ() {
				parts := strings.SplitN(envVar, "=", 2)
				if len(parts) != 2 {
					continue
				}

				key, val := parts[0], parts[1]
				if !strings.HasPrefix(key, providerEnvPrefix) {
					continue
				}

				suffix := key[len(providerEnvPrefix):]
				nameParts := strings.SplitN(suffix, "_", 2)
				if len(nameParts) != 2 || nameParts[0] == "" || nameParts[1] == "" {
					continue
				}

				providerName := strings.ToLower(nameParts[0])
				configKey := strings.ReplaceAll(strings.ToLower(nameParts[1]), "_", "-")

				if curCtx.Providers == nil {
					curCtx.Providers = make(map[string]map[string]string)
				}
				if curCtx.Providers[providerName] == nil {
					curCtx.Providers[providerName] = make(map[string]string)
				}
				curCtx.Providers[providerName][configKey] = val
			}

			return nil
		},
	}

	if l.ctxName != "" {
		overrides = append(overrides, func(cfg *config.Config) error {
			if !cfg.HasContext(l.ctxName) {
				return config.ContextNotFound(l.ctxName)
			}
			cfg.CurrentContext = l.ctxName
			return nil
		})
	}

	overrides = append(overrides, func(cfg *config.Config) error {
		if !cfg.HasContext(cfg.CurrentContext) {
			return config.ContextNotFound(cfg.CurrentContext)
		}
		return cfg.GetCurrentContext().Validate()
	})

	loaded, err := config.Load(ctx, source, overrides...)
	if err != nil {
		return config.NamespacedRESTConfig{}, err
	}

	if !loaded.HasContext(loaded.CurrentContext) {
		return config.NamespacedRESTConfig{}, fmt.Errorf("context %q not found", loaded.CurrentContext)
	}

	return loaded.GetCurrentContext().ToRESTConfig(ctx), nil
}

func (l *configLoader) configSource() config.Source {
	if l.configFile != "" {
		return config.ExplicitConfigFile(l.configFile)
	}
	return config.StandardLocation()
}
