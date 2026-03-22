package runner

import (
	"fmt"
	"strings"
)

// WrapCommand wraps a "run:" script for execution with the given shell.
// Returns the shell command and the script content to write to a temp file.
func WrapCommand(run, shell, scriptPath string) ([]string, string) {
	if shell == "" {
		shell = "bash"
	}

	switch shell {
	case "bash":
		return []string{"bash", "--noprofile", "--norc", "-eo", "pipefail", scriptPath}, run
	case "sh":
		return []string{"sh", "-e", scriptPath}, run
	case "python", "python3":
		return []string{"python3", scriptPath}, run
	case "pwsh", "powershell":
		return []string{"pwsh", "-command", scriptPath}, run
	case "cmd":
		// Batch files on Windows -- not supported
		return []string{"cmd", "/D", "/E:ON", "/V:OFF", "/S", "/C", scriptPath}, run
	default:
		// Custom shell specified directly
		if strings.Contains(shell, "{0}") {
			// Replace {0} with script path
			parts := strings.Fields(strings.ReplaceAll(shell, "{0}", scriptPath))
			return parts, run
		}
		return []string{shell, scriptPath}, run
	}
}

// ScriptPath returns the path to use for step N's script inside the container.
func ScriptPath(stepIndex int) string {
	return fmt.Sprintf("/tmp/__ci_debugger_step_%d.sh", stepIndex)
}
