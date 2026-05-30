package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:          "portmap",
	Short:        "A local port registry for developers",
	SilenceUsage: true,
}

func Execute() {
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	root.AddCommand(versionCmd)
}
