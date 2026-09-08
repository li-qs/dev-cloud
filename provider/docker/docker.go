package docker

import (
	"context"
	"devcloud/provider"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type Docker struct {
	client *client.Client
}

func NewProvider(
	host string,
	tlsEnabled bool,
	tlsCA string,
	tlsCert string,
	tlsKey string,
) (*Docker, error) {
	opts := []client.Opt{
		client.WithUserAgent("dev-cloud/0.1.0"), // TODO: appname version
		client.WithHost(host),
	}
	if tlsEnabled {
		opts = append(opts,
			client.WithTLSClientConfig(
				tlsCA,
				tlsCert,
				tlsKey,
			),
		)
	}

	c, err := client.New(opts...)
	if err != nil {
		return nil, err
	}

	return &Docker{
		client: c,
	}, nil
}

// Ping 检查 Docker daemon 是否可达，供 /ready 就绪探针使用。
func (d *Docker) Ping(ctx context.Context) error {
	_, err := d.client.Ping(ctx, client.PingOptions{})
	return err
}

func (d *Docker) Create(ctx context.Context, resource provider.ResourceSpec) (runtimeID string, err error) {
	res, err := d.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Name:  resource.Name,
		Image: resource.Image,
	})
	if err != nil {
		return "", err
	}
	return res.ID, nil
}

func (d *Docker) Start(ctx context.Context, runtimeID string) error {
	_, err := d.client.ContainerStart(ctx, runtimeID, client.ContainerStartOptions{})
	return err
}

func (d *Docker) Stop(ctx context.Context, runtimeID string) error {
	_, err := d.client.ContainerStop(ctx, runtimeID, client.ContainerStopOptions{})
	return err
}

func (d *Docker) Restart(ctx context.Context, runtimeID string) error {
	_, err := d.client.ContainerRestart(ctx, runtimeID, client.ContainerRestartOptions{})
	return err
}

func (d *Docker) Remove(ctx context.Context, runtimeID string) error {
	_, err := d.client.ContainerRemove(ctx, runtimeID, client.ContainerRemoveOptions{Force: true})
	return err
}

func (d *Docker) Inspect(ctx context.Context, runtimeID string) (*provider.ResourceInfo, error) {
	res, err := d.client.ContainerInspect(ctx, runtimeID, client.ContainerInspectOptions{})
	if err != nil {
		return nil, err
	}
	createdAt, err := time.Parse(time.RFC3339Nano, res.Container.Created)
	if err != nil {
		return nil, fmt.Errorf("parse container created time: %w", err)
	}
	return &provider.ResourceInfo{
		RuntimeID: res.Container.ID,
		Name:      res.Container.Name,
		Status:    string(res.Container.State.Status),
		Image:     res.Container.Image,
		CreatedAt: createdAt,
	}, nil
}

func (d *Docker) Logs(ctx context.Context, runtimeID string) (io.ReadCloser, error) {
	res, err := d.client.ContainerLogs(ctx, runtimeID, client.ContainerLogsOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Docker) Stats(ctx context.Context, runtimeID string) (*provider.ResourceStats, error) {
	res, err := d.client.ContainerStats(ctx, runtimeID, client.ContainerStatsOptions{})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var stats container.StatsResponse
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return &provider.ResourceStats{
		CPUPercent:  calculateCPUPercent(&stats),
		MemoryUsage: stats.MemoryStats.Usage,
		MemoryLimit: stats.MemoryStats.Limit,
		NetworkRx:   networkRx(&stats),
		NetworkTx:   networkTx(&stats),
	}, nil
}

func calculateCPUPercent(stats *container.StatsResponse) float64 {
	cpuDelta := float64(
		stats.CPUStats.CPUUsage.TotalUsage -
			stats.PreCPUStats.CPUUsage.TotalUsage,
	)

	systemDelta := float64(
		stats.CPUStats.SystemUsage -
			stats.PreCPUStats.SystemUsage,
	)

	if cpuDelta <= 0 || systemDelta <= 0 {
		return 0
	}

	onlineCPUs := stats.CPUStats.OnlineCPUs
	if onlineCPUs == 0 {
		onlineCPUs = uint32(len(stats.CPUStats.CPUUsage.PercpuUsage))
	}

	if onlineCPUs == 0 {
		return 0
	}

	return cpuDelta / systemDelta * float64(onlineCPUs) * 100
}

func networkRx(stats *container.StatsResponse) uint64 {
	var total uint64

	for _, network := range stats.Networks {
		total += network.RxBytes
	}

	return total
}

func networkTx(stats *container.StatsResponse) uint64 {
	var total uint64

	for _, network := range stats.Networks {
		total += network.TxBytes
	}

	return total
}
