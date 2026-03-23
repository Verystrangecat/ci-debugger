package docker

// ContainerOpts configures a new container.
type ContainerOpts struct {
	Image   string
	Name    string
	Env     []string // KEY=VALUE pairs
	Binds   []string // host:container[:options]
	WorkDir string
	Labels  map[string]string
	Network string // Docker network ID or name to attach to
}

// ExecOpts configures an exec in a running container.
type ExecOpts struct {
	Cmd     []string
	Env     []string // KEY=VALUE pairs
	WorkDir string
	TTY     bool
	Stdin   bool
}

// ExecResult holds the output of a completed exec.
type ExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}
