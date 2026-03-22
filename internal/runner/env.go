package runner

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
