package runner

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/murataslan1/ci-debugger/internal/debugger"
	"github.com/murataslan1/ci-debugger/internal/docker"
	"github.com/murataslan1/ci-debugger/internal/ui"
	"github.com/murataslan1/ci-debugger/internal/workflow"
)

// Runner executes GitHub Actions workflows locally.
type Runner struct {
	docker   *docker.Client
	wf       *workflow.Workflow
	cfg      *RunConfig
	secrets  map[string]string
	env      map[string]string
	debugger *debugger.Debugger
	renderer *ui.Renderer
}

// NewRunner creates a new Runner.
func NewRunner(wf *workflow.Workflow, cfg *RunConfig, dockerClient *docker.Client) *Runner {
	return &Runner{
		docker:   dockerClient,
		wf:       wf,
		renderer: ui.New(cfg.Verbose),
		cfg:    cfg,
	}
}

// SetSecrets sets the secrets map.
func (r *Runner) SetSecrets(secrets map[string]string) {
	r.secrets = secrets
}

// SetEnv sets extra environment variables.
func (r *Runner) SetEnv(env map[string]string) {
	r.env = env
}

// SetDebugger attaches a debugger to the runner.
func (r *Runner) SetDebugger(d *debugger.Debugger) {
	r.debugger = d
}

// Run executes the workflow and returns the result.
func (r *Runner) Run(ctx context.Context) (*RunResult, error) {
	start := time.Now()

	// Determine job execution order
	layers, err := workflow.ResolveJobOrder(r.wf.Jobs)
	if err != nil {
		return nil, fmt.Errorf("resolving job order: %w", err)
	}

	result := &RunResult{
		WorkflowName: r.wf.Name,
	}

	for _, layer := range layers {
		for _, jobID := range layer {
			// Apply job filter
			if r.cfg.JobFilter != "" && jobID != r.cfg.JobFilter {
				continue
			}

			job := r.wf.Jobs[jobID]
			jobResult, err := r.executeJob(ctx, jobID, job)
			if err != nil {
				return nil, err
			}
			result.JobResults = append(result.JobResults, jobResult)

			// Stop on job failure
			if jobResult.Status == JobStatusFailed {
				result.Duration = time.Since(start)
				return result, nil
			}
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

func (r *Runner) executeJob(ctx context.Context, jobID string, job *workflow.Job) (*JobResult, error) {
	start := time.Now()
	jobName := job.DisplayName(jobID)

	// Resolve Docker image
	image, err := workflow.ResolveImage(job.RunsOn, r.cfg.PlatformOverrides)
	if err != nil {
		return nil, fmt.Errorf("job %q: %w", jobID, err)
	}

	r.renderer.RenderJobStart(jobID, job, image)

	// Pull image if needed
	exists, _ := r.docker.ImageExists(ctx, image)
	if !exists {
		fmt.Printf("  Pulling image %s...\n", image)
		if err := r.docker.PullImage(ctx, image); err != nil {
			return nil, fmt.Errorf("pulling image: %w", err)
		}
	}

	// Determine workspace
	workspaceDir := r.cfg.WorkspaceDir
	if workspaceDir == "" {
		workspaceDir, _ = os.Getwd()
	}

	// Build environment
	githubEnv := BuildGitHubEnv(r.wf, job, workspaceDir)
	mergedEnv := MergeEnvMaps(githubEnv, r.wf.Env, job.Env, r.env)

	// Create container
	containerID, err := r.docker.CreateContainer(ctx, docker.ContainerOpts{
		Image:   image,
		WorkDir: "/github/workspace",
		Env:     EnvMapToSlice(mergedEnv),
		Binds:   []string{workspaceDir + ":/github/workspace"},
	})
	if err != nil {
		return nil, fmt.Errorf("creating container: %w", err)
	}
	defer r.docker.StopAndRemove(ctx, containerID)

	if err := r.docker.StartContainer(ctx, containerID); err != nil {
		return nil, fmt.Errorf("starting container: %w", err)
	}

	// Initialize GitHub special files inside container
	initCmd := "mkdir -p /github/home /tmp && touch /tmp/github_output /tmp/github_env /tmp/github_path"
	r.docker.ExecInContainer(ctx, containerID, docker.ExecOpts{ //nolint
		Cmd: []string{"sh", "-c", initCmd},
	})

	jobCtx := &JobContext{
		JobID:        jobID,
		Job:          job,
		ContainerID:  containerID,
		WorkspaceDir: workspaceDir,
		Env:          mergedEnv,
		Secrets:      r.secrets,
		StepOutputs:  map[string]map[string]string{},
	}

	// Attach debugger to container
	if r.debugger != nil {
		r.debugger.SetContainer(containerID, "/github/workspace")
	}

	result := &JobResult{
		JobID:   jobID,
		JobName: jobName,
		Status:  JobStatusRunning,
	}

	for i, step := range job.Steps {
		// Pre-step breakpoint
		if r.debugger != nil && r.debugger.IsEnabled() && r.debugger.ShouldBreakBefore(step, i) {
			action, err := r.debugger.HandleBreakpointBefore(step, i, jobCtx)
			if err != nil {
				return nil, err
			}
			switch action {
			case debugger.ActionQuit:
				result.Status = JobStatusFailed
				result.Duration = time.Since(start)
				return result, nil
			case debugger.ActionSkip:
				skipped := &StepResult{Step: step, Index: i, Status: StepStatusSkipped}
				result.StepResults = append(result.StepResults, skipped)
				jobCtx.StepResults = append(jobCtx.StepResults, skipped)
				continue
			case debugger.ActionShell:
				r.debugger.OpenShell(func(id, wd string) error { return r.docker.OpenInteractiveShell(ctx, id, wd) }) //nolint
				// After shell, continue with step
			case debugger.ActionInspect:
				r.debugger.InspectStep(step, i)
				// Re-show prompt — call HandleBreakpointBefore again would loop; just continue for now
			}
		}

		stepResult := r.executeStep(ctx, jobCtx, step, i)
		result.StepResults = append(result.StepResults, stepResult)
		jobCtx.StepResults = append(jobCtx.StepResults, stepResult)

		// Post-step breakpoint
		if r.debugger != nil && r.debugger.IsEnabled() && r.debugger.ShouldBreakAfter(step, stepResult) {
			action, err := r.debugger.HandleBreakpointAfter(step, stepResult)
			if err != nil {
				return nil, err
			}
			switch action {
			case debugger.ActionQuit:
				result.Status = JobStatusFailed
				result.Duration = time.Since(start)
				return result, nil
			case debugger.ActionRetry:
				// Re-execute step
				stepResult = r.executeStep(ctx, jobCtx, step, i)
				result.StepResults[len(result.StepResults)-1] = stepResult
				jobCtx.StepResults[len(jobCtx.StepResults)-1] = stepResult
			case debugger.ActionShell:
				r.debugger.OpenShell(func(id, wd string) error { return r.docker.OpenInteractiveShell(ctx, id, wd) }) //nolint
			case debugger.ActionInspect:
				r.debugger.InspectStep(step, i)
			}
		}

		// Capture step outputs
		if step.ID != "" && len(stepResult.Outputs) > 0 {
			jobCtx.StepOutputs[step.ID] = stepResult.Outputs
		}

		if stepResult.Status == StepStatusFailed && !step.ContinueOnError {
			result.Status = JobStatusFailed
			result.Duration = time.Since(start)
			return result, nil
		}
	}

	result.Status = JobStatusPassed
	result.Duration = time.Since(start)
	return result, nil
}

func (r *Runner) executeStep(ctx context.Context, jobCtx *JobContext, step *workflow.Step, index int) *StepResult {
	start := time.Now()

	result := &StepResult{
		Step:   step,
		Index:  index,
		Status: StepStatusRunning,
	}

	// Check if condition
	if step.If != "" {
		if !r.evaluateCondition(step.If, jobCtx) {
			result.Status = StepStatusSkipped
			result.Duration = time.Since(start)
			r.renderer.RenderStepComplete(result, len(jobCtx.Job.Steps))
			return result
		}
	}

	r.renderer.RenderStepStart(step, index, len(jobCtx.Job.Steps))

	// Handle "uses:" steps
	if step.Uses != "" {
		result = r.executeUsesStep(ctx, jobCtx, step, index, start)
		r.renderer.RenderStepComplete(result, len(jobCtx.Job.Steps))
		return result
	}

	// Handle "run:" steps
	if step.Run == "" {
		result.Status = StepStatusSkipped
		result.Duration = time.Since(start)
		r.renderer.RenderStepComplete(result, len(jobCtx.Job.Steps))
		return result
	}

	// Merge step env
	stepEnv := MergeEnvMaps(jobCtx.Env, step.Env)

	// Expand expressions in the run script
	runScript := r.expandExpressions(step.Run, jobCtx)

	// Write script to container
	scriptPath := ScriptPath(index)
	if err := r.docker.WriteScript(ctx, jobCtx.ContainerID, scriptPath, runScript); err != nil {
		result.Status = StepStatusFailed
		result.Stderr = fmt.Sprintf("failed to write script: %v", err)
		result.Duration = time.Since(start)
		r.renderer.RenderStepComplete(result, len(jobCtx.Job.Steps))
		return result
	}

	// Determine working directory
	workDir := "/github/workspace"
	if step.WorkingDir != "" {
		if strings.HasPrefix(step.WorkingDir, "/") {
			workDir = step.WorkingDir
		} else {
			workDir = "/github/workspace/" + step.WorkingDir
		}
	}

	// Build exec command
	cmd, _ := WrapCommand(runScript, step.Shell, scriptPath)

	// Execute with streaming output
	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutWriter := io.MultiWriter(&stdoutBuf, &prefixWriter{prefix: "    ", out: os.Stdout, verbose: r.cfg.Verbose})
	stderrWriter := io.MultiWriter(&stderrBuf, &prefixWriter{prefix: "    \033[31m", suffix: "\033[0m", out: os.Stderr, verbose: r.cfg.Verbose})

	exitCode, err := r.docker.ExecStreaming(ctx, jobCtx.ContainerID, docker.ExecOpts{
		Cmd:     cmd,
		Env:     EnvMapToSlice(stepEnv),
		WorkDir: workDir,
	}, stdoutWriter, stderrWriter)

	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()

	if err != nil {
		result.Status = StepStatusFailed
		result.Stderr += "\n" + err.Error()
		result.Duration = time.Since(start)
		r.renderer.RenderStepComplete(result, len(jobCtx.Job.Steps))
		return result
	}

	result.ExitCode = exitCode
	if exitCode == 0 {
		result.Status = StepStatusPassed
	} else {
		result.Status = StepStatusFailed
	}

	// Capture GITHUB_OUTPUT
	result.Outputs = r.captureOutputs(ctx, jobCtx.ContainerID)

	// Apply GITHUB_ENV for subsequent steps
	r.applyGitHubEnv(ctx, jobCtx)

	result.Duration = time.Since(start)
	r.renderer.RenderStepComplete(result, len(jobCtx.Job.Steps))
	return result
}

func (r *Runner) executeUsesStep(ctx context.Context, jobCtx *JobContext, step *workflow.Step, index int, start time.Time) *StepResult {
	result := &StepResult{
		Step:  step,
		Index: index,
	}

	// Special-case: actions/checkout — workspace is already mounted, treat as no-op
	if strings.HasPrefix(step.Uses, "actions/checkout") {
		fmt.Fprintf(os.Stdout, "    (actions/checkout: workspace already mounted at /github/workspace)\n")
		result.Status = StepStatusPassed
		result.Duration = time.Since(start)
		return result
	}

	// Try to fetch the action definition from GitHub
	fmt.Fprintf(os.Stdout, "    ↓ fetching %s...\n", step.Uses)
	actionDef, err := fetchActionDef(step.Uses)
	if err != nil {
		fmt.Fprintf(os.Stdout, "    ⚠ could not fetch action: %v — skipping\n", err)
		result.Status = StepStatusSkipped
		result.Duration = time.Since(start)
		return result
	}

	switch actionDef.Runs.Using {
	case "composite":
		fmt.Fprintf(os.Stdout, "    ▶ running composite action (%d step(s))\n", len(actionDef.Runs.Steps))
		return r.executeCompositeAction(ctx, jobCtx, step, actionDef, index, start)
	default:
		fmt.Fprintf(os.Stdout, "    ⚠ action type %q not supported locally — skipping\n", actionDef.Runs.Using)
		result.Status = StepStatusSkipped
		result.Duration = time.Since(start)
		return result
	}
}

func (r *Runner) executeCompositeAction(ctx context.Context, jobCtx *JobContext, step *workflow.Step, actionDef *ActionDef, index int, start time.Time) *StepResult {
	result := &StepResult{
		Step:  step,
		Index: index,
	}

	// Build INPUT_* env vars and inputs map for expression expansion
	inputEnv := buildInputEnv(actionDef.Inputs, step.With)
	inputs := inputMap(actionDef.Inputs, step.With)

	compositeOutputs := map[string]string{}

	for i, actionStep := range actionDef.Runs.Steps {
		// Evaluate if: condition
		if actionStep.If != "" && !r.evaluateCondition(actionStep.If, jobCtx) {
			continue
		}

		if actionStep.Run == "" {
			// uses: inside composite — not yet supported, skip silently
			continue
		}

		// Merge env: job env + INPUT_* + action step env
		stepEnv := MergeEnvMaps(jobCtx.Env, inputEnv, actionStep.Env)

		// Expand expressions (including ${{ inputs.X }})
		runScript := r.expandExpressionsWithInputs(actionStep.Run, jobCtx, inputs)

		scriptPath := fmt.Sprintf("/tmp/__ci_debugger_step_%d_%d.sh", index, i)
		if err := r.docker.WriteScript(ctx, jobCtx.ContainerID, scriptPath, runScript); err != nil {
			result.Status = StepStatusFailed
			result.Stderr = fmt.Sprintf("composite action: failed to write script: %v", err)
			result.Duration = time.Since(start)
			return result
		}

		shell := actionStep.Shell
		if shell == "" {
			shell = "bash"
		}
		cmd, _ := WrapCommand(runScript, shell, scriptPath)

		workDir := "/github/workspace"
		if actionStep.WorkingDir != "" {
			if strings.HasPrefix(actionStep.WorkingDir, "/") {
				workDir = actionStep.WorkingDir
			} else {
				workDir = "/github/workspace/" + actionStep.WorkingDir
			}
		}

		var stdoutBuf, stderrBuf bytes.Buffer
		stdoutWriter := io.MultiWriter(&stdoutBuf, &prefixWriter{prefix: "    ", out: os.Stdout, verbose: r.cfg.Verbose})
		stderrWriter := io.MultiWriter(&stderrBuf, &prefixWriter{prefix: "    \033[31m", suffix: "\033[0m", out: os.Stderr, verbose: r.cfg.Verbose})

		exitCode, err := r.docker.ExecStreaming(ctx, jobCtx.ContainerID, docker.ExecOpts{
			Cmd:     cmd,
			Env:     EnvMapToSlice(stepEnv),
			WorkDir: workDir,
		}, stdoutWriter, stderrWriter)

		if (err != nil || exitCode != 0) && !actionStep.ContinueOnError {
			result.Status = StepStatusFailed
			result.ExitCode = exitCode
			result.Stdout = stdoutBuf.String()
			result.Stderr = stderrBuf.String()
			result.Duration = time.Since(start)
			return result
		}

		// Collect outputs from this composite sub-step
		for k, v := range r.captureOutputs(ctx, jobCtx.ContainerID) {
			compositeOutputs[k] = v
		}
		r.applyGitHubEnv(ctx, jobCtx)
	}

	result.Status = StepStatusPassed
	result.Outputs = compositeOutputs
	result.Duration = time.Since(start)
	return result
}

func (r *Runner) captureOutputs(ctx context.Context, containerID string) map[string]string {
	res, err := r.docker.ExecInContainer(ctx, containerID, docker.ExecOpts{
		Cmd: []string{"cat", "/tmp/github_output"},
	})
	if err != nil || res.ExitCode != 0 {
		return nil
	}

	outputs := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(res.Stdout))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			outputs[parts[0]] = parts[1]
		}
	}

	// Clear the output file after reading
	r.docker.ExecInContainer(ctx, containerID, docker.ExecOpts{ //nolint
		Cmd: []string{"sh", "-c", "> /tmp/github_output"},
	})

	return outputs
}

