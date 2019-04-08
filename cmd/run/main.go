package run

import (
	"fmt"
	"log"

	"github.com/bmeg/sifter/manager"
	"github.com/bmeg/sifter/webserver"
	"github.com/spf13/cobra"
)

var graph string = "test-data"
var runOnce bool = false
var workDir string = "./"
var server int = 0
var dbServer string = "grip://localhost:8202"

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run importer",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		man, err := manager.Init(manager.Config{GripServer: dbServer, WorkDir: workDir})
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

		playFile := args[0]
		dataFile := args[1]

		inputs := map[string]interface{}{}
		if err := manager.ParseDataFile(dataFile, &inputs); err != nil {
			log.Printf("%s", err)
			return err
		}

		fmt.Printf("Starting: %s\n", playFile)
		pb := manager.Playbook{}
		if err := manager.ParseFile(playFile, &pb); err != nil {
			log.Printf("%s", err)
			return err
		}
		if server != 0 {
			go pb.Execute(man, graph, inputs)
			conf := webserver.WebServerHandler{
				PostPlaybookHandler:        nil,
				GetPlaybookHandler:         nil,
				GetStatusHandler:           manager.NewManagerStatusHandler(man),
				PostPlaybookIDGraphHandler: nil,
			}
			webserver.RunServer(conf, server, "")
		} else {
			pb.Execute(man, graph, inputs)
		}
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.BoolVar(&runOnce, "run-once", false, "Only Run if database is unintialized")
	flags.StringVar(&graph, "graph", graph, "Destination Graph")
	flags.StringVar(&workDir, "workdir", workDir, "Workdir")
	flags.IntVar(&server, "server", server, "ServerPort")
	flags.StringVar(&dbServer, "db", dbServer, "Destination Server")
}
