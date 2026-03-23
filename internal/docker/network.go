package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/network"
)

// CreateNetwork creates a bridge network and returns its ID.
func (c *Client) CreateNetwork(ctx context.Context, name string) (string, error) {
	resp, err := c.cli.NetworkCreate(ctx, name, network.CreateOptions{
		Driver: "bridge",
		Labels: map[string]string{"ci-debugger": "true"},
	})
	if err != nil {
		return "", fmt.Errorf("creating network %q: %w", name, err)
	}
	return resp.ID, nil
}

// RemoveNetwork removes a Docker network by ID.
func (c *Client) RemoveNetwork(ctx context.Context, networkID string) {
	_ = c.cli.NetworkRemove(ctx, networkID)
}
