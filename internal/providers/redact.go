package providers

const redacted = "**REDACTED**"

// RedactSecrets redacts secret values in providerConfigs according to the
// ConfigKey metadata declared by each registered provider.
//
// Security model (secure by default):
//   - Known provider, Secret=true key  → REDACTED
//   - Known provider, Secret=false key → left as-is
//   - Known provider, undeclared key   → REDACTED
//   - Unknown provider (not registered) → ALL values REDACTED
//   - Empty values                      → left empty
//
// The map is mutated in place.
func RedactSecrets(providerConfigs map[string]map[string]string, registered []Provider) {
	// Build a set of non-secret keys per registered provider.
	nonSecretKeys := make(map[string]map[string]bool, len(registered))
	for _, p := range registered {
		safe := make(map[string]bool)
		for _, k := range p.ConfigKeys() {
			if !k.Secret {
				safe[k.Name] = true
			}
		}
		nonSecretKeys[p.Name()] = safe
	}

	for providerName, providerCfg := range providerConfigs {
		safeKeys, isRegistered := nonSecretKeys[providerName]
		for key, val := range providerCfg {
			if val == "" {
				continue
			}
			if !isRegistered || !safeKeys[key] {
				providerCfg[key] = redacted
			}
		}
	}
}
