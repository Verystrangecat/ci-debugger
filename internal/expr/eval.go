package expr

import (
	"regexp"
	"strings"
)

// Context holds all namespaced values available for expression expansion.
type Context struct {
	Env     map[string]string
	Secrets map[string]string
	Matrix  map[string]string
	Inputs  map[string]string
	Github  map[string]string            // github.sha, github.ref, github.repository, …
	Needs   map[string]map[string]string // needs.JOB.outputs.KEY
	Steps   map[string]map[string]string // steps.ID.outputs.KEY
	Job     map[string]string            // job.status
}

var exprRe = regexp.MustCompile(`\$\{\{(.+?)\}\}`)

// Expand replaces all ${{ … }} expressions in s using ctx.
// Unknown expressions resolve to "".
func Expand(s string, ctx *Context) string {
	return exprRe.ReplaceAllStringFunc(s, func(match string) string {
		inner := strings.TrimSpace(match[3 : len(match)-2])
		return resolveExpr(inner, ctx)
	})
}

// resolveExpr resolves a single bare expression (no ${{ }}) against ctx.
func resolveExpr(e string, ctx *Context) string {
	dot := strings.Index(e, ".")
	if dot < 0 {
		return ""
	}
	ns, rest := e[:dot], e[dot+1:]

	switch ns {
	case "env":
		if ctx.Env != nil {
			return ctx.Env[rest]
		}
	case "secrets":
		if ctx.Secrets != nil {
			return ctx.Secrets[rest]
		}
	case "matrix":
		if ctx.Matrix != nil {
			return ctx.Matrix[rest]
		}
	case "inputs":
		if ctx.Inputs != nil {
			return ctx.Inputs[rest]
		}
	case "github":
		if ctx.Github != nil {
			return ctx.Github[rest]
		}
	case "job":
		if ctx.Job != nil {
			return ctx.Job[rest]
		}
	case "needs":
		// needs.JOB.outputs.KEY
		if ctx.Needs != nil {
			if sub := strings.SplitN(rest, ".outputs.", 2); len(sub) == 2 {
				if outs, ok := ctx.Needs[sub[0]]; ok {
					return outs[sub[1]]
				}
			}
		}
	case "steps":
		// steps.ID.outputs.KEY
		if ctx.Steps != nil {
			if sub := strings.SplitN(rest, ".outputs.", 2); len(sub) == 2 {
				if outs, ok := ctx.Steps[sub[0]]; ok {
					return outs[sub[1]]
				}
			}
		}
	}
	return ""
}

// EvalBool evaluates an if: condition string against ctx.
// Strips the ${{ }} wrapper if present, then delegates to evalExpr.
func EvalBool(condition string, ctx *Context) bool {
	cond := strings.TrimSpace(condition)
	if strings.HasPrefix(cond, "${{") && strings.HasSuffix(cond, "}}") {
		cond = strings.TrimSpace(cond[3 : len(cond)-2])
	}
	return evalBoolExpr(cond, ctx)
}

func evalBoolExpr(cond string, ctx *Context) bool {
	cond = strings.TrimSpace(cond)

	// || — lowest precedence
	if idx := findLogicalOp(cond, "||"); idx >= 0 {
		return evalBoolExpr(cond[:idx], ctx) || evalBoolExpr(cond[idx+2:], ctx)
	}

	// &&
	if idx := findLogicalOp(cond, "&&"); idx >= 0 {
		return evalBoolExpr(cond[:idx], ctx) && evalBoolExpr(cond[idx+2:], ctx)
	}

	// ! prefix
	if strings.HasPrefix(cond, "!") {
		return !evalBoolExpr(cond[1:], ctx)
	}

	// Strip outer parentheses
	if strings.HasPrefix(cond, "(") && strings.HasSuffix(cond, ")") {
		return evalBoolExpr(cond[1:len(cond)-1], ctx)
	}

	// Bare functions / keywords
	switch cond {
	case "success()":
		if ctx.Job != nil {
			return ctx.Job["status"] == "" || ctx.Job["status"] == "success"
		}
		return true
	case "failure()":
		if ctx.Job != nil {
			return ctx.Job["status"] == "failure"
		}
		return false
	case "always()":
		return true
	case "cancelled()":
		return false
	case "true":
		return true
	case "false":
		return false
	}

	// String functions
	if strings.HasPrefix(cond, "contains(") && strings.HasSuffix(cond, ")") {
		a, b := splitTwoArgs(cond[9 : len(cond)-1])
		return strings.Contains(expandArg(a, ctx), expandArg(b, ctx))
	}
	if strings.HasPrefix(cond, "startsWith(") && strings.HasSuffix(cond, ")") {
		a, b := splitTwoArgs(cond[11 : len(cond)-1])
		return strings.HasPrefix(expandArg(a, ctx), expandArg(b, ctx))
	}
	if strings.HasPrefix(cond, "endsWith(") && strings.HasSuffix(cond, ")") {
		a, b := splitTwoArgs(cond[9 : len(cond)-1])
		return strings.HasSuffix(expandArg(a, ctx), expandArg(b, ctx))
	}

	// Comparison operators
	if idx := strings.Index(cond, " == "); idx >= 0 {
		left := Expand(strings.TrimSpace(cond[:idx]), ctx)
		right := Expand(strings.TrimSpace(cond[idx+4:]), ctx)
		return left == right
	}
	if idx := strings.Index(cond, " != "); idx >= 0 {
		left := Expand(strings.TrimSpace(cond[:idx]), ctx)
		right := Expand(strings.TrimSpace(cond[idx+4:]), ctx)
		return left != right
	}

	// Default: permissive (unknown expressions run)
	return true
}

// findLogicalOp finds the index of op (|| or &&) outside parentheses.
func findLogicalOp(s, op string) int {
	depth := 0
	for i := 0; i < len(s)-1; i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
		default:
			if depth == 0 && strings.HasPrefix(s[i:], op) {
				return i
			}
		}
	}
	return -1
}

// splitTwoArgs splits "arg1, arg2" respecting a single comma.
func splitTwoArgs(s string) (string, string) {
	idx := strings.Index(s, ",")
	if idx < 0 {
		return strings.TrimSpace(s), ""
	}
	return strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+1:])
}

// expandArg expands a function argument.
//   - Quoted literals ('foo' or "foo") → strip quotes.
//   - Already-wrapped expressions (${{ ... }}) → expand directly.
//   - Bare identifiers (github.ref) → wrap in ${{ }} then expand.
func expandArg(arg string, ctx *Context) string {
	if (strings.HasPrefix(arg, "'") && strings.HasSuffix(arg, "'")) ||
		(strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"")) {
		return arg[1 : len(arg)-1]
	}
	if strings.Contains(arg, "${{") {
		return Expand(arg, ctx)
	}
	return Expand("${{ "+arg+" }}", ctx)
}
