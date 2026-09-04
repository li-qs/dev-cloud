package provider

import (
	"context"
	"devcloud/ent/resource"
	"fmt"
)

type ResourceSpec struct {
	Name       string
	Type       string
	Config     map[string]any
	Credential string
}

type Provider interface {
	Create(ctx context.Context, resource *ResourceSpec) (runtimeID string, err error)
	Start(ctx context.Context, runtimeID string) error
	Stop(ctx context.Context, runtimeID string) error
	Restart(ctx context.Context, runtimeID string) error
	Delete(ctx context.Context, runtimeID string) error
}

type Registry struct {
	providers map[resource.Provider]Provider
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[resource.Provider]Provider),
	}
}

func (r *Registry) Register(name resource.Provider, p Provider) {
	r.providers[name] = p
}

func (r *Registry) Get(name resource.Provider) (Provider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %q not found", name)
	}
	return p, nil
}
