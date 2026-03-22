package debugger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/murataslan1/ci-debugger/internal/debugger"
	"github.com/murataslan1/ci-debugger/internal/types"
	"github.com/murataslan1/ci-debugger/internal/workflow"
)

func TestShouldBreakBefore_StepMode(t *testing.T) {
	d := debugger.New(debugger.BreakpointConfig{StepMode: true})
	step := &workflow.Step{Name: "Any Step"}
	assert.True(t, d.ShouldBreakBefore(step, 0))
}

func TestShouldBreakBefore_ByName(t *testing.T) {
	d := debugger.New(debugger.BreakpointConfig{BreakBefore: []string{"Run tests"}})

	step := &workflow.Step{Name: "Run tests"}
	assert.True(t, d.ShouldBreakBefore(step, 0))

	step2 := &workflow.Step{Name: "Build"}
	assert.False(t, d.ShouldBreakBefore(step2, 0))
}

func TestShouldBreakBefore_ByID(t *testing.T) {
	d := debugger.New(debugger.BreakpointConfig{BreakBefore: []string{"run-tests"}})

	step := &workflow.Step{ID: "run-tests", Name: "Run Tests"}
	assert.True(t, d.ShouldBreakBefore(step, 0))
}

func TestShouldBreakBefore_CaseInsensitive(t *testing.T) {
	d := debugger.New(debugger.BreakpointConfig{BreakBefore: []string{"RUN TESTS"}})

	step := &workflow.Step{Name: "run tests"}
	assert.True(t, d.ShouldBreakBefore(step, 0))
}

func TestShouldBreakBefore_NoMatch(t *testing.T) {
	d := debugger.New(debugger.BreakpointConfig{BreakBefore: []string{"deploy"}})
	step := &workflow.Step{Name: "Run tests"}
	assert.False(t, d.ShouldBreakBefore(step, 0))
}

func TestShouldBreakAfter_BreakOnError(t *testing.T) {
	d := debugger.New(debugger.BreakpointConfig{BreakOnError: true})
	step := &workflow.Step{Name: "Failing step"}

	failResult := &types.StepResult{Status: types.StepStatusFailed}
	assert.True(t, d.ShouldBreakAfter(step, failResult))

	passResult := &types.StepResult{Status: types.StepStatusPassed}
	assert.False(t, d.ShouldBreakAfter(step, passResult))
}

func TestShouldBreakAfter_ByName(t *testing.T) {
	d := debugger.New(debugger.BreakpointConfig{BreakAfter: []string{"Build"}})

	step := &workflow.Step{Name: "Build"}
	result := &types.StepResult{Status: types.StepStatusPassed}
	assert.True(t, d.ShouldBreakAfter(step, result))

	step2 := &workflow.Step{Name: "Test"}
	assert.False(t, d.ShouldBreakAfter(step2, result))
}

func TestIsEnabled(t *testing.T) {
	d1 := debugger.New(debugger.BreakpointConfig{})
	assert.False(t, d1.IsEnabled())

	d2 := debugger.New(debugger.BreakpointConfig{StepMode: true})
	assert.True(t, d2.IsEnabled())

	d3 := debugger.New(debugger.BreakpointConfig{BreakOnError: true})
	assert.True(t, d3.IsEnabled())

	d4 := debugger.New(debugger.BreakpointConfig{BreakBefore: []string{"build"}})
	assert.True(t, d4.IsEnabled())
}
