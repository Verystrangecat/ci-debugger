// Package azdevops parses Azure DevOps pipeline YAML files and converts
// them to workflow.Workflow so the existing runner can execute them.
package azdevops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/murataslan1/ci-debugger/internal/workflow"
)

// AzureImageMap maps Azure DevOps vmImage labels to Docker images.
var AzureImageMap = map[string]string{
	"ubuntu-latest":  "ghcr.io/catthehacker/ubuntu:act-latest",
	"ubuntu-22.04":   "ghcr.io/catthehacker/ubuntu:act-22.04",
	"ubuntu-20.04":   "ghcr.io/catthehacker/ubuntu:act-20.04",
	"ubuntu-18.04":   "ghcr.io/catthehacker/ubuntu:act-18.04",
	// Windows/macOS not natively supported — fall back to ubuntu
	"windows-latest": "ghcr.io/catthehacker/ubuntu:act-latest",
	"windows-2022":   "ghcr.io/catthehacker/ubuntu:act-latest",
	"windows-2019":   "ghcr.io/catthehacker/ubuntu:act-latest",
	"macos-latest":   "ghcr.io/catthehacker/ubuntu:act-latest",
	"macos-13":       "ghcr.io/catthehacker/ubuntu:act-latest",
	"macos-12":       "ghcr.io/catthehacker/ubuntu:act-latest",
}

// IsPipelineFile returns true if path looks like an Azure DevOps pipeline file.
func IsPipelineFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return base == "azure-pipelines.yml" ||
		base == "azure-pipelines.yaml" ||
		strings.HasPrefix(base, "azure-pipelines-")
}

// DiscoverPipelines finds azure-pipelines.yml / azure-pipelines.yaml in dir.
func DiscoverPipelines(dir string) ([]*workflow.Workflow, error) {
	patterns := []string{
		filepath.Join(dir, "azure-pipelines.yml"),
		filepath.Join(dir, "azure-pipelines.yaml"),
	}

	var workflows []*workflow.Workflow
	for _, p := range patterns {
		if _, err := os.Stat(p); err == nil {
			wf, err := ParseFile(p)
			if err != nil {
				continue
			}
			workflows = append(workflows, wf)
		}
	}
	return workflows, nil
}

// ParseFile parses an Azure DevOps pipeline YAML file and converts it to
// a workflow.Workflow that the existing runner can execute unchanged.
func ParseFile(path string) (*workflow.Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading pipeline file: %w", err)
	}

	var pipeline Pipeline
	if err := yaml.Unmarshal(data, &pipeline); err != nil {
		return nil, fmt.Errorf("parsing pipeline YAML: %w", err)
	}

	return convert(&pipeline, path)
}

func convert(p *Pipeline, path string) (*workflow.Workflow, error) {
	name := p.Name
	if name == "" {
		name = filepath.Base(path)
	}

	wf := &workflow.Workflow{
		Name:     name,
		FileName: path,
		Env:      map[string]string(p.Variables),
		Jobs:     map[string]*workflow.Job{},
	}

	switch {
	case len(p.Stages) > 0:
		for _, stage := range p.Stages {
			for _, job := range stage.Jobs {
				id := makeID(stage.Stage, job.Job)
				wfJob := convertJob(job, p.Pool, p.Variables)
				// Propagate stage-level dependsOn as job needs
				for _, dep := range stage.DependsOn {
					wfJob.Needs = append(wfJob.Needs, dep)
				}
				wf.Jobs[id] = wfJob
			}
		}
	case len(p.Jobs) > 0:
		for i, job := range p.Jobs {
			id := job.Job
			if id == "" {
				id = sanitize(job.DisplayName)
			}
			if id == "" {
				id = fmt.Sprintf("job%d", i+1)
			}
			wf.Jobs[id] = convertJob(job, p.Pool, p.Variables)
		}
	case len(p.Steps) > 0:
		// Pipeline with no explicit jobs — wrap steps in a single job
		wf.Jobs["job"] = &workflow.Job{
			RunsOn: workflow.StringOrSlice{resolveImage(p.Pool)},
			Env:    map[string]string(p.Variables),
			Steps:  convertSteps(p.Steps),
		}
	default:
		return nil, fmt.Errorf("pipeline %q has no stages, jobs, or steps", path)
	}

	return wf, nil
}

func convertJob(job Job, pipelinePool *Pool, pipelineVars VariablesField) *workflow.Job {
	pool := job.Pool
	if pool == nil {
		pool = pipelinePool
	}

	// Merge pipeline-level and job-level variables
	merged := make(map[string]string)
	for k, v := range pipelineVars {
		merged[k] = v
	}
	for k, v := range job.Variables {
		merged[k] = v
	}

	return &workflow.Job{
		Name:   job.DisplayName,
		RunsOn: workflow.StringOrSlice{resolveImage(pool)},
		Needs:  workflow.StringOrSlice(job.DependsOn),
		If:     mapCondition(job.Condition),
		Env:    merged,
		Steps:  convertSteps(job.Steps),
	}
}

func convertSteps(steps []Step) []*workflow.Step {
	var out []*workflow.Step
	for i, step := range steps {
		out = append(out, convertStep(step, i))
	}
	return out
}

func convertStep(step Step, _ int) *workflow.Step {
	wfStep := &workflow.Step{
		ID:              step.Name,
		Name:            step.DisplayName,
		Env:             step.Env,
		If:              mapCondition(step.Condition),
		WorkingDir:      step.WorkingDirectory,
		ContinueOnError: step.ContinueOnError,
	}

	switch {
	case step.Checkout != "":
		// checkout: self is equivalent to actions/checkout
		wfStep.Uses = "actions/checkout@v3"
	case step.Script != "":
		wfStep.Run = step.Script
		wfStep.Shell = "bash"
	case step.Bash != "":
		wfStep.Run = step.Bash
		wfStep.Shell = "bash"
	case step.PowerShell != "":
		// PowerShell — run via bash as best effort
		wfStep.Run = step.PowerShell
		wfStep.Shell = "bash"
	case step.Task != "":
		wfStep.Uses = step.Task
		wfStep.With = step.Inputs
	}

	return wfStep
}

func resolveImage(pool *Pool) string {
	if pool == nil {
		return AzureImageMap["ubuntu-latest"]
	}
	if img, ok := AzureImageMap[pool.VMImage]; ok {
		return img
	}
	if pool.VMImage != "" {
		// Might already be a Docker image reference
		return pool.VMImage
	}
	return AzureImageMap["ubuntu-latest"]
}

// mapCondition converts Azure DevOps condition syntax to GitHub Actions if: syntax.
func mapCondition(condition string) string {
	switch strings.ToLower(strings.TrimSpace(condition)) {
	case "", "succeeded()":
		return ""
	case "succeededorfailed()":
		return "always()"
	case "failed()":
		return "failure()"
	case "always()":
		return "always()"
	case "canceled()":
		return "cancelled()"
	default:
		return ""
	}
}

func makeID(stage, job string) string {
	if stage != "" && job != "" {
		return sanitize(stage + "_" + job)
	}
	if job != "" {
		return sanitize(job)
	}
	return sanitize(stage)
}

func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			return r
		default:
			return '_'
		}
	}, s)
}
