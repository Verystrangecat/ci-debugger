package workflow

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ParseFile parses a single GitHub Actions workflow YAML file.
func ParseFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading workflow file: %w", err)
	}

	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("parsing workflow YAML: %w", err)
	}

	wf.FileName = path

	if wf.Jobs == nil || len(wf.Jobs) == 0 {
		return nil, fmt.Errorf("workflow %q has no jobs defined", path)
	}

	return &wf, nil
}

// DiscoverWorkflows finds all workflow YAML files in dir/.github/workflows/.
func DiscoverWorkflows(dir string) ([]*Workflow, error) {
	patterns := []string{
		filepath.Join(dir, ".github", "workflows", "*.yml"),
		filepath.Join(dir, ".github", "workflows", "*.yaml"),
	}

	var workflows []*Workflow
	seen := map[string]bool{}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			abs, _ := filepath.Abs(match)
			if seen[abs] {
				continue
			}
			seen[abs] = true

			wf, err := ParseFile(abs)
			if err != nil {
				// Skip invalid files but continue
				continue
			}
			workflows = append(workflows, wf)
		}
	}

	if len(workflows) == 0 {
		return nil, fmt.Errorf("no workflow files found in %s/.github/workflows/", dir)
	}

	return workflows, nil
}

// ResolveJobOrder performs a topological sort on job dependencies.
// Returns layers of job IDs that can be run in parallel within each layer.
func ResolveJobOrder(jobs map[string]*Job) ([][]string, error) {
	// Build adjacency and in-degree maps
	inDegree := make(map[string]int)
	dependents := make(map[string][]string) // job -> jobs that depend on it

	for id := range jobs {
		if _, exists := inDegree[id]; !exists {
			inDegree[id] = 0
		}
	}

	for id, job := range jobs {
		for _, dep := range job.Needs {
			if _, exists := jobs[dep]; !exists {
				return nil, fmt.Errorf("job %q needs %q which doesn't exist", id, dep)
			}
			inDegree[id]++
			dependents[dep] = append(dependents[dep], id)
		}
	}

	// Kahn's algorithm
	var layers [][]string
	var current []string

	for id, degree := range inDegree {
		if degree == 0 {
			current = append(current, id)
		}
	}

	processed := 0
	for len(current) > 0 {
		layers = append(layers, current)
		var next []string
		for _, id := range current {
			processed++
			for _, dep := range dependents[id] {
				inDegree[dep]--
				if inDegree[dep] == 0 {
					next = append(next, dep)
				}
			}
		}
		current = next
	}

	if processed != len(jobs) {
		return nil, fmt.Errorf("circular dependency detected in job needs")
	}

	return layers, nil
}
