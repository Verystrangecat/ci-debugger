package workflow

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Workflow represents a parsed .github/workflows/*.yml file.
type Workflow struct {
	Name     string            `yaml:"name"`
	FileName string            `yaml:"-"`
	On       TriggerConfig     `yaml:"on"`
	Env      map[string]string `yaml:"env"`
	Jobs     map[string]*Job   `yaml:"jobs"`
}

// Job represents a single job in the workflow.
type Job struct {
	Name           string             `yaml:"name"`
	RunsOn         StringOrSlice      `yaml:"runs-on"`
	Needs          StringOrSlice      `yaml:"needs"`
	If             string             `yaml:"if"`
	Env            map[string]string  `yaml:"env"`
	Steps          []*Step            `yaml:"steps"`
	TimeoutMinutes int                `yaml:"timeout-minutes"`
	Outputs        map[string]string  `yaml:"outputs"`
	Container      *ContainerConfig   `yaml:"container"`
	Services       map[string]*Service `yaml:"services"`
}

// DisplayName returns a human-friendly name for the job.
func (j *Job) DisplayName(id string) string {
	if j.Name != "" {
		return j.Name
	}
	return id
}

// Step represents a single step within a job.
type Step struct {
	ID              string            `yaml:"id"`
	Name            string            `yaml:"name"`
	Uses            string            `yaml:"uses"`
	Run             string            `yaml:"run"`
	Shell           string            `yaml:"shell"`
	With            map[string]string `yaml:"with"`
	Env             map[string]string `yaml:"env"`
	If              string            `yaml:"if"`
	WorkingDir      string            `yaml:"working-directory"`
	ContinueOnError bool              `yaml:"continue-on-error"`
	TimeoutMinutes  int               `yaml:"timeout-minutes"`
}

// DisplayName returns a human-friendly name for the step.
func (s *Step) DisplayName(index int) string {
	if s.Name != "" {
		return s.Name
	}
	if s.Uses != "" {
		return s.Uses
	}
	if s.Run != "" {
		line := s.Run
		for i, c := range line {
			if c == '\n' {
				line = line[:i]
				break
			}
		}
		if len(line) > 50 {
			line = line[:50] + "..."
		}
		return "Run: " + line
	}
	return fmt.Sprintf("Step %d", index+1)
}

// TriggerConfig handles the flexible "on:" field in GitHub Actions.
type TriggerConfig struct {
	raw interface{}
}

func (t *TriggerConfig) UnmarshalYAML(node *yaml.Node) error {
	var v interface{}
	if err := node.Decode(&v); err != nil {
		return err
	}
	t.raw = v
	return nil
}

// String returns a string representation of the trigger.
func (t TriggerConfig) String() string {
	if t.raw == nil {
		return ""
	}
	switch v := t.raw.(type) {
	case string:
		return v
	case []interface{}:
		var parts []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// StringOrSlice handles YAML fields that can be either a string or []string.
type StringOrSlice []string

func (s *StringOrSlice) UnmarshalYAML(node *yaml.Node) error {
	// Try single string first
	var single string
	if err := node.Decode(&single); err == nil {
		*s = StringOrSlice{single}
		return nil
	}
	// Try slice
	var multi []string
	if err := node.Decode(&multi); err != nil {
		return err
	}
	*s = StringOrSlice(multi)
	return nil
}

// ContainerConfig represents a service container configuration.
type ContainerConfig struct {
	Image   string            `yaml:"image"`
	Env     map[string]string `yaml:"env"`
	Ports   []string          `yaml:"ports"`
	Options string            `yaml:"options"`
}

// Service represents a service defined in a job.
type Service struct {
	Image   string            `yaml:"image"`
	Env     map[string]string `yaml:"env"`
	Ports   []string          `yaml:"ports"`
	Options string            `yaml:"options"`
}
