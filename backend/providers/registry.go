package providers

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// MappingResolver resolves custom model aliases to provider and model
type MappingResolver interface {
	Resolve(ctx context.Context, alias string) *ResolvedMapping
}

// ResolvedMapping contains the resolved provider and model for an alias
type ResolvedMapping struct {
	ProviderID string
	ModelName  string
}

// Registry manages provider instances with thread-safe operations
type Registry struct {
	mu              sync.RWMutex
	providers       map[string]Provider
	mappingResolver MappingResolver
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// SetMappingResolver sets the custom mapping resolver
func (r *Registry) SetMappingResolver(resolver MappingResolver) {
	r.mappingResolver = resolver
}

// Register adds a provider to the registry
func (r *Registry) Register(id string, provider Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[id] = provider
}

// Get retrieves a provider by ID
func (r *Registry) Get(id string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[id]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", id)
	}

	return provider, nil
}

// List returns all registered providers
func (r *Registry) List() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		list = append(list, provider)
	}

	return list
}

// GetByModel retrieves a provider based on model name
// Returns: provider, resolvedModelName, error
// If custom mapping exists, resolvedModelName may differ from input model
func (r *Registry) GetByModel(model string) (Provider, string, error) {
	// 1. Check custom mapping first
	if r.mappingResolver != nil {
		if mapping := r.mappingResolver.Resolve(context.Background(), model); mapping != nil {
			provider, err := r.Get(mapping.ProviderID)
			if err != nil {
				return nil, "", err
			}
			return provider, mapping.ModelName, nil
		}
	}

	// 2. Fallback to prefix matching
	providerID := r.routeModel(model)
	if providerID == "" {
		return nil, "", fmt.Errorf("no provider found for model: %s", model)
	}

	provider, err := r.Get(providerID)
	if err != nil {
		return nil, "", err
	}
	return provider, model, nil
}

// routeModel maps model names to provider IDs based on prefix matching
func (r *Registry) routeModel(model string) string {
	// Normalize model to lowercase for matching
	modelLower := strings.ToLower(model)

	// Model routing logic
	switch {
	case strings.HasPrefix(modelLower, "gemini-"):
		return "antigravity"
	case strings.HasPrefix(modelLower, "claude-sonnet-"):
		return "antigravity"
	case strings.HasPrefix(modelLower, "claude-opus-"):
		return "antigravity"
	case strings.HasPrefix(modelLower, "claude-haiku-"):
		return "antigravity"
	case strings.HasPrefix(modelLower, "gpt-"):
		return "openai"
	case strings.HasPrefix(modelLower, "glm-"):
		return "glm"
	default:
		return ""
	}
}

// Remove removes a provider from the registry
func (r *Registry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, id)
}

// Update updates an existing provider in the registry
func (r *Registry) Update(id string, provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[id]; !exists {
		return fmt.Errorf("provider not found: %s", id)
	}

	r.providers[id] = provider
	return nil
}

// Exists checks if a provider exists in the registry
func (r *Registry) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.providers[id]
	return exists
}

// ListActive returns all registered providers (all are considered active)
func (r *Registry) ListActive() []Provider {
	return r.List()
}

// Count returns the total number of registered providers
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.providers)
}

// Clear removes all providers from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = make(map[string]Provider)
}
