package run

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/bmeg/sifter/loader"
	"github.com/bmeg/sifter/playbook"

	"github.com/spf13/cobra"
)

var workDir string = "./"
var outDir string = "./out"
var resume string = ""
var graph string = ""
var inputFile string = ""
var toStdout bool
var keep bool
var cmdInputs map[string]string

var proxy = ""
var port = 8888

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run importer",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var lps *LoadProxyServer
		if proxy != "" {
			lps = NewLoadProxyServer(port, proxy)
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

		ld, err := loader.NewLoader(driver)
		if err != nil {
			log.Printf("Error stating load manager: %s", err)
			return err
		}
		if lps != nil {
			ld = loader.NewLoadCounter(ld, 1000, func(i uint64) { lps.UpdateCount(i) })
		}
		defer ld.Close()

		inputs := map[string]interface{}{}
		if inputFile != "" {
			if err := playbook.ParseDataFile(inputFile, &inputs); err != nil {
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
		for _, playFile := range args {
			Execute(playFile, dir, outDir, inputs)
		}
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
	flags.IntVar(&port, "port", port, "Proxy Port")
	flags.StringToStringVarP(&cmdInputs, "inputs", "i", cmdInputs, "Input variables")
	flags.StringVarP(&inputFile, "inputfile", "f", inputFile, "Input variables file")
}
