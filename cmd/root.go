package cmd

import (
	"os"

	"github.com/bmeg/sifter/cmd/manifest"
	"github.com/bmeg/sifter/cmd/run"
	"github.com/bmeg/sifter/cmd/graph"
	"github.com/spf13/cobra"
)

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use:           "sifter",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	RootCmd.AddCommand(run.Cmd)
	RootCmd.AddCommand(manifest.Cmd)
	RootCmd.AddCommand(graph.Cmd)
}

var genBashCompletionCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash completions file",
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.GenBashCompletion(os.Stdout)
	},
}
