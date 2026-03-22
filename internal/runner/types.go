package runner

// Re-export types from the shared types package for convenience.
import t "github.com/murataslan1/ci-debugger/internal/types"

type (
	StepStatus = t.StepStatus
	StepResult = t.StepResult
	JobStatus  = t.JobStatus
	JobResult  = t.JobResult
	RunResult  = t.RunResult
	JobContext = t.JobContext
	RunConfig  = t.RunConfig
)

const (
	StepStatusPending = t.StepStatusPending
	StepStatusRunning = t.StepStatusRunning
	StepStatusPassed  = t.StepStatusPassed
	StepStatusFailed  = t.StepStatusFailed
	StepStatusSkipped = t.StepStatusSkipped

	JobStatusPending = t.JobStatusPending
	JobStatusRunning = t.JobStatusRunning
	JobStatusPassed  = t.JobStatusPassed
	JobStatusFailed  = t.JobStatusFailed
	JobStatusSkipped = t.JobStatusSkipped
)
