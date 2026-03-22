package debugger

// BreakpointConfig defines where execution should pause.
type BreakpointConfig struct {
	BreakBefore  []string // step names/IDs to break before
	BreakAfter   []string // step names/IDs to break after
	BreakOnError bool     // break on any step failure
	StepMode     bool     // break before every step
}

// DebugAction is the user's response at a breakpoint.
type DebugAction int

const (
	ActionContinue DebugAction = iota // run this step and continue
	ActionSkip                        // skip this step
	ActionRetry                       // re-run previous step
	ActionShell                       // open interactive shell
	ActionInspect                     // show step details
	ActionQuit                        // abort the run
)
