package run

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/bmeg/sifter/datastore"
	"github.com/bmeg/sifter/manager"

	"github.com/spf13/cobra"
)

var workDir string = "./"
var outDir string = "./out"
var resume string = ""
var toStdout bool
var keep bool
var cmdInputs map[string]string

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run importer",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {


		if _, err := os.Stat(outDir); os.IsNotExist(err) {
			os.MkdirAll(outDir, 0777)
		}

		driver := fmt.Sprintf("dir://%s", outDir)
		if toStdout {
			driver = "stdout://"
		}

		//TODO: This needs to be configurable
		dsConfig := datastore.Config{URL: "mongodb://localhost:27017", Database: "sifter", Collection: "cache"}


		man, err := manager.Init(manager.Config{Driver: driver, WorkDir: workDir, DataStore: &dsConfig})
		if err != nil {
			log.Printf("Error stating load manager: %s", err)
			return err
		}
		defer man.Close()

		man.AllowLocalFiles = true

		playFile := args[0]
		pb := manager.Playbook{}
		if err := manager.ParseFile(playFile, &pb); err != nil {
			log.Printf("%s", err)
			return err
		}

		inputs := map[string]interface{}{}
		if len(args) > 1 {
			dataFile := args[1]
			if err := manager.ParseDataFile(dataFile, &inputs); err != nil {
				log.Printf("%s", err)
				return err
			}
		}
		for k, v := range cmdInputs {
			inputs[k] = v
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

		pb.Execute(man, inputs, dir)
		if !keep {
			os.RemoveAll(dir)
		}
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
