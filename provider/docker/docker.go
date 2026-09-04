package docker

import (
	"context"
	"devcloud/provider"
)

type Docker struct {
}

func NewProvider() *Docker {
	return &Docker{}
}

func (d *Docker) Create(ctx context.Context, resource *provider.ResourceSpec) (runtimeID string, err error) {
	return "", nil
}

func (d *Docker) Start(ctx context.Context, runtimeID string) error {
	return nil
}

func (d *Docker) Stop(ctx context.Context, runtimeID string) error {
	return nil
}

func (d *Docker) Restart(ctx context.Context, runtimeID string) error {
	return nil
}

func (d *Docker) Delete(ctx context.Context, runtimeID string) error {
	return nil
}
