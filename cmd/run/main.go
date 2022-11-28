package run

import (
	"log"

	"github.com/bmeg/sifter/playbook"

	"github.com/spf13/cobra"
)

var workDir string = "./"
var outDir string = ""
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
		inputs := map[string]string{}
		if inputFile != "" {
			if err := playbook.ParseStringFile(inputFile, &inputs); err != nil {
				log.Printf("%s", err)
				return err
			}
		}
		for k, v := range cmdInputs {
			inputs[k] = v
		}
		for _, playFile := range args {
			if err := Execute(playFile, "./", outDir, inputs); err != nil {
				return err
			}
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
