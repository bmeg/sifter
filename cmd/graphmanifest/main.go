package graphmanifest

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var workerCount = 1

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "graph-manifest <directory>",
	Short: "Build manifest for graph file archive",
	Long: `Build manifest for graph file archive.
Only works on .Vertex.json.gz and .Edge.json.gz files`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for info := range ScanDir(args[0], workerCount) {
			o, _ := json.Marshal(info)
			fmt.Printf("%s\n", o)
		}
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.IntVarP(&workerCount, "workers", "n", workerCount, "Worker Count")
	//flags.BoolVar(&runOnce, "run-once", false, "Only Run if database is unintialized")
	//flags.StringVar(&graph, "graph", graph, "Destination Graph")
	//flags.StringVar(&workDir, "workdir", workDir, "Workdir")
	//flags.StringVar(&gripServer, "server", gripServer, "Destination Server")
}
