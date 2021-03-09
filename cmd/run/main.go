package run

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/bmeg/sifter/datastore"
	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/manager"
	"github.com/bmeg/sifter/playbook"

	"github.com/spf13/cobra"
)

var workDir string = "./"
var outDir string = "./out"
var resume string = ""
var graph string = ""
var toStdout bool
var keep bool
var cmdInputs map[string]string

var proxy = ""

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run importer",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var lps *LoadProxyServer
		if proxy != "" {
			lps = NewLoadProxyServer(proxy)
			lps.Start()
		}

		if _, err := os.Stat(outDir); os.IsNotExist(err) {
			os.MkdirAll(outDir, 0777)
		}

		driver := fmt.Sprintf("dir://%s", outDir)
		if toStdout {
			driver = "stdout://"
		}
		if graph != "" {
			driver = graph
		}

		//TODO: This needs to be configurable
		dsConfig := datastore.Config{URL: "mongodb://localhost:27017", Database: "sifter", Collection: "cache"}

		ld, err := loader.NewLoader(driver)
		if err != nil {
			log.Printf("Error stating load manager: %s", err)
			return err
		}
		defer ld.Close()

		man, err := manager.Init(manager.Config{Loader: ld, WorkDir: workDir, DataStore: &dsConfig})
		if err != nil {
			log.Printf("Error stating load manager: %s", err)
			return err
		}
		defer man.Close()

		man.AllowLocalFiles = true

		playFile := args[0]

		inputs := map[string]interface{}{}
		if len(args) > 1 {
			dataFile := args[1]
			if err := playbook.ParseDataFile(dataFile, &inputs); err != nil {
				log.Printf("%s", err)
				return err
			}
		}
		for k, v := range cmdInputs {
			inputs[k] = v
		}
		dir := resume
		if dir == "" {
			d, err := ioutil.TempDir(workDir, "sifterwork_")
			if err != nil {
				log.Fatal(err)
				return err
			}
			dir = d
		}
		Execute(playFile, dir, outDir, inputs, man)
		if !keep {
			os.RemoveAll(dir)
		}
		if lps != nil {
			lps.StartProxy()
		}
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&workDir, "workdir", "w", workDir, "Workdir")
	flags.BoolVarP(&toStdout, "stdout", "s", toStdout, "To STDOUT")
	flags.BoolVarP(&keep, "keep", "k", keep, "Keep Working Directory")
	flags.StringVarP(&outDir, "out", "o", outDir, "Output Dir")
	flags.StringVarP(&resume, "resume", "r", resume, "Resume Directory")
	flags.StringVarP(&graph, "graph", "g", graph, "Output to graph")

	flags.StringVar(&proxy, "proxy", proxy, "Proxy site")

	flags.StringToStringVarP(&cmdInputs, "inputs", "i", cmdInputs, "Input variables")
}
