package graphmanifest

import (
	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "graph-manifest",
	Short: "Build manifest for graph file archive",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ScanDir(args[0])
		return nil
	},
}

/*
func init() {
	flags := Cmd.Flags()
	flags.BoolVar(&runOnce, "run-once", false, "Only Run if database is unintialized")
	flags.StringVar(&graph, "graph", graph, "Destination Graph")
	flags.StringVar(&workDir, "workdir", workDir, "Workdir")
	flags.StringVar(&gripServer, "server", gripServer, "Destination Server")
}
*/
