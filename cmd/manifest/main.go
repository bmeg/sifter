package manifest

import (
	"log"
	"io/ioutil"

	"github.com/bmeg/sifter/manager"
	"github.com/bmeg/sifter/steps"
	"github.com/spf13/cobra"
)

var graph string = "test-data"
var runOnce bool = false
var workDir string = "./"
var gripServer string = "grip://localhost:8202"

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "manifest",
	Short: "Import manifest file <manifest URL> <download base URL>",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		man, err := manager.Init(manager.Config{GripServer: gripServer, WorkDir: workDir})
		if err != nil {
			log.Printf("Error stating load manager: %s", err)
			return err
		}
		defer man.Close()

		if runOnce {
			if man.GraphExists(graph) {
				log.Printf("Graph found, exiting")
				return nil
			}
		}

		manifestURL := args[0]
		baseURL := args[1]

		dir, err := ioutil.TempDir(workDir, "sifterwork_")
		if err != nil {
			log.Printf("%s", err)
			return err
		}

		run, err := man.NewRuntime(graph, dir)
		if err != nil {
			log.Printf("Error stating load runtime: %s", err)
			return err
		}

		task := run.NewTask(map[string]interface{}{})
		_, err = task.DownloadFile(manifestURL, "input.manifest")
		if err != nil {
			log.Printf("Error downloading manifest %s : %s", manifestURL, err)
			return err
		}

		mani := steps.ManifestLoadStep{
			Input:   "input.manifest",
			BaseURL: baseURL,
		}
		mani.Run(task)
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.BoolVar(&runOnce, "run-once", false, "Only Run if database is unintialized")
	flags.StringVar(&graph, "graph", graph, "Destination Graph")
	flags.StringVar(&workDir, "workdir", workDir, "Workdir")
	flags.StringVar(&gripServer, "server", gripServer, "Destination Server")
}
