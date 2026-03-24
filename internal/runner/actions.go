package runner

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/murataslan1/ci-debugger/internal/docker"
	"github.com/murataslan1/ci-debugger/internal/workflow"
	"gopkg.in/yaml.v3"
)

// ActionDef represents a parsed action.yml / action.yaml file.
type ActionDef struct {
	Name        string                `yaml:"name"`
	Description string                `yaml:"description"`
	Inputs      map[string]ActionInput `yaml:"inputs"`
	Outputs     map[string]ActionOutput `yaml:"outputs"`
	Runs        ActionRuns            `yaml:"runs"`
}

// ActionInput describes a single action input.
type ActionInput struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default"`
}

// ActionOutput describes a single action output.
type ActionOutput struct {
	Description string `yaml:"description"`
	Value       string `yaml:"value"`
}

// ActionRuns describes how the action is executed.
type ActionRuns struct {
	Using string        `yaml:"using"` // "composite", "node20", "node16", "docker"
	Main  string        `yaml:"main"`  // node: entry point
	Image string        `yaml:"image"` // docker: container image
	Steps []*ActionStep `yaml:"steps"` // composite: inline steps
}

// ActionStep is a step inside a composite action.
type ActionStep struct {
	ID              string            `yaml:"id"`
	Name            string            `yaml:"name"`
	If              string            `yaml:"if"`
	Uses            string            `yaml:"uses"`
	Run             string            `yaml:"run"`
	Shell           string            `yaml:"shell"`
	With            map[string]string `yaml:"with"`
	Env             map[string]string `yaml:"env"`
	WorkingDir      string            `yaml:"working-directory"`
	ContinueOnError bool              `yaml:"continue-on-error"`
}

var (
	actionCacheMu sync.Mutex
	actionCache   = map[string]*ActionDef{}
	repoCache     sync.Map // uses string → local dir string
)

// fetchActionDef downloads and parses an action.yml from GitHub.
// uses is in the format "owner/repo@ref" or "owner/repo/path@ref".
// Returns an error if the action cannot be fetched or parsed.
func fetchActionDef(uses string) (*ActionDef, error) {
	actionCacheMu.Lock()
	if def, ok := actionCache[uses]; ok {
		actionCacheMu.Unlock()
		return def, nil
	}
	actionCacheMu.Unlock()

	// Parse ref: "owner/repo@ref" or "owner/repo/path@ref"
	ref := "main"
	actionPath := uses
	if idx := strings.LastIndex(uses, "@"); idx >= 0 {
		actionPath = uses[:idx]
		ref = uses[idx+1:]
	}

	// Try action.yml then action.yaml
	for _, filename := range []string{"action.yml", "action.yaml"} {
		url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", actionPath, ref, filename)
		def, err := downloadAndParse(url)
		if err != nil {
			continue
		}
		actionCacheMu.Lock()
		actionCache[uses] = def
		actionCacheMu.Unlock()
		return def, nil
	}

	return nil, fmt.Errorf("could not fetch action definition for %q (tried action.yml and action.yaml)", uses)
}

func downloadAndParse(url string) (*ActionDef, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var def ActionDef
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, err
	}

	return &def, nil
}

// buildInputEnv converts action inputs from step.With into INPUT_* env vars,
// applying defaults for any missing inputs.
func buildInputEnv(inputs map[string]ActionInput, with map[string]string) map[string]string {
	env := map[string]string{}

	// Start with defaults
	for k, input := range inputs {
		if input.Default != "" {
			key := "INPUT_" + strings.ToUpper(strings.ReplaceAll(k, "-", "_"))
			env[key] = input.Default
		}
	}

	// Override with provided values
	for k, v := range with {
		key := "INPUT_" + strings.ToUpper(strings.ReplaceAll(k, "-", "_"))
		env[key] = v
	}

	return env
}

// inputMap builds a name→value map from step.With and action defaults,
// used for ${{ inputs.X }} expression expansion.
func inputMap(inputs map[string]ActionInput, with map[string]string) map[string]string {
	m := map[string]string{}
	for k, input := range inputs {
		if input.Default != "" {
			m[k] = input.Default
		}
	}
	for k, v := range with {
		m[k] = v
	}
	return m
}

// downloadActionRepo downloads and extracts an action repo from GitHub.
// uses is in the format "owner/repo@ref" or "owner/repo/subdir@ref".
// Returns the path to the action root on the local filesystem.
func downloadActionRepo(uses string) (string, error) {
	if v, ok := repoCache.Load(uses); ok {
		return v.(string), nil
	}

	// Parse ref
	ref := "main"
	usesNoRef := uses
	if idx := strings.LastIndex(uses, "@"); idx >= 0 {
		usesNoRef = uses[:idx]
		ref = uses[idx+1:]
	}

	parts := strings.SplitN(usesNoRef, "/", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid uses format: %q", uses)
	}
	owner, repo := parts[0], parts[1]
	subdir := ""
	if len(parts) == 3 {
		subdir = parts[2]
	}

	localDir := filepath.Join(os.TempDir(), "__ci_debugger_actions", owner, repo+"@"+ref)

	// Return cached extraction if it exists
	if _, err := os.Stat(localDir); err == nil {
		result := localDir
		if subdir != "" {
			result = filepath.Join(localDir, subdir)
		}
		repoCache.Store(uses, result)
		return result, nil
	}

	// Download tarball via GitHub API (follows redirect automatically)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s", owner, repo, ref)
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("downloading %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading %s: HTTP %d", url, resp.StatusCode)
	}

	if err := os.MkdirAll(localDir, 0755); err != nil {
		return "", fmt.Errorf("creating dir %s: %w", localDir, err)
	}

	if err := extractTarGzStripped(resp.Body, localDir); err != nil {
		os.RemoveAll(localDir)
		return "", fmt.Errorf("extracting action: %w", err)
	}

	result := localDir
	if subdir != "" {
		result = filepath.Join(localDir, subdir)
	}
	repoCache.Store(uses, result)
	return result, nil
}

