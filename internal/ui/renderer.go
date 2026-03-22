package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/murataslan1/ci-debugger/internal/types"
	"github.com/murataslan1/ci-debugger/internal/workflow"
)

// Renderer handles all terminal output formatting.
type Renderer struct {
	out     io.Writer
	verbose bool
}

// New creates a new Renderer.
func New(verbose bool) *Renderer {
	return &Renderer{out: os.Stdout, verbose: verbose}
}

// NewWithWriter creates a renderer with a custom writer.
func NewWithWriter(w io.Writer, verbose bool) *Renderer {
	return &Renderer{out: w, verbose: verbose}
}

// RenderWorkflowStart prints the workflow header.
func (r *Renderer) RenderWorkflowStart(wf *workflow.Workflow, trigger string) {
	name := wf.Name
	if name == "" {
		name = wf.FileName
	}

	title := StyleHeader.Render("ci-debugger") + "  " + StyleBold.Render(name)
	if trigger != "" {
		title += "  " + StyleDim.Render("on: "+trigger)
	}
	fmt.Fprintln(r.out, "\n"+title)
	fmt.Fprintln(r.out, StyleDim.Render(strings.Repeat("─", 55)))
}

// RenderJobStart prints the job header.
func (r *Renderer) RenderJobStart(jobID string, job *workflow.Job, image string) {
	name := job.DisplayName(jobID)
	line := "\n" + StyleJobName.Render("▶ "+name)
	if image != "" {
		line += "  " + StyleDim.Render("("+image+")")
	}
	fmt.Fprintln(r.out, line)
}

// RenderStepStart prints a step's running indicator.
func (r *Renderer) RenderStepStart(step *workflow.Step, index, total int) {
	name := step.DisplayName(index)
	progress := StyleDim.Render(fmt.Sprintf("[%d/%d]", index+1, total))
	indicator := StyleRunning.Render("⟳")
	fmt.Fprintf(r.out, "  %s %s %s\n", indicator, progress, name)
}

// RenderStepOutput writes a line of step output (respects verbose mode).
func (r *Renderer) RenderStepOutput(line string, isStderr bool) {
	if !r.verbose {
		return
	}
	if isStderr {
		fmt.Fprintf(r.out, "    %s\n", StyleFailed.Render(line))
	} else {
		fmt.Fprintf(r.out, "    %s\n", StyleDim.Render(line))
	}
}

// RenderStepComplete prints the step completion status.
func (r *Renderer) RenderStepComplete(result *types.StepResult, total int) {
	name := result.Step.DisplayName(result.Index)
	dur := StyleDim.Render("(" + result.Duration.Round(time.Millisecond).String() + ")")
	progress := StyleDim.Render(fmt.Sprintf("[%d/%d]", result.Index+1, total))

	switch result.Status {
	case types.StepStatusPassed:
		icon := StylePassed.Render("✓")
		fmt.Fprintf(r.out, "  %s %s %s  %s\n", icon, progress, name, dur)
	case types.StepStatusFailed:
		icon := StyleFailed.Render("✗")
		exitInfo := StyleDim.Render(fmt.Sprintf("(exit %d, %s)", result.ExitCode, result.Duration.Round(time.Millisecond)))
		fmt.Fprintf(r.out, "  %s %s %s  %s\n", icon, progress, name, exitInfo)
		// Show last stderr lines on failure (non-verbose mode)
		if !r.verbose && result.Stderr != "" {
			lines := strings.Split(strings.TrimSpace(result.Stderr), "\n")
			fmt.Fprintf(r.out, "    %s\n", StyleFailed.Render("── stderr ──"))
			start := 0
			if len(lines) > 10 {
				start = len(lines) - 10
				fmt.Fprintf(r.out, "    %s\n", StyleDim.Render(fmt.Sprintf("... %d lines omitted", start)))
			}
			for _, l := range lines[start:] {
				fmt.Fprintf(r.out, "    %s\n", StyleFailed.Render(l))
			}
		}
	case types.StepStatusSkipped:
		icon := StyleSkipped.Render("⊘")
		fmt.Fprintf(r.out, "  %s %s %s  %s\n", icon, progress, name, StyleSkipped.Render("(skipped)"))
	}
}

// RenderJobComplete prints the job completion status.
func (r *Renderer) RenderJobComplete(result *types.JobResult) {
	dur := StyleDim.Render("(" + result.Duration.Round(time.Millisecond).String() + ")")
	if result.Status == types.JobStatusPassed {
		fmt.Fprintf(r.out, "  %s  %s  %s\n", StylePassed.Render("✓ Job passed"), StyleBold.Render(result.JobName), dur)
	} else {
		fmt.Fprintf(r.out, "  %s  %s  %s\n", StyleFailed.Render("✗ Job failed"), StyleBold.Render(result.JobName), dur)
	}
}

// RenderSummary prints the final run summary.
func (r *Renderer) RenderSummary(result *types.RunResult, total time.Duration) {
	fmt.Fprintln(r.out, "\n"+StyleDim.Render(strings.Repeat("─", 55)))
	fmt.Fprintln(r.out, StyleHeader.Render("Summary"))

	// Count stats
	passed, failed, skipped := 0, 0, 0

	var rows []string
	rows = append(rows, fmt.Sprintf("  %-3s  %-35s  %-8s  %s", "#", "Step", "Status", "Duration"))
	rows = append(rows, "  "+strings.Repeat("─", 60))

	stepNum := 0
	for _, jr := range result.JobResults {
		// Job header row
		jobStatus := StylePassed.Render("passed")
		if jr.Status == types.JobStatusFailed {
			jobStatus = StyleFailed.Render("failed")
		}
		rows = append(rows, fmt.Sprintf("\n  %s  (%s)", StyleJobName.Render("Job: "+jr.JobName), jobStatus))

		for _, sr := range jr.StepResults {
			stepNum++
			name := sr.Step.DisplayName(sr.Index)
			if len(name) > 35 {
				name = name[:32] + "..."
			}

			var statusStr string
			switch sr.Status {
			case types.StepStatusPassed:
				passed++
				statusStr = StylePassed.Render("passed")
			case types.StepStatusFailed:
				failed++
				statusStr = StyleFailed.Render("FAILED")
			case types.StepStatusSkipped:
				skipped++
				statusStr = StyleSkipped.Render("skipped")
			}
			rows = append(rows, fmt.Sprintf("  %-3d  %-35s  %-8s  %s",
				stepNum, name, statusStr, StyleDim.Render(sr.Duration.Round(time.Millisecond).String())))
		}
	}

	// Print in a styled box
	content := strings.Join(rows, "\n")
	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorPurple).
		Padding(0, 1).
		Render(content)
	fmt.Fprintln(r.out, box)

	// Stats line
	statsLine := fmt.Sprintf("  Total: %s", StyleBold.Render(total.Round(time.Millisecond).String()))
	if passed > 0 {
		statsLine += "  " + StylePassed.Render(fmt.Sprintf("%d passed", passed))
	}
	if failed > 0 {
		statsLine += "  " + StyleFailed.Render(fmt.Sprintf("%d failed", failed))
	}
	if skipped > 0 {
		statsLine += "  " + StyleSkipped.Render(fmt.Sprintf("%d skipped", skipped))
	}
	fmt.Fprintln(r.out, statsLine)
}

// RenderError prints an error in a styled box.
func (r *Renderer) RenderError(err error) {
	msg := StyleErrorBox.Render("Error: " + err.Error())
	fmt.Fprintln(os.Stderr, msg)
}
