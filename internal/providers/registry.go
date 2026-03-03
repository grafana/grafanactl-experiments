package providers

// All returns all compile-time registered providers.
// When new providers are implemented, they should be added to the
// returned slice.
func All() []Provider {
	return []Provider{}
}