// extractTarGzStripped extracts a .tar.gz, stripping the leading path component
// (the top-level directory GitHub wraps tarballs in).
func extractTarGzStripped(r io.Reader, destDir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}

		// Strip the leading component (e.g. "owner-repo-sha123/")
		name := hdr.Name
		idx := strings.Index(name, "/")
		if idx < 0 {
			continue // root directory entry — skip
		}
		name = name[idx+1:]
		if name == "" {
			continue
		}

		target := filepath.Join(destDir, filepath.FromSlash(name))

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(f, tr)
			f.Close()
			if copyErr != nil {
				return copyErr
			}
		}
	}
	return nil
}

// executeNodeAction runs a node20/node16/node12 action inside the job container.
func (r *Runner) executeNodeAction(ctx context.Context, jobCtx *JobContext, step *workflow.Step, actionDef *ActionDef, uses string, index int, start time.Time) *StepResult {
	result := &StepResult{Step: step, Index: index}

	localDir, err := downloadActionRepo(uses)
	if err != nil {
		result.Status = StepStatusFailed
		result.Stderr = fmt.Sprintf("executeNodeAction: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Unique path inside the container for this action
	sanitized := strings.NewReplacer("/", "_", "@", "-").Replace(uses)
	containerPath := "/tmp/__ci_actions/" + sanitized + "/"

	// Ensure the destination exists
	r.docker.ExecInContainer(ctx, jobCtx.ContainerID, docker.ExecOpts{ //nolint
		Cmd: []string{"mkdir", "-p", containerPath},
	})

	if err := r.docker.CopyDirToContainer(ctx, jobCtx.ContainerID, localDir, containerPath); err != nil {
		result.Status = StepStatusFailed
		result.Stderr = fmt.Sprintf("executeNodeAction: copy to container: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	inputEnv := buildInputEnv(actionDef.Inputs, step.With)
	stepEnv := MergeEnvMaps(jobCtx.Env, inputEnv)

	cmd := []string{"node", containerPath + actionDef.Runs.Main}

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutWriter := io.MultiWriter(&stdoutBuf, &prefixWriter{prefix: "    ", out: os.Stdout, verbose: r.cfg.Verbose})
	stderrWriter := io.MultiWriter(&stderrBuf, &prefixWriter{prefix: "    \033[31m", suffix: "\033[0m", out: os.Stderr, verbose: r.cfg.Verbose})

	exitCode, execErr := r.docker.ExecStreaming(ctx, jobCtx.ContainerID, docker.ExecOpts{
		Cmd:     cmd,
		Env:     EnvMapToSlice(stepEnv),
		WorkDir: "/github/workspace",
	}, stdoutWriter, stderrWriter)

	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()
	result.ExitCode = exitCode

	if execErr != nil || exitCode != 0 {
		result.Status = StepStatusFailed
	} else {
		result.Status = StepStatusPassed
	}
	result.Duration = time.Since(start)
	return result
}

// executeDockAction runs a docker action by spinning up the action's image.
func (r *Runner) executeDockAction(ctx context.Context, jobCtx *JobContext, step *workflow.Step, actionDef *ActionDef, index int, start time.Time) *StepResult {
	result := &StepResult{Step: step, Index: index}

	if actionDef.Runs.Image == "Dockerfile" {
		fmt.Fprintf(os.Stdout, "    ⚠ docker action with Dockerfile not supported locally — skipping\n")
		result.Status = StepStatusSkipped
		result.Duration = time.Since(start)
		return result
	}

	image := strings.TrimPrefix(actionDef.Runs.Image, "docker://")

	exists, _ := r.docker.ImageExists(ctx, image)
	if !exists {
		fmt.Printf("    Pulling image %s...\n", image)
		if err := r.docker.PullImage(ctx, image); err != nil {
			result.Status = StepStatusFailed
			result.Stderr = fmt.Sprintf("pulling image: %v", err)
			result.Duration = time.Since(start)
			return result
		}
	}

	inputEnv := buildInputEnv(actionDef.Inputs, step.With)
	envSlice := EnvMapToSlice(MergeEnvMaps(jobCtx.Env, inputEnv))

	containerID, err := r.docker.CreateServiceContainer(ctx, docker.ContainerOpts{
		Image: image,
		Env:   envSlice,
		Binds: []string{jobCtx.WorkspaceDir + ":/github/workspace"},
	})
	if err != nil {
		result.Status = StepStatusFailed
		result.Stderr = fmt.Sprintf("creating docker action container: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer r.docker.StopAndRemove(ctx, containerID) //nolint

	if err := r.docker.StartContainer(ctx, containerID); err != nil {
		result.Status = StepStatusFailed
		result.Stderr = fmt.Sprintf("starting docker action container: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	exitCode, err := r.docker.WaitContainer(ctx, containerID)
	result.ExitCode = exitCode
	if err != nil || exitCode != 0 {
		result.Status = StepStatusFailed
		if err != nil {
			result.Stderr = err.Error()
		}
	} else {
		result.Status = StepStatusPassed
	}
	result.Duration = time.Since(start)
	return result
}
