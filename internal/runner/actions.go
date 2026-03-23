package runner

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

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
