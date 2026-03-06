// Package smcfg defines the shared config loader interface for the synth provider.
package smcfg

import "context"

// Loader can load SM credentials and the current namespace from config.
type Loader interface {
	LoadSMConfig(ctx context.Context) (baseURL, token, namespace string, err error)
}