func (r *Runner) applyGitHubEnv(ctx context.Context, jobCtx *JobContext) {
	res, err := r.docker.ExecInContainer(ctx, jobCtx.ContainerID, docker.ExecOpts{
		Cmd: []string{"cat", "/tmp/github_env"},
	})
	if err != nil || res.ExitCode != 0 || res.Stdout == "" {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(res.Stdout))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			jobCtx.Env[parts[0]] = parts[1]
		}
	}

	// Clear it
	r.docker.ExecInContainer(ctx, jobCtx.ContainerID, docker.ExecOpts{ //nolint
		Cmd: []string{"sh", "-c", "> /tmp/github_env"},
	})
}

// evaluateCondition evaluates a simple "if:" expression.
func (r *Runner) evaluateCondition(condition string, jobCtx *JobContext) bool {
	cond := strings.TrimSpace(condition)
	// Strip ${{ }} wrapper if present
	cond = strings.TrimPrefix(cond, "${{")
	cond = strings.TrimSuffix(cond, "}}")
	cond = strings.TrimSpace(cond)

	switch cond {
	case "success()", "":
		// Check all previous steps passed
		for _, sr := range jobCtx.StepResults {
			if sr.Status == StepStatusFailed {
				return false
			}
		}
		return true
	case "failure()":
		for _, sr := range jobCtx.StepResults {
			if sr.Status == StepStatusFailed {
				return true
			}
		}
		return false
	case "always()":
		return true
	case "cancelled()":
		return false
	default:
		// Best-effort: just run it
		return true
	}
}

