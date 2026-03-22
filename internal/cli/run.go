package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/murataslan1/ci-debugger/internal/debugger"
	"github.com/murataslan1/ci-debugger/internal/docker"
	"github.com/murataslan1/ci-debugger/internal/runner"
	"github.com/murataslan1/ci-debugger/internal/types"
	"github.com/murataslan1/ci-debugger/internal/ui"
	"github.com/murataslan1/ci-debugger/internal/workflow"
)

func newRunCmd() *cobra.Command {
	var (
		workflowFile  string
		jobFilter     string
		envFile       string
		secretFile    string
		platformPairs []string
		// Debugger flags
		stepMode     bool
		breakBefore  []string
		breakAfter   []string
		breakOnError bool
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a workflow locally",
		Example: `  ci-debugger run
  ci-debugger run -W .github/workflows/ci.yml
  ci-debugger run -j test
  ci-debugger run -v`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Parse platform overrides (key=value)
			platformOverrides := map[string]string{}
			for _, pair := range platformPairs {
				parts := strings.SplitN(pair, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid --platform value %q (expected key=value)", pair)
				}
				platformOverrides[parts[0]] = parts[1]
			}

			// Find workflow
			var wf *workflow.Workflow
			if workflowFile != "" {
				var err error
				wf, err = workflow.ParseFile(workflowFile)
				if err != nil {
					return err
				}
			} else {
				cwd, _ := os.Getwd()
				workflows, err := workflow.DiscoverWorkflows(cwd)
				if err != nil {
					return err
				}
				if len(workflows) == 1 {
					wf = workflows[0]
					fmt.Printf("Using workflow: %s\n", wf.FileName)
				} else {
					// Show picker
					fmt.Println("Multiple workflows found. Select one with -W:")
					for _, w := range workflows {
						fmt.Printf("  %s  (%s)\n", w.FileName, w.Name)
					}
					return fmt.Errorf("use -W to specify a workflow file")
				}
			}

			// Connect to Docker
			dockerClient, err := docker.NewClient()
			if err != nil {
				return err
			}
			defer dockerClient.Close()

			if err := dockerClient.Ping(ctx); err != nil {
				return err
			}

			// Load env and secrets
			envVars, err := runner.LoadEnvFile(envFile)
			if err != nil {
				return err
			}
			secrets, err := runner.LoadSecretFile(secretFile)
			if err != nil {
				return err
			}

			cwd, _ := os.Getwd()
			cfg := &runner.RunConfig{
				WorkflowPath:      workflowFile,
				JobFilter:         jobFilter,
				EnvFile:           envFile,
				SecretFile:        secretFile,
				PlatformOverrides: platformOverrides,
				Verbose:           verbose,
				WorkspaceDir:      cwd,
				StepMode:          stepMode,
				BreakBefore:       breakBefore,
				BreakAfter:        breakAfter,
				BreakOnError:      breakOnError,
			}

			r := runner.NewRunner(wf, cfg, dockerClient)
			r.SetEnv(envVars)
			r.SetSecrets(secrets)

			// Attach debugger if any debug flags are set
			if stepMode || breakOnError || len(breakBefore) > 0 || len(breakAfter) > 0 {
				dbg := debugger.New(debugger.BreakpointConfig{
					StepMode:     stepMode,
					BreakBefore:  breakBefore,
					BreakAfter:   breakAfter,
					BreakOnError: breakOnError,
				})
				r.SetDebugger(dbg)
				fmt.Printf("\033[1;33m◆ Debugger enabled\033[0m")
				if stepMode {
					fmt.Printf("  (step mode)")
				}
				if breakOnError {
					fmt.Printf("  (break-on-error)")
				}
				fmt.Println()
			}

			if noColor {
				ui.NoColor()
			}

			renderer := ui.New(verbose)
			renderer.RenderWorkflowStart(wf, wf.On.String())

			start := time.Now()
			result, err := r.Run(ctx)
			if err != nil {
				renderer.RenderError(err)
				return err
			}

			// Print summary
			renderer.RenderSummary(result, time.Since(start))

			// Exit with non-zero if any job failed
			for _, jr := range result.JobResults {
				if jr.Status == types.JobStatusFailed {
					os.Exit(1)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&workflowFile, "workflow", "W", "", "Workflow file to run")
	cmd.Flags().StringVarP(&jobFilter, "job", "j", "", "Run only this job")
	cmd.Flags().StringVar(&envFile, "env-file", ".env", "Environment variables file")
	cmd.Flags().StringVar(&secretFile, "secret-file", ".secrets", "Secrets file")
	cmd.Flags().StringArrayVar(&platformPairs, "platform", nil, "Platform image override (e.g. ubuntu-latest=custom:image)")
	// Debugger flags
	cmd.Flags().BoolVar(&stepMode, "step", false, "Step-by-step mode: pause before each step")
	cmd.Flags().StringArrayVar(&breakBefore, "break-before", nil, "Break before a named step (can be specified multiple times)")
	cmd.Flags().StringArrayVar(&breakAfter, "break-after", nil, "Break after a named step")
	cmd.Flags().BoolVar(&breakOnError, "break-on-error", false, "Break after any step that fails")

	return cmd
}

