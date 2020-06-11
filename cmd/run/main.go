package run

import (
	"fmt"
	"log"
	"os"
	"io/ioutil"
	"path/filepath"

	"github.com/bmeg/sifter/datastore"
	"github.com/bmeg/sifter/manager"
	//"github.com/bmeg/sifter/webserver"
	"github.com/bmeg/sifter/evaluate"

	"github.com/spf13/cobra"
)

var workDir string = "./"
var outDir  string = "./out"
var resume  string = ""
var toStdout bool
var keep bool
var cmdInputs map[string]string

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run importer",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		driver := fmt.Sprintf("dir://%s", outDir)
		if toStdout {
			driver = "stdout://"
		}

		dsConfig := datastore.Config{ URL:"mongodb://localhost:27017", Database:"sifter", Collection:"cache" }

		man, err := manager.Init(manager.Config{Driver: driver, WorkDir: workDir, DataStore:&dsConfig})
		if err != nil {
			log.Printf("Error stating load manager: %s", err)
			return err
		}
		defer man.Close()

		man.AllowLocalFiles = true

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
					if i.Type == "File" || i.Type == "Directory" {
						inputs[k], _ = filepath.Abs(v.(string))
					} else {
						inputs[k] = v
					}
				}
			}
		}

		for k, v := range cmdInputs {
			if i, ok := pb.Inputs[k]; ok {
				if i.Type == "File" || i.Type == "Directory" {
					inputs[k], _ = filepath.Abs(v)
				} else {
					inputs[k] = v
				}
			}
		}

		for k, v := range pb.Inputs {
			if _, ok := inputs[k]; !ok {
				if v.Default != "" {
					if v.Type == "File" || v.Type == "Directory" {
						defaultPath := filepath.Join(filepath.Dir(playFile), v.Default)
						inputs[k], _ = filepath.Abs(defaultPath)
					} else {
						inputs[k] = v.Default
					}
				}
			}
		}

		if pb.Schema != "" {
			schema, _ := evaluate.ExpressionString(pb.Schema, inputs, nil)
			pb.Schema = schema
			log.Printf("Schema: %s", schema)
		}

		fmt.Printf("Starting: %s\n", playFile)

		dir := resume
		if dir == "" {
			d, err := ioutil.TempDir(workDir, "sifterwork_")
			if err != nil {
				log.Fatal(err)
				return err
			}
			dir = d
		}

		/*
		if server != 0 {
			go pb.Execute(man, graph, inputs, dir)
			conf := webserver.WebServerHandler{
				PostPlaybookHandler:        nil,
				GetPlaybookHandler:         nil,
				GetStatusHandler:           manager.NewManagerStatusHandler(man),
				PostPlaybookIDGraphHandler: nil,
			}
			webserver.RunServer(conf, server, "")
		} else {
			*/
			pb.Execute(man, inputs, dir)
			if !keep {
				os.RemoveAll(dir)
			}
		//}
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVar(&workDir, "workdir", workDir, "Workdir")
	flags.BoolVar(&toStdout, "s", toStdout, "To STDOUT")
	flags.BoolVar(&keep, "k", keep, "Keep Working Directory")
	flags.StringVar(&outDir, "o", outDir, "Output Dir")
	flags.StringVar(&resume, "r", resume, "Resume Directory")

	flags.StringToStringVarP(&cmdInputs, "inputs", "i", cmdInputs, "Input variables")
}
