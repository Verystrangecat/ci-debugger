// Package types contains shared types used by runner and debugger packages.
package types

import (
	"time"

	"github.com/murataslan1/ci-debugger/internal/workflow"
)

// StepStatus represents the outcome of a step.
type StepStatus int

const (
	StepStatusPending StepStatus = iota
	StepStatusRunning
	StepStatusPassed
	StepStatusFailed
	StepStatusSkipped
)

func (s StepStatus) String() string {
	switch s {
	case StepStatusPending:
		return "pending"
	case StepStatusRunning:
		return "running"
	case StepStatusPassed:
		return "passed"
	case StepStatusFailed:
		return "failed"
	case StepStatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// StepResult captures the outcome of a single step execution.
type StepResult struct {
	Step     *workflow.Step
	Index    int
	Status   StepStatus
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	Outputs  map[string]string
}

// JobStatus represents the outcome of a job.
type JobStatus int

const (
	JobStatusPending JobStatus = iota
	JobStatusRunning
	JobStatusPassed
	JobStatusFailed
	JobStatusSkipped
)

func (s JobStatus) String() string {
	switch s {
	case JobStatusPending:
		return "pending"
	case JobStatusRunning:
		return "running"
	case JobStatusPassed:
		return "passed"
	case JobStatusFailed:
		return "failed"
	case JobStatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// JobResult captures the outcome of a job execution.
type JobResult struct {
	JobID       string
	JobName     string
	Status      JobStatus
	StepResults []*StepResult
	Duration    time.Duration
}

// RunResult captures the outcome of a full workflow run.
type RunResult struct {
	WorkflowName string
	JobResults   []*JobResult
	Duration     time.Duration
}

// JobContext holds runtime state for a job execution.
type JobContext struct {
	JobID        string
	Job          *workflow.Job
	ContainerID  string
	WorkspaceDir string
	StepResults  []*StepResult
	Env          map[string]string
	Secrets      map[string]string
	StepOutputs  map[string]map[string]string // stepID -> outputs
	Matrix       map[string]string            // matrix variable values for this run
}

// RunConfig holds options for a workflow run.
type RunConfig struct {
	WorkflowPath      string
	JobFilter         string
	EnvFile           string
	SecretFile        string
	PlatformOverrides map[string]string
	Verbose           bool
	WorkspaceDir      string

	// Debugger options
	StepMode     bool
	BreakBefore  []string
	BreakAfter   []string
	BreakOnError bool

	// Reporting
	EnvReport bool
}
