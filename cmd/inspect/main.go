package inspect

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
	"github.com/spf13/cobra"
)

var outDir string = ""
var inputFile string = ""
var cmdInputs map[string]string

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect script",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputs := map[string]string{}

		for k, v := range cmdInputs {
			inputs[k] = v
		}

		playFile := args[0]

		pb := playbook.Playbook{}
		if err := playbook.ParseFile(playFile, &pb); err != nil {
			log.Printf("%s", err)
			return err
		}
		var err error
		inputs, err = pb.PrepConfig(inputs, "./")
		if err != nil {
			return err
		}

		if outDir == "" {
			outDir = pb.GetDefaultOutDir()
		}

		log.Printf("outdir: %s", outDir)

		p, _ := filepath.Abs(playFile)
		baseDir := filepath.Dir(p)
		task := task.NewTask(pb.Name, baseDir, "./", outDir, inputs)

		out := map[string]any{}

		cf := map[string]string{}
		for _, f := range pb.GetConfigFields() {
			cf[f.Name] = f.Name //f.Type
		}
		out["configFields"] = cf

		ins := pb.GetConfigFields()
		out["config"] = ins

		outputs := map[string]any{}

		sinks, _ := pb.GetOutputs(task)
		for k, v := range sinks {
			outputs[k] = v
		}

		emitters, _ := pb.GetEmitters(task)
		for k, v := range emitters {
			outputs[k] = v
		}

		out["outputs"] = outputs

		jsonOut, _ := json.MarshalIndent(out, "", "    ")
		fmt.Printf("%s\n", string(jsonOut))

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringToStringVarP(&cmdInputs, "inputs", "i", cmdInputs, "Input variables")
}
