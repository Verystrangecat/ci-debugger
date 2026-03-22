package debugger

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/murataslan1/ci-debugger/internal/types"
	"github.com/murataslan1/ci-debugger/internal/workflow"
)

// Debugger manages breakpoints and step-by-step execution.
type Debugger struct {
	cfg         BreakpointConfig
	containerID string
	workDir     string
}

// New creates a new Debugger.
func New(cfg BreakpointConfig) *Debugger {
	return &Debugger{cfg: cfg}
}

// SetContainer sets the container ID for shell access.
func (d *Debugger) SetContainer(id, workDir string) {
	d.containerID = id
	d.workDir = workDir
}

// ShouldBreakBefore returns true if execution should pause before this step.
func (d *Debugger) ShouldBreakBefore(step *workflow.Step, index int) bool {
	if d.cfg.StepMode {
		return true
	}
	name := step.DisplayName(index)
	for _, b := range d.cfg.BreakBefore {
		if strings.EqualFold(b, name) || strings.EqualFold(b, step.ID) || strings.EqualFold(b, step.Name) {
			return true
		}
	}
	return false
}

// ShouldBreakAfter returns true if execution should pause after this step.
func (d *Debugger) ShouldBreakAfter(step *workflow.Step, result *types.StepResult) bool {
	if d.cfg.BreakOnError && result.Status == types.StepStatusFailed {
		return true
	}
	name := step.DisplayName(result.Index)
	for _, b := range d.cfg.BreakAfter {
		if strings.EqualFold(b, name) || strings.EqualFold(b, step.ID) || strings.EqualFold(b, step.Name) {
			return true
		}
	}
	return false
}

// HandleBreakpointBefore handles a pre-step breakpoint. Returns the action to take.
func (d *Debugger) HandleBreakpointBefore(step *workflow.Step, index int, jobCtx interface{}) (DebugAction, error) {
	name := step.DisplayName(index)

	fmt.Printf("\n\033[1;33m◆ BREAKPOINT\033[0m  before step \033[1m%s\033[0m\n", name)
	if step.Run != "" {
		// Show first few lines of the command
		lines := strings.Split(strings.TrimSpace(step.Run), "\n")
		fmt.Printf("  \033[90mCommand:\033[0m\n")
		max := 5
		if len(lines) < max {
			max = len(lines)
		}
		for _, l := range lines[:max] {
			fmt.Printf("    \033[90m%s\033[0m\n", l)
		}
		if len(lines) > 5 {
			fmt.Printf("    \033[90m... (%d more lines)\033[0m\n", len(lines)-5)
		}
	} else if step.Uses != "" {
		fmt.Printf("  \033[90mUses: %s\033[0m\n", step.Uses)
	}

	return d.prompt(false)
}

// HandleBreakpointAfter handles a post-step breakpoint. Returns the action to take.
func (d *Debugger) HandleBreakpointAfter(step *workflow.Step, result *types.StepResult) (DebugAction, error) {
	name := step.DisplayName(result.Index)

	statusLabel := "\033[32mpassed\033[0m"
	if result.Status == types.StepStatusFailed {
		statusLabel = fmt.Sprintf("\033[31mfailed (exit %d)\033[0m", result.ExitCode)
	}

	fmt.Printf("\n\033[1;33m◆ BREAKPOINT\033[0m  after step \033[1m%s\033[0m  %s\n", name, statusLabel)

	if result.Stderr != "" {
		lines := strings.Split(strings.TrimSpace(result.Stderr), "\n")
		fmt.Printf("  \033[90mLast stderr:\033[0m\n")
		start := 0
		if len(lines) > 5 {
			start = len(lines) - 5
		}
		for _, l := range lines[start:] {
			fmt.Printf("    \033[31m%s\033[0m\n", l)
		}
	}

	return d.prompt(true)
}

func (d *Debugger) prompt(afterStep bool) (DebugAction, error) {
	options := "  \033[1m[C]\033[0m Continue  \033[1m[S]\033[0m Skip"
	if !afterStep {
		// Only show skip before a step
	}
	if afterStep {
		options += "  \033[1m[R]\033[0m Retry"
	}
	if d.containerID != "" {
		options += "  \033[1m[D]\033[0m Shell"
	}
	options += "  \033[1m[I]\033[0m Inspect  \033[1m[Q]\033[0m Quit"

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("\n%s\n  → ", options)

		line, err := reader.ReadString('\n')
		if err != nil {
			return ActionQuit, err
		}

		choice := strings.TrimSpace(strings.ToLower(line))
		switch choice {
		case "c", "continue", "":
			return ActionContinue, nil
		case "s", "skip":
			return ActionSkip, nil
		case "r", "retry":
			if afterStep {
				return ActionRetry, nil
			}
			fmt.Println("  \033[33mRetry is only available after a step runs.\033[0m")
		case "d", "shell":
			if d.containerID != "" {
				return ActionShell, nil
			}
			fmt.Println("  \033[33mNo container available for shell access.\033[0m")
		case "i", "inspect":
			return ActionInspect, nil
		case "q", "quit":
			return ActionQuit, nil
		default:
			fmt.Printf("  \033[33mUnknown option %q. Try C, S, D, I, or Q.\033[0m\n", choice)
		}
	}
}

// InspectStep prints detailed information about a step.
func (d *Debugger) InspectStep(step *workflow.Step, index int) {
	fmt.Printf("\n\033[1mStep Inspection\033[0m\n")
	fmt.Printf("  Name:  %s\n", step.DisplayName(index))
	if step.ID != "" {
		fmt.Printf("  ID:    %s\n", step.ID)
	}
	if step.Uses != "" {
		fmt.Printf("  Uses:  %s\n", step.Uses)
	}
	if step.Run != "" {
		fmt.Printf("  Run:\n")
		for _, l := range strings.Split(step.Run, "\n") {
			fmt.Printf("    %s\n", l)
		}
	}
	if step.Shell != "" {
		fmt.Printf("  Shell: %s\n", step.Shell)
	}
	if step.WorkingDir != "" {
		fmt.Printf("  WorkDir: %s\n", step.WorkingDir)
	}
	if len(step.Env) > 0 {
		fmt.Printf("  Env:\n")
		for k, v := range step.Env {
			fmt.Printf("    %s=%s\n", k, v)
		}
	}
	if step.If != "" {
		fmt.Printf("  If: %s\n", step.If)
	}
	if step.ContinueOnError {
		fmt.Printf("  ContinueOnError: true\n")
	}
	fmt.Println()
}

// OpenShell drops into an interactive shell in the container.
// This requires the docker client to be passed in.
func (d *Debugger) OpenShell(openFn func(containerID, workDir string) error) error {
	if d.containerID == "" {
		return fmt.Errorf("no container available")
	}
	return openFn(d.containerID, d.workDir)
}

// IsEnabled returns true if any breakpoint config is set.
func (d *Debugger) IsEnabled() bool {
	return d.cfg.StepMode ||
		d.cfg.BreakOnError ||
		len(d.cfg.BreakBefore) > 0 ||
		len(d.cfg.BreakAfter) > 0
}
