package runner_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/murataslan1/ci-debugger/internal/runner"
)

func TestWrapCommand_Bash(t *testing.T) {
	cmd, script := runner.WrapCommand("echo hello", "bash", "/tmp/step.sh")
	assert.Equal(t, []string{"bash", "--noprofile", "--norc", "-eo", "pipefail", "/tmp/step.sh"}, cmd)
	assert.Equal(t, "echo hello", script)
}

func TestWrapCommand_Sh(t *testing.T) {
	cmd, _ := runner.WrapCommand("echo hello", "sh", "/tmp/step.sh")
	assert.Equal(t, []string{"sh", "-e", "/tmp/step.sh"}, cmd)
}

func TestWrapCommand_Python(t *testing.T) {
	cmd, _ := runner.WrapCommand("print('hi')", "python", "/tmp/step.sh")
	assert.Equal(t, []string{"python3", "/tmp/step.sh"}, cmd)
}

func TestWrapCommand_Default(t *testing.T) {
	// Empty shell defaults to bash
	cmd, _ := runner.WrapCommand("echo hi", "", "/tmp/step.sh")
	assert.Equal(t, []string{"bash", "--noprofile", "--norc", "-eo", "pipefail", "/tmp/step.sh"}, cmd)
}

func TestScriptPath(t *testing.T) {
	path := runner.ScriptPath(3)
	assert.Equal(t, "/tmp/__ci_debugger_step_3.sh", path)
}
