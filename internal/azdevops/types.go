package azdevops

import "gopkg.in/yaml.v3"

// Pipeline represents a parsed azure-pipelines.yml file.
type Pipeline struct {
	Name      string        `yaml:"name"`
	Trigger   interface{}   `yaml:"trigger"`
	Pool      *Pool         `yaml:"pool"`
	Variables VariablesField `yaml:"variables"`
	Jobs      []Job          `yaml:"jobs"`
	Stages    []Stage        `yaml:"stages"`
	Steps     []Step         `yaml:"steps"` // top-level steps (no explicit job)
}

// Pool represents an Azure DevOps agent pool (string name or vmImage object).
type Pool struct {
	VMImage string
	Name    string
}

func (p *Pool) UnmarshalYAML(node *yaml.Node) error {
	// pool can be a bare string (pool name) or an object {vmImage: ...}
	var name string
	if err := node.Decode(&name); err == nil {
		p.Name = name
		return nil
	}
	var obj struct {
		VMImage string `yaml:"vmImage"`
		Name    string `yaml:"name"`
	}
	if err := node.Decode(&obj); err != nil {
		return err
	}
	p.VMImage = obj.VMImage
	p.Name = obj.Name
	return nil
}

// Stage represents an Azure DevOps stage.
type Stage struct {
	Stage       string        `yaml:"stage"`
	DisplayName string        `yaml:"displayName"`
	DependsOn   StringOrSlice `yaml:"dependsOn"`
	Jobs        []Job         `yaml:"jobs"`
}

// Job represents an Azure DevOps job.
type Job struct {
	Job              string        `yaml:"job"`
	DisplayName      string        `yaml:"displayName"`
	Pool             *Pool         `yaml:"pool"`
	DependsOn        StringOrSlice `yaml:"dependsOn"`
	Variables        VariablesField `yaml:"variables"`
	Steps            []Step        `yaml:"steps"`
	Condition        string        `yaml:"condition"`
	TimeoutInMinutes int           `yaml:"timeoutInMinutes"`
}

// Step represents a single Azure DevOps step.
type Step struct {
	Script           string            `yaml:"script"`
	Bash             string            `yaml:"bash"`
	PowerShell       string            `yaml:"powershell"`
	Task             string            `yaml:"task"`
	Checkout         string            `yaml:"checkout"`
	DisplayName      string            `yaml:"displayName"`
	Name             string            `yaml:"name"`
	Env              map[string]string `yaml:"env"`
	Inputs           map[string]string `yaml:"inputs"`
	Condition        string            `yaml:"condition"`
	WorkingDirectory string            `yaml:"workingDirectory"`
	ContinueOnError  bool              `yaml:"continueOnError"`
	TimeoutInMinutes int               `yaml:"timeoutInMinutes"`
}

// StringOrSlice handles YAML fields that can be a string or []string.
type StringOrSlice []string

func (s *StringOrSlice) UnmarshalYAML(node *yaml.Node) error {
	var single string
	if err := node.Decode(&single); err == nil {
		*s = StringOrSlice{single}
		return nil
	}
	var multi []string
	if err := node.Decode(&multi); err != nil {
		return err
	}
	*s = StringOrSlice(multi)
	return nil
}

// VariablesField handles Azure DevOps variables as a map or a list of {name, value} objects.
type VariablesField map[string]string

func (v *VariablesField) UnmarshalYAML(node *yaml.Node) error {
	// Try simple map: {key: value}
	var m map[string]string
	if err := node.Decode(&m); err == nil {
		*v = VariablesField(m)
		return nil
	}
	// Try list: [{name: key, value: val}, ...]
	var list []struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	}
	if err := node.Decode(&list); err != nil {
		return err
	}
	result := make(VariablesField)
	for _, item := range list {
		if item.Name != "" {
			result[item.Name] = item.Value
		}
	}
	*v = result
	return nil
}
