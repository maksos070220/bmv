package registry

import (
	"fmt"

	"github.com/maksos070220/bmv/internal/browser"
)

type ProviderRegistry struct {
	providers map[browser.Browser]browser.Provider
}

func New(providers ...browser.Provider) *ProviderRegistry {
	registry := &ProviderRegistry{
		providers: make(map[browser.Browser]browser.Provider),
	}

	for _, provider := range providers {
		registry.providers[provider.Browser()] = provider
	}

	return registry
}

func (r *ProviderRegistry) Get(name browser.Browser) (browser.Provider, error) {
	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("unsupported browser: %s", name)
	}

	return provider, nil
}