// expandExpressions does basic ${{ }} expansion.
func (r *Runner) expandExpressions(s string, jobCtx *JobContext) string {
	// Very simple: replace ${{ env.X }} and ${{ secrets.X }}
	result := s
	for k, v := range jobCtx.Env {
		result = strings.ReplaceAll(result, "${{ env."+k+" }}", v)
		result = strings.ReplaceAll(result, "${{env."+k+"}}", v)
	}
	for k, v := range jobCtx.Secrets {
		result = strings.ReplaceAll(result, "${{ secrets."+k+" }}", v)
		result = strings.ReplaceAll(result, "${{secrets."+k+"}}", v)
	}
	return result
}

// expandExpressionsWithInputs extends expandExpressions with ${{ inputs.X }}
// and ${{ steps.ID.outputs.KEY }} support for composite action steps.
func (r *Runner) expandExpressionsWithInputs(s string, jobCtx *JobContext, inputs map[string]string) string {
	result := r.expandExpressions(s, jobCtx)
	for k, v := range inputs {
		result = strings.ReplaceAll(result, "${{ inputs."+k+" }}", v)
		result = strings.ReplaceAll(result, "${{inputs."+k+"}}", v)
	}
	for stepID, outputs := range jobCtx.StepOutputs {
		for outKey, outVal := range outputs {
			result = strings.ReplaceAll(result, "${{ steps."+stepID+".outputs."+outKey+" }}", outVal)
		}
	}
	return result
}

// prefixWriter adds a prefix to each line written to it.
type prefixWriter struct {
	prefix  string
	suffix  string
	out     io.Writer
	verbose bool
	buf     strings.Builder
}

func (pw *prefixWriter) Write(p []byte) (n int, err error) {
	if !pw.verbose {
		return len(p), nil
	}
	pw.buf.Write(p)
	for {
		s := pw.buf.String()
		idx := strings.IndexByte(s, '\n')
		if idx < 0 {
			break
		}
		line := s[:idx]
		fmt.Fprintf(pw.out, "%s%s%s\n", pw.prefix, line, pw.suffix)
		pw.buf.Reset()
		pw.buf.WriteString(s[idx+1:])
	}
	return len(p), nil
}
