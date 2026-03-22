package cli

import (
	"github.com/spf13/cobra"
)

var (
	verbose bool
	noColor bool
)

// NewRootCmd creates the root cobra command.
func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "ci-debugger",
		Short: "Debug GitHub Actions workflows locally — with breakpoints",
		Long: `ci-debugger runs your GitHub Actions workflows locally using Docker.

Unlike act, ci-debugger lets you set breakpoints, step through execution,
and drop into an interactive shell inside the container at any point.

No more blind YAML commits.`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show full step output")
	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	root.AddCommand(newRunCmd())
	root.AddCommand(newListCmd())

	return root
}
