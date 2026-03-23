package runner

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/murataslan1/ci-debugger/internal/workflow"
)

// LoadEnvFile parses a .env file into a map.
func LoadEnvFile(path string) (map[string]string, error) {
	return parseKVFile(path)
}

// LoadSecretFile parses a .secrets file into a map.
func LoadSecretFile(path string) (map[string]string, error) {
	return parseKVFile(path)
}

func parseKVFile(path string) (map[string]string, error) {
	result := map[string]string{}
	if path == "" {
		return result, nil
	}

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return result, nil
	}
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Remove optional "export " prefix
		line = strings.TrimPrefix(line, "export ")
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Remove surrounding quotes
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		result[key] = val
	}
	return result, scanner.Err()
}

// BuildGitHubEnv constructs the standard GitHub Actions environment variables.
func BuildGitHubEnv(wf *workflow.Workflow, job *workflow.Job, workspaceDir string) map[string]string {
	env := map[string]string{
		"GITHUB_WORKSPACE":  "/github/workspace",
		"GITHUB_OUTPUT":     "/tmp/github_output",
		"GITHUB_ENV":        "/tmp/github_env",
		"GITHUB_PATH":       "/tmp/github_path",
		"GITHUB_HOME":       "/github/home",
		"HOME":              "/github/home",
		"RUNNER_OS":         "Linux",
		"RUNNER_ARCH":       "X64",
		"RUNNER_TEMP":       "/tmp",
		"RUNNER_TOOL_CACHE": "/opt/hostedtoolcache",
		"CI":                "true",
		"GITHUB_ACTIONS":    "true",
	}

	// Try to get git info
	if sha := gitOutput("rev-parse", "HEAD"); sha != "" {
		env["GITHUB_SHA"] = sha
	}
	if ref := gitOutput("symbolic-ref", "HEAD"); ref != "" {
		env["GITHUB_REF"] = ref
		// Extract branch name
		branch := strings.TrimPrefix(ref, "refs/heads/")
		env["GITHUB_REF_NAME"] = branch
	}
	if remote := gitOutput("remote", "get-url", "origin"); remote != "" {
		// Convert git URL to owner/repo format
		env["GITHUB_SERVER_URL"] = "https://github.com"
		repo := extractRepo(remote)
		env["GITHUB_REPOSITORY"] = repo
	}

	if wf.Name != "" {
		env["GITHUB_WORKFLOW"] = wf.Name
	}

	return env
}

func gitOutput(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func extractRepo(remote string) string {
	// Handle both https://github.com/owner/repo.git and git@github.com:owner/repo.git
	remote = strings.TrimSuffix(remote, ".git")
	if idx := strings.Index(remote, "github.com/"); idx >= 0 {
		return remote[idx+len("github.com/"):]
	}
	if idx := strings.Index(remote, "github.com:"); idx >= 0 {
		return remote[idx+len("github.com:"):]
	}
	return remote
}

// envVarCategory classifies an environment variable for the report.
type envVarCategory struct {
	name     string
	status   string // "real", "stubbed", "unavailable"
	value    string
	note     string
}

// unavailableVars lists GitHub Actions vars that are injected by GitHub
// and cannot be replicated locally.
var unavailableVars = []envVarCategory{
	{name: "GITHUB_TOKEN", status: "unavailable", note: "injected by GitHub, use --secret-file to provide locally"},
	{name: "ACTIONS_ID_TOKEN_REQUEST_URL", status: "unavailable", note: "OIDC not available locally"},
	{name: "ACTIONS_ID_TOKEN_REQUEST_TOKEN", status: "unavailable", note: "OIDC not available locally"},
	{name: "GITHUB_RUN_ID", status: "unavailable", note: "assigned by GitHub at queue time"},
	{name: "GITHUB_RUN_NUMBER", status: "unavailable", note: "assigned by GitHub at queue time"},
	{name: "GITHUB_RUN_ATTEMPT", status: "unavailable", note: "assigned by GitHub at queue time"},
	{name: "GITHUB_ACTOR", status: "unavailable", note: "GitHub user who triggered the run"},
	{name: "GITHUB_TRIGGERING_ACTOR", status: "unavailable", note: "GitHub user who triggered the run"},
	{name: "GITHUB_EVENT_NAME", status: "unavailable", note: "trigger event (push, pull_request, etc.)"},
	{name: "GITHUB_EVENT_PATH", status: "unavailable", note: "path to event payload JSON"},
	{name: "GITHUB_JOB", status: "unavailable", note: "current job ID"},
	{name: "RUNNER_NAME", status: "unavailable", note: "GitHub-hosted runner name"},
	{name: "RUNNER_ENVIRONMENT", status: "unavailable", note: "github-hosted or self-hosted"},
}

// realVars are populated from the local git repo at runtime.
var realVarNames = map[string]bool{
	"GITHUB_SHA":        true,
	"GITHUB_REF":        true,
	"GITHUB_REF_NAME":   true,
	"GITHUB_REPOSITORY": true,
	"GITHUB_WORKFLOW":   true,
}

// PrintEnvReport prints a table showing which GitHub env vars are real,
// stubbed with local values, or unavailable locally.
func PrintEnvReport(wf *workflow.Workflow) {
	// Build the env map using a nil job (env report doesn't need job context)
	envMap := BuildGitHubEnv(wf, nil, "")

	const (
		colorReset  = "\033[0m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorRed    = "\033[31m"
		colorBold   = "\033[1m"
	)

	fmt.Printf("\n%sEnvironment Variables Report%s\n", colorBold, colorReset)
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("  %-42s %-12s %s\n", "Variable", "Status", "Value / Note")
	fmt.Println(strings.Repeat("─", 80))

	// Sort and print real/stubbed vars from the built env
	keys := make([]string, 0, len(envMap))
	for k := range envMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := envMap[k]
		display := v
		if len(display) > 30 {
			display = display[:27] + "..."
		}
		if realVarNames[k] {
			fmt.Printf("  %s%-42s%s %-12s %s\n", colorGreen, k, colorReset, "real", display)
		} else {
			fmt.Printf("  %s%-42s%s %-12s %s\n", colorYellow, k, colorReset, "stubbed", display)
		}
	}

	// Print unavailable vars
	fmt.Println()
	for _, v := range unavailableVars {
		fmt.Printf("  %s%-42s%s %-12s %s\n", colorRed, v.name, colorReset, "unavailable", v.note)
	}

	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("\n  %sreal%s      = derived from local git repo\n", colorGreen, colorReset)
	fmt.Printf("  %sstubbed%s   = hardcoded local value (may differ from GitHub runners)\n", colorYellow, colorReset)
	fmt.Printf("  %sunavailable%s = not present locally (injected by GitHub infrastructure)\n\n", colorRed, colorReset)
}

// MergeEnvMaps merges multiple maps, later values override earlier ones.
func MergeEnvMaps(maps ...map[string]string) map[string]string {
	result := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// EnvMapToSlice converts a map to KEY=VALUE slice.
func EnvMapToSlice(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, k+"="+v)
	}
	return result
}
