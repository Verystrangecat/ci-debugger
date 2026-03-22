package runner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/murataslan1/ci-debugger/internal/runner"
)

func TestLoadEnvFile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	err := os.WriteFile(envFile, []byte(`
# comment
KEY1=value1
KEY2="quoted value"
KEY3='single quoted'
export KEY4=exported
EMPTY=
`), 0644)
	require.NoError(t, err)

	env, err := runner.LoadEnvFile(envFile)
	require.NoError(t, err)

	assert.Equal(t, "value1", env["KEY1"])
	assert.Equal(t, "quoted value", env["KEY2"])
	assert.Equal(t, "single quoted", env["KEY3"])
	assert.Equal(t, "exported", env["KEY4"])
	assert.Equal(t, "", env["EMPTY"])
}

func TestLoadEnvFile_Missing(t *testing.T) {
	env, err := runner.LoadEnvFile("/nonexistent/.env")
	require.NoError(t, err)
	assert.Empty(t, env)
}

func TestLoadEnvFile_Empty(t *testing.T) {
	env, err := runner.LoadEnvFile("")
	require.NoError(t, err)
	assert.Empty(t, env)
}

func TestMergeEnvMaps(t *testing.T) {
	base := map[string]string{"A": "1", "B": "2"}
	override := map[string]string{"B": "overridden", "C": "3"}

	result := runner.MergeEnvMaps(base, override)
	assert.Equal(t, "1", result["A"])
	assert.Equal(t, "overridden", result["B"])
	assert.Equal(t, "3", result["C"])
}
