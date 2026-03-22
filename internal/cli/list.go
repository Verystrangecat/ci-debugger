package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/murataslan1/ci-debugger/internal/workflow"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available workflows, jobs, and steps",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			workflows, err := workflow.DiscoverWorkflows(cwd)
			if err != nil {
				return err
			}

			for _, wf := range workflows {
				name := wf.Name
				if name == "" {
					name = wf.FileName
				}
				fmt.Printf("\033[1m%s\033[0m  \033[90m(%s)\033[0m\n", name, wf.FileName)

				// Resolve job order for display
				layers, err := workflow.ResolveJobOrder(wf.Jobs)
				if err != nil {
					fmt.Printf("  \033[31mWarning: %v\033[0m\n", err)
					// Fall back to unordered
					for id, job := range wf.Jobs {
						printJob(id, job)
					}
					continue
				}

				for _, layer := range layers {
					for _, jobID := range layer {
						job := wf.Jobs[jobID]
						printJob(jobID, job)
					}
				}
				fmt.Println()
			}
			return nil
		},
	}
}

func printJob(id string, job *workflow.Job) {
	name := job.DisplayName(id)
	runsOn := ""
	if len(job.RunsOn) > 0 {
		runsOn = job.RunsOn[0]
	}
	fmt.Printf("  \033[35m▶ %s\033[0m", name)
	if id != name {
		fmt.Printf("  \033[90m(id: %s)\033[0m", id)
	}
	if runsOn != "" {
		fmt.Printf("  \033[90m[%s]\033[0m", runsOn)
	}
	fmt.Println()

	for i, step := range job.Steps {
		stepName := step.DisplayName(i)
		if step.Uses != "" {
			fmt.Printf("      \033[90m%d. uses: %s\033[0m\n", i+1, stepName)
		} else {
			fmt.Printf("      %d. %s\n", i+1, stepName)
		}
	}
}
