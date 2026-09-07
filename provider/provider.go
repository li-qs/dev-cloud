package provider

import (
	"context"
	"devcloud/ent/resource"
	"fmt"
	"io"
	"time"
)

type Credential struct {
	Username string
	Password string
}

type ResourceSpec struct {
	Name       string
	Image      string
	Config     map[string]any
	Credential Credential
}

type ResourceInfo struct {
	RuntimeID string
	Name      string
	Status    string
	Image     string
	CreatedAt time.Time
}

type ResourceStats struct {
	CPUPercent  float64
	MemoryUsage uint64
	MemoryLimit uint64
	NetworkRx   uint64
	NetworkTx   uint64
}

type Provider interface {
	Create(ctx context.Context, resource ResourceSpec) (runtimeID string, err error)
	Start(ctx context.Context, runtimeID string) error
	Stop(ctx context.Context, runtimeID string) error
	Restart(ctx context.Context, runtimeID string) error
	Remove(ctx context.Context, runtimeID string) error
	Inspect(ctx context.Context, runtimeID string) (*ResourceInfo, error)
	Logs(ctx context.Context, runtimeID string) (io.ReadCloser, error)
	Stats(ctx context.Context, runtimeID string) (*ResourceStats, error)
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
