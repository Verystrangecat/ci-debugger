package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
)

const (
	// ManagedLabel marks containers created by ci-debugger for cleanup.
	ManagedLabel = "ci-debugger=true"
)

// CreateContainer creates and returns a container ID.
// The container uses "tail -f /dev/null" as entrypoint to stay alive.
func (c *Client) CreateContainer(ctx context.Context, opts ContainerOpts) (string, error) {
	labels := map[string]string{"ci-debugger": "true"}
	for k, v := range opts.Labels {
		labels[k] = v
	}

	cfg := &container.Config{
		Image:      opts.Image,
		Entrypoint: []string{"tail", "-f", "/dev/null"},
		Env:        opts.Env,
		WorkingDir: opts.WorkDir,
		Labels:     labels,
	}

	hostCfg := &container.HostConfig{
		Binds: opts.Binds,
	}

	var netCfg *network.NetworkingConfig
	if opts.Network != "" {
		netCfg = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				opts.Network: {},
			},
		}
	}

	resp, err := c.cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, nil, opts.Name)
	if err != nil {
		return "", fmt.Errorf("creating container: %w", err)
	}

	return resp.ID, nil
}

// CreateServiceContainer creates a service sidecar container using the image's
// own CMD/ENTRYPOINT (not overridden), and registers the service name as a
// network alias so job steps can reach it by hostname.
func (c *Client) CreateServiceContainer(ctx context.Context, opts ContainerOpts) (string, error) {
	labels := map[string]string{"ci-debugger": "true"}
	for k, v := range opts.Labels {
		labels[k] = v
	}

	cfg := &container.Config{
		Image:  opts.Image,
		Env:    opts.Env,
		Labels: labels,
		// No Entrypoint override — let the image run its own process
	}

	hostCfg := &container.HostConfig{
		Binds: opts.Binds,
	}

	var netCfg *network.NetworkingConfig
	if opts.Network != "" {
		netCfg = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				opts.Network: {
					Aliases: []string{opts.Name},
				},
			},
		}
	}

	resp, err := c.cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, nil, opts.Name)
	if err != nil {
		return "", fmt.Errorf("creating service container %q: %w", opts.Name, err)
	}

	return resp.ID, nil
}

// StartContainer starts a container by ID.
func (c *Client) StartContainer(ctx context.Context, containerID string) error {
	err := c.cli.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("starting container: %w", err)
	}
	return nil
}

// StopAndRemove stops and removes a container.
func (c *Client) StopAndRemove(ctx context.Context, containerID string) error {
	// Try to stop gracefully first
	timeout := 5
	_ = c.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})

	err := c.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("removing container: %w", err)
	}
	return nil
}

// WriteScript writes a shell script to a path inside the container via tar archive.
func (c *Client) WriteScript(ctx context.Context, containerID, path, content string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	name := filepath.Base(path)
	if err := tw.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0755,
		Size: int64(len(content)),
	}); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		return err
	}
	tw.Close()

	dir := filepath.Dir(path)
	return c.cli.CopyToContainer(ctx, containerID, dir, &buf, container.CopyToContainerOptions{})
}

// ExecInContainer runs a command in a running container and captures output.
func (c *Client) ExecInContainer(ctx context.Context, containerID string, opts ExecOpts) (*ExecResult, error) {
	execCfg := container.ExecOptions{
		Cmd:          opts.Cmd,
		Env:          opts.Env,
		WorkingDir:   opts.WorkDir,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          opts.TTY,
	}

	execID, err := c.cli.ContainerExecCreate(ctx, containerID, execCfg)
	if err != nil {
		return nil, fmt.Errorf("creating exec: %w", err)
	}

	resp, err := c.cli.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{Tty: opts.TTY})
	if err != nil {
		return nil, fmt.Errorf("attaching exec: %w", err)
	}
	defer resp.Close()

	var stdout, stderr bytes.Buffer
	if opts.TTY {
		_, err = io.Copy(&stdout, resp.Reader)
	} else {
		_, err = stdcopy.StdCopy(&stdout, &stderr, resp.Reader)
	}
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("reading exec output: %w", err)
	}

	inspect, err := c.cli.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return nil, fmt.Errorf("inspecting exec: %w", err)
	}

	return &ExecResult{
		ExitCode: inspect.ExitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}

// ExecStreaming runs a command and streams output line-by-line via callbacks.
func (c *Client) ExecStreaming(ctx context.Context, containerID string, opts ExecOpts, outWriter, errWriter io.Writer) (int, error) {
	execCfg := container.ExecOptions{
		Cmd:          opts.Cmd,
		Env:          opts.Env,
		WorkingDir:   opts.WorkDir,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          opts.TTY,
	}

	execID, err := c.cli.ContainerExecCreate(ctx, containerID, execCfg)
	if err != nil {
		return -1, fmt.Errorf("creating exec: %w", err)
	}

	resp, err := c.cli.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{Tty: opts.TTY})
	if err != nil {
		return -1, fmt.Errorf("attaching exec: %w", err)
	}
	defer resp.Close()

	if outWriter == nil {
		outWriter = os.Stdout
	}
	if errWriter == nil {
		errWriter = os.Stderr
	}

	if opts.TTY {
		_, err = io.Copy(outWriter, resp.Reader)
	} else {
		_, err = stdcopy.StdCopy(outWriter, errWriter, resp.Reader)
	}
	if err != nil && err != io.EOF {
		return -1, fmt.Errorf("reading exec output: %w", err)
	}

	inspect, err := c.cli.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return -1, fmt.Errorf("inspecting exec: %w", err)
	}

	return inspect.ExitCode, nil
}

// OpenInteractiveShell starts an interactive bash session in the container.
func (c *Client) OpenInteractiveShell(ctx context.Context, containerID, workDir string) error {
	execCfg := container.ExecOptions{
		Cmd:          []string{"/bin/bash"},
		WorkingDir:   workDir,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}

	execID, err := c.cli.ContainerExecCreate(ctx, containerID, execCfg)
	if err != nil {
		// Try sh as fallback
		execCfg.Cmd = []string{"/bin/sh"}
		execID, err = c.cli.ContainerExecCreate(ctx, containerID, execCfg)
		if err != nil {
			return fmt.Errorf("creating shell exec: %w", err)
		}
	}

	resp, err := c.cli.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{Tty: true})
	if err != nil {
		return fmt.Errorf("attaching shell: %w", err)
	}
	defer resp.Close()

	// Set terminal to raw mode
	inFd := int(os.Stdin.Fd())
	oldState, err := makeRaw(inFd)
	if err == nil {
		defer restoreTerminal(inFd, oldState)
	}

	fmt.Fprintf(os.Stderr, "\n  \033[33m[ci-debugger] Dropped into container shell. Type 'exit' to return.\033[0m\n\n")

	// Copy stdin -> container
	done := make(chan struct{})
	go func() {
		io.Copy(resp.Conn, os.Stdin)
		close(done)
	}()

	// Copy container -> stdout
	io.Copy(os.Stdout, resp.Reader)
	<-done

	return nil
}

// mapEnv converts a map to KEY=VALUE slice.
func mapEnv(m map[string]string) []string {
	var result []string
	for k, v := range m {
		result = append(result, k+"="+v)
	}
	return result
}

// mergeEnvSlices merges multiple KEY=VALUE slices, later values override.
func mergeEnvSlices(slices ...[]string) []string {
	m := map[string]string{}
	for _, slice := range slices {
		for _, kv := range slice {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				m[parts[0]] = parts[1]
			}
		}
	}
	return mapEnv(m)
}
