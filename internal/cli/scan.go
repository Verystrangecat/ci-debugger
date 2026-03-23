package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/murataslan1/ci-debugger/internal/workflow"
)

type scanIssue struct {
	level    string // "error" or "warning"
	location string // e.g. "ci.yml > build > step 3"
	message  string
}

func newScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan .github/workflows/ for common issues",
		Example: `  ci-debugger scan
  ci-debugger scan --no-color`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			workflows, err := workflow.DiscoverWorkflows(cwd)
			if err != nil {
				return err
			}

			if len(workflows) == 0 {
				fmt.Println("No workflows found in .github/workflows/")
				return nil
			}

			var allIssues []scanIssue
			for _, wf := range workflows {
				issues := analyzeWorkflow(wf)
				allIssues = append(allIssues, issues...)
			}

			printScanResults(workflows, allIssues)

			// Exit non-zero if there are errors
			for _, issue := range allIssues {
				if issue.level == "error" {
					os.Exit(1)
				}
			}
			return nil
		},
	}
}

func analyzeWorkflow(wf *workflow.Workflow) []scanIssue {
	var issues []scanIssue
	file := wf.FileName

	// Check for circular dependencies
	if _, err := workflow.ResolveJobOrder(wf.Jobs); err != nil {
		issues = append(issues, scanIssue{
			level:    "error",
			location: file,
			message:  fmt.Sprintf("circular dependency in jobs: %v", err),
		})
	}

	for jobID, job := range wf.Jobs {
		loc := fmt.Sprintf("%s > %s", file, job.DisplayName(jobID))

		// Check for needs: referencing non-existent jobs
		for _, dep := range job.Needs {
			if _, ok := wf.Jobs[dep]; !ok {
				issues = append(issues, scanIssue{
					level:    "error",
					location: loc,
					message:  fmt.Sprintf("needs: references unknown job %q", dep),
				})
			}
		}

		// Check steps
		for i, step := range job.Steps {
			stepLoc := fmt.Sprintf("%s > step %d", loc, i+1)
			if step.Name != "" {
				stepLoc = fmt.Sprintf("%s > %q", loc, step.Name)
			}

			// Unsupported uses: actions (not actions/checkout)
			if step.Uses != "" && !strings.HasPrefix(step.Uses, "actions/checkout") {
				issues = append(issues, scanIssue{
					level:    "warning",
					location: stepLoc,
					message:  fmt.Sprintf("uses: %q is not supported locally — will be skipped", step.Uses),
				})
			}

			// Step has neither run nor uses
			if step.Run == "" && step.Uses == "" {
				issues = append(issues, scanIssue{
					level:    "error",
					location: stepLoc,
					message:  "step has no 'run' or 'uses' field",
				})
			}

			// Unclosed expression braces
			if step.Run != "" {
				if strings.Count(step.Run, "${{") != strings.Count(step.Run, "}}") {
					issues = append(issues, scanIssue{
						level:    "warning",
						location: stepLoc,
						message:  "possible unclosed expression: mismatched ${{ and }}",
					})
				}
			}
		}

		// Warn if job has no steps
		if len(job.Steps) == 0 {
			issues = append(issues, scanIssue{
				level:    "warning",
				location: loc,
				message:  "job has no steps",
			})
		}
	}

	return issues
}

func printScanResults(workflows []*workflow.Workflow, issues []scanIssue) {
	const (
		colorReset  = "\033[0m"
		colorBold   = "\033[1m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorRed    = "\033[31m"
		colorGray   = "\033[90m"
	)

	errorCount := 0
	warnCount := 0
	for _, issue := range issues {
		if issue.level == "error" {
			errorCount++
		} else {
			warnCount++
		}
	}

	fmt.Printf("\n%sScan Results%s  %s(%d workflow(s))%s\n",
		colorBold, colorReset, colorGray, len(workflows), colorReset)
	fmt.Println(strings.Repeat("─", 72))

	if len(issues) == 0 {
		fmt.Printf("  %s✓ No issues found%s\n", colorGreen, colorReset)
	} else {
		for _, issue := range issues {
			var icon, color string
			if issue.level == "error" {
				icon = "✖"
				color = colorRed
			} else {
				icon = "⚠"
				color = colorYellow
			}
			fmt.Printf("  %s%s %s%s\n", color, icon, issue.level, colorReset)
			fmt.Printf("    %slocation:%s %s\n", colorGray, colorReset, issue.location)
			fmt.Printf("    %s\n\n", issue.message)
		}
	}

	fmt.Println(strings.Repeat("─", 72))

	summary := fmt.Sprintf("  %d error(s), %d warning(s)", errorCount, warnCount)
	if errorCount > 0 {
		fmt.Printf("%s%s%s\n\n", colorRed, summary, colorReset)
	} else if warnCount > 0 {
		fmt.Printf("%s%s%s\n\n", colorYellow, summary, colorReset)
	} else {
		fmt.Printf("%s%s%s\n\n", colorGreen, summary, colorReset)
	}
}
