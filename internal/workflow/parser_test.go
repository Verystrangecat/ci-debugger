package workflow_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/murataslan1/ci-debugger/internal/workflow"
)

func testdataPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filepath.Dir(filepath.Dir(file)))
	return filepath.Join(dir, "testdata", name)
}

func TestParseFile_Simple(t *testing.T) {
	wf, err := workflow.ParseFile(testdataPath("simple.yml"))
	require.NoError(t, err)

	assert.Equal(t, "Simple Test", wf.Name)
	assert.Contains(t, wf.Jobs, "test")
	assert.Len(t, wf.Jobs["test"].Steps, 3)
	assert.Equal(t, "Say hello", wf.Jobs["test"].Steps[0].Name)
}

func TestParseFile_MultiJob(t *testing.T) {
	wf, err := workflow.ParseFile(testdataPath("multi_job.yml"))
	require.NoError(t, err)

	assert.Equal(t, "from-workflow", wf.Env["GLOBAL_VAR"])
	assert.Contains(t, wf.Jobs, "build")
	assert.Contains(t, wf.Jobs, "test")
	assert.Equal(t, workflow.StringOrSlice{"build"}, wf.Jobs["test"].Needs)
}

func TestParseFile_MissingFile(t *testing.T) {
	_, err := workflow.ParseFile("/nonexistent/file.yml")
	assert.Error(t, err)
}

func TestResolveJobOrder_Linear(t *testing.T) {
	jobs := map[string]*workflow.Job{
		"build": {RunsOn: []string{"ubuntu-latest"}},
		"test":  {RunsOn: []string{"ubuntu-latest"}, Needs: []string{"build"}},
		"deploy": {RunsOn: []string{"ubuntu-latest"}, Needs: []string{"test"}},
	}

	layers, err := workflow.ResolveJobOrder(jobs)
	require.NoError(t, err)
	require.Len(t, layers, 3)
	assert.Equal(t, []string{"build"}, layers[0])
	assert.Equal(t, []string{"test"}, layers[1])
	assert.Equal(t, []string{"deploy"}, layers[2])
}

func TestResolveJobOrder_Parallel(t *testing.T) {
	jobs := map[string]*workflow.Job{
		"test-unit": {RunsOn: []string{"ubuntu-latest"}},
		"test-int":  {RunsOn: []string{"ubuntu-latest"}},
		"deploy":    {RunsOn: []string{"ubuntu-latest"}, Needs: []string{"test-unit", "test-int"}},
	}

	layers, err := workflow.ResolveJobOrder(jobs)
	require.NoError(t, err)
	require.Len(t, layers, 2)
	assert.Len(t, layers[0], 2)
	assert.Contains(t, layers[0], "test-unit")
	assert.Contains(t, layers[0], "test-int")
	assert.Equal(t, []string{"deploy"}, layers[1])
}

func TestResolveJobOrder_CircularDependency(t *testing.T) {
	jobs := map[string]*workflow.Job{
		"a": {Needs: []string{"b"}},
		"b": {Needs: []string{"a"}},
	}
	_, err := workflow.ResolveJobOrder(jobs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

func TestResolveJobOrder_MissingDependency(t *testing.T) {
	jobs := map[string]*workflow.Job{
		"a": {Needs: []string{"nonexistent"}},
	}
	_, err := workflow.ResolveJobOrder(jobs)
	assert.Error(t, err)
}

func TestStringOrSlice_Single(t *testing.T) {
	wf, err := workflow.ParseFile(testdataPath("simple.yml"))
	require.NoError(t, err)
	assert.Equal(t, workflow.StringOrSlice{"ubuntu-latest"}, wf.Jobs["test"].RunsOn)
}

func TestStepDisplayName(t *testing.T) {
	step := &workflow.Step{Name: "My Step"}
	assert.Equal(t, "My Step", step.DisplayName(0))

	step2 := &workflow.Step{Uses: "actions/checkout@v4"}
	assert.Equal(t, "actions/checkout@v4", step2.DisplayName(0))

	step3 := &workflow.Step{Run: "echo hello"}
	assert.Equal(t, "Run: echo hello", step3.DisplayName(0))

	step4 := &workflow.Step{}
	assert.Contains(t, step4.DisplayName(2), "Step 3")
}
