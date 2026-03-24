package expr

import (
	"testing"
)

func TestExpand(t *testing.T) {
	ctx := &Context{
		Env:     map[string]string{"FOO": "bar", "GREETING": "hello"},
		Secrets: map[string]string{"TOKEN": "s3cr3t"},
		Matrix:  map[string]string{"os": "ubuntu"},
		Inputs:  map[string]string{"version": "1.2.3"},
		Github:  map[string]string{"sha": "abc123", "ref": "refs/heads/main"},
		Needs: map[string]map[string]string{
			"build": {"artifact": "myapp.tar.gz"},
		},
		Steps: map[string]map[string]string{
			"setup": {"path": "/usr/local/bin"},
		},
		Job: map[string]string{"status": "success"},
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"env", "${{ env.FOO }}", "bar"},
		{"env no space", "${{env.GREETING}}", "hello"},
		{"secret", "${{ secrets.TOKEN }}", "s3cr3t"},
		{"matrix", "${{ matrix.os }}", "ubuntu"},
		{"inputs", "${{ inputs.version }}", "1.2.3"},
		{"github sha", "${{ github.sha }}", "abc123"},
		{"github ref", "${{ github.ref }}", "refs/heads/main"},
		{"needs output", "${{ needs.build.outputs.artifact }}", "myapp.tar.gz"},
		{"steps output", "${{ steps.setup.outputs.path }}", "/usr/local/bin"},
		{"unknown", "${{ unknown.key }}", ""},
		{"mixed", "version=${{ inputs.version }} os=${{ matrix.os }}", "version=1.2.3 os=ubuntu"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Expand(tt.input, ctx)
			if got != tt.want {
				t.Errorf("Expand(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEvalBool_Functions(t *testing.T) {
	successCtx := &Context{Job: map[string]string{"status": "success"}}
	failCtx := &Context{Job: map[string]string{"status": "failure"}}

	tests := []struct {
		cond string
		ctx  *Context
		want bool
	}{
		{"success()", successCtx, true},
		{"success()", failCtx, false},
		{"failure()", successCtx, false},
		{"failure()", failCtx, true},
		{"always()", successCtx, true},
		{"always()", failCtx, true},
		{"cancelled()", successCtx, false},
		{"${{ success() }}", successCtx, true},
	}

	for _, tt := range tests {
		t.Run(tt.cond, func(t *testing.T) {
			got := EvalBool(tt.cond, tt.ctx)
			if got != tt.want {
				t.Errorf("EvalBool(%q) = %v, want %v", tt.cond, got, tt.want)
			}
		})
	}
}

func TestEvalBool_Comparisons(t *testing.T) {
	ctx := &Context{
		Github: map[string]string{"ref": "refs/heads/main"},
		Env:    map[string]string{"NODE_ENV": "production"},
	}

	tests := []struct {
		cond string
		want bool
	}{
		{"${{ github.ref }} == refs/heads/main", true},
		{"${{ github.ref }} != refs/heads/main", false},
		{"${{ env.NODE_ENV }} == production", true},
		{"${{ env.NODE_ENV }} == staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.cond, func(t *testing.T) {
			got := EvalBool(tt.cond, ctx)
			if got != tt.want {
				t.Errorf("EvalBool(%q) = %v, want %v", tt.cond, got, tt.want)
			}
		})
	}
}

func TestEvalBool_LogicalOps(t *testing.T) {
	successCtx := &Context{Job: map[string]string{"status": "success"}}

	tests := []struct {
		cond string
		want bool
	}{
		{"success() && always()", true},
		{"failure() && always()", false},
		{"failure() || success()", true},
		{"failure() || failure()", false},
		{"!failure()", true},
		{"!success()", false},
	}

	for _, tt := range tests {
		t.Run(tt.cond, func(t *testing.T) {
			got := EvalBool(tt.cond, successCtx)
			if got != tt.want {
				t.Errorf("EvalBool(%q) = %v, want %v", tt.cond, got, tt.want)
			}
		})
	}
}

func TestEvalBool_StringFunctions(t *testing.T) {
	ctx := &Context{
		Github: map[string]string{"ref": "refs/heads/feature-x"},
	}

	tests := []struct {
		cond string
		want bool
	}{
		{"contains(${{ github.ref }}, 'feature')", true},
		{"contains(${{ github.ref }}, 'main')", false},
		{"startsWith(${{ github.ref }}, 'refs/heads')", true},
		{"startsWith(${{ github.ref }}, 'refs/tags')", false},
		{"endsWith(${{ github.ref }}, 'feature-x')", true},
		{"endsWith(${{ github.ref }}, 'main')", false},
	}

	for _, tt := range tests {
		t.Run(tt.cond, func(t *testing.T) {
			got := EvalBool(tt.cond, ctx)
			if got != tt.want {
				t.Errorf("EvalBool(%q) = %v, want %v", tt.cond, got, tt.want)
			}
		})
	}
}
