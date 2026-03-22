package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/image"
	dockerclient "github.com/docker/docker/client"
)

// Client wraps the Docker SDK client.
type Client struct {
	cli *dockerclient.Client
}

// NewClient creates a new Docker client using environment configuration.
func NewClient() (*Client, error) {
	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}
	return &Client{cli: cli}, nil
}

// Ping verifies the Docker daemon is reachable.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	if err != nil {
		return fmt.Errorf("docker daemon not reachable: %w\n\nMake sure Docker is running.", err)
	}
	return nil
}

// ImageExists checks if an image is available locally.
func (c *Client) ImageExists(ctx context.Context, imageName string) (bool, error) {
	images, err := c.cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				return true, nil
			}
		}
	}
	return false, nil
}

// PullImage pulls a Docker image, showing progress.
func (c *Client) PullImage(ctx context.Context, imageName string) error {
	reader, err := c.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pulling image %q: %w", imageName, err)
	}
	defer reader.Close()

	// Decode and display progress
	decoder := json.NewDecoder(reader)
	type progressDetail struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	}
	type pullEvent struct {
		Status         string         `json:"status"`
		ProgressDetail progressDetail `json:"progressDetail"`
		ID             string         `json:"id"`
	}

	layers := map[string]string{}
	for {
		var event pullEvent
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if event.ID != "" && event.Status != "" {
			layers[event.ID] = event.Status
		}
	}

	fmt.Fprintf(os.Stderr, "  Pulled %s\n", imageName)
	return nil
}

// Close closes the Docker client.
func (c *Client) Close() error {
	return c.cli.Close()
}

// Raw returns the underlying Docker client (for advanced operations).
func (c *Client) Raw() *dockerclient.Client {
	return c.cli
}
