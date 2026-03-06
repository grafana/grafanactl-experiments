package synth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/grafana/grafanactl/internal/config"
	"github.com/grafana/grafanactl/internal/providers"
	"github.com/grafana/grafanactl/internal/providers/synth/checks"
	"github.com/grafana/grafanactl/internal/providers/synth/probes"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() { //nolint:gochecknoinits // Self-registration pattern (like database/sql drivers).
	providers.Register(&SynthProvider{})
}

// SynthProvider manages Grafana Synthetic Monitoring resources.
type SynthProvider struct{}

// Name returns the unique identifier for this provider.
func (p *SynthProvider) Name() string { return "synth" }

// ShortDesc returns a one-line description of the provider.
func (p *SynthProvider) ShortDesc() string {
	return "Manage Grafana Synthetic Monitoring resources."
}

// Commands returns the Cobra commands contributed by this provider.
func (p *SynthProvider) Commands() []*cobra.Command {
	loader := &configLoader{}

	synthCmd := &cobra.Command{
		Use:   "synth",
		Short: p.ShortDesc(),
	}

	// Bind config flags on the parent — all subcommands inherit these.
	loader.bindFlags(synthCmd.PersistentFlags())

	synthCmd.AddCommand(checks.Commands(loader))
	synthCmd.AddCommand(probes.Commands(loader))

	return []*cobra.Command{synthCmd}
}

// Validate checks that the given provider configuration is valid.
func (p *SynthProvider) Validate(cfg map[string]string) error {
	if cfg["sm-url"] == "" {
		return errors.New("sm-url is required for the synth provider")
	}
	if cfg["sm-token"] == "" {
		return errors.New("sm-token is required for the synth provider")
	}
	return nil
}

// ConfigKeys returns the configuration keys used by this provider.
func (p *SynthProvider) ConfigKeys() []providers.ConfigKey {
	return []providers.ConfigKey{
		{Name: "sm-url", Secret: false},
		{Name: "sm-token", Secret: true},
	}
}

// configLoader loads SM credentials from the grafanactl config + env vars.
type configLoader struct {
	configFile string
	ctxName    string
}

func (l *configLoader) bindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&l.configFile, "config", "", "Path to the configuration file to use")
	flags.StringVar(&l.ctxName, "context", "", "Name of the context to use")
}

// LoadSMConfig loads the SM base URL, token, and K8s namespace from config.
// Priority (highest first):
//  1. GRAFANA_SM_URL / GRAFANA_SM_TOKEN env vars (explicit)
//  2. GRAFANA_PROVIDER_SYNTH_SM_URL / _TOKEN env vars (generic provider prefix)
//  3. Config file: providers.synth.sm-url / sm-token
func (l *configLoader) LoadSMConfig(ctx context.Context) (string, string, string, error) {
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

			// Resolve GRAFANA_PROVIDER_{NAME}_{KEY} environment variables.
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

	// Validate that the context exists (but don't require Grafana server config
	// since SM uses its own URL/token — only validate if grafana is configured).
	overrides = append(overrides, func(cfg *config.Config) error {
		if !cfg.HasContext(cfg.CurrentContext) {
			return config.ContextNotFound(cfg.CurrentContext)
		}
		return nil
	})

	loaded, err := config.Load(ctx, source, overrides...)
	if err != nil {
		return "", "", "", err
	}

	if !loaded.HasContext(loaded.CurrentContext) {
		return "", "", "", fmt.Errorf("context %q not found", loaded.CurrentContext)
	}

	curCtx := loaded.GetCurrentContext()

	// Extract SM credentials from providers config.
	var smURL, smToken string
	if prov := curCtx.Providers["synth"]; prov != nil {
		smURL = prov["sm-url"]
		smToken = prov["sm-token"]
	}

	// Explicit GRAFANA_SM_URL / GRAFANA_SM_TOKEN env vars override everything.
	if v := os.Getenv("GRAFANA_SM_URL"); v != "" {
		smURL = v
	}
	if v := os.Getenv("GRAFANA_SM_TOKEN"); v != "" {
		smToken = v
	}

	if smURL == "" {
		return "", "", "", errors.New(
			"SM URL not configured: set providers.synth.sm_url in config or GRAFANA_SM_URL env var")
	}
	if smToken == "" {
		return "", "", "", errors.New(
			"SM token not configured: set providers.synth.sm_token in config or GRAFANA_SM_TOKEN env var")
	}

	// Derive namespace from the Grafana config for K8s envelope metadata.
	// Falls back to "default" if no Grafana config is available.
	namespace := "default"
	if curCtx.Grafana != nil && !curCtx.Grafana.IsEmpty() {
		restCfg := curCtx.ToRESTConfig(ctx)
		namespace = restCfg.Namespace
	}

	return smURL, smToken, namespace, nil
}

func (l *configLoader) configSource() config.Source {
	if l.configFile != "" {
		return config.ExplicitConfigFile(l.configFile)
	}
	return config.StandardLocation()
}
