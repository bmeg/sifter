package run

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/bmeg/sifter/manager"
	"github.com/bmeg/sifter/webserver"
	"github.com/bmeg/sifter/evaluate"

	"github.com/spf13/cobra"
)

var graph string = "test-data"
var runOnce bool = false
var workDir string = "./"
var server int = 0
var dbServer string = "grip://localhost:8202"
var cmdInputs map[string]string

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run importer",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		man, err := manager.Init(manager.Config{GripServer: dbServer, WorkDir: workDir})
		if err != nil {
			log.Printf("Error stating load manager: %s", err)
			return err
		}
		defer man.Close()

		man.AllowLocalFiles = true

		if runOnce {
			if man.GraphExists(graph) {
				log.Printf("Graph found, exiting")
				return nil
			}
		}

		inputs := map[string]interface{}{}

		playFile := args[0]
		pb := manager.Playbook{}
		if err := manager.ParseFile(playFile, &pb); err != nil {
			log.Printf("%s", err)
			return err
		}

		if len(args) > 1 {
			dataFile := args[1]
			fileInputs := map[string]interface{}{}
			if err := manager.ParseDataFile(dataFile, &fileInputs); err != nil {
				log.Printf("%s", err)
				return err
			}
			for k, v := range fileInputs {
				if i, ok := pb.Inputs[k]; ok {
					if i.Type == "File" {
						inputs[k], _ = filepath.Abs(v.(string))
					} else {
						inputs[k] = v
					}
				}
			}
		}

		for k, v := range cmdInputs {
			if i, ok := pb.Inputs[k]; ok {
				if i.Type == "File" {
					inputs[k], _ = filepath.Abs(v)
				} else {
					inputs[k] = v
				}
			}
		}

		for k, v := range pb.Inputs {
			if _, ok := inputs[k]; !ok {
				if v.Default != "" {
					defaultPath := filepath.Join(filepath.Dir(playFile), v.Default)
					inputs[k], _ = filepath.Abs(defaultPath)
				}
			}
		}

		if pb.Schema != "" {
			schema, _ := evaluate.ExpressionString(pb.Schema, inputs, nil)
			pb.Schema = schema
			log.Printf("Schema: %s", schema)
		}

		fmt.Printf("Starting: %s\n", playFile)

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
	flags.StringToStringVarP(&cmdInputs, "inputs", "i", cmdInputs, "Input variables")
}
