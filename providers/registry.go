package providers

import (
	"fmt"
	"strings"
	"sync"

	"aigateway/models"
)

// Registry manages provider instances with thread-safe operations
type Registry struct {
	mu        sync.RWMutex
	providers map[string]*models.Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]*models.Provider),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(id string, provider *models.Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[id] = provider
}

// Get retrieves a provider by ID
func (r *Registry) Get(id string) (*models.Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[id]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", id)
	}

	if !provider.IsActive {
		return nil, fmt.Errorf("provider is inactive: %s", id)
	}

	return provider, nil
}

// List returns all registered providers
func (r *Registry) List() []*models.Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*models.Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		list = append(list, provider)
	}

	return list
}

// GetByModel retrieves a provider based on model prefix matching
func (r *Registry) GetByModel(model string) (*models.Provider, error) {
	providerID := r.routeModel(model)
	if providerID == "" {
		return nil, fmt.Errorf("no provider found for model: %s", model)
	}

	return r.Get(providerID)
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
func (r *Registry) Update(id string, provider *models.Provider) error {
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

// ListActive returns only active providers
func (r *Registry) ListActive() []*models.Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*models.Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		if provider.IsActive {
			list = append(list, provider)
		}
	}

	return list
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
	r.providers = make(map[string]*models.Provider)
}
