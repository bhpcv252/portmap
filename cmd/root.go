package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

var noColor bool

var root = &cobra.Command{
	Use:           "portmap",
	Short:         "Show and manage active ports",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := root.Execute(); err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		if err.Error() != "" {
			printError(err.Error(), "")
		}
		os.Exit(1)
	}
}

func init() {
	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	root.AddCommand(
		versionCmd,
		lsCmd,
		checkCmd,
		suggestCmd,
		killCmd,
	)
}
