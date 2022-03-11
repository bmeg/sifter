package inspect

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
	"github.com/spf13/cobra"
)

var inputFile string = ""
var cmdInputs map[string]string

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect script",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

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

		playFile := args[0]

		pb := playbook.Playbook{}
		if err := playbook.ParseFile(playFile, &pb); err != nil {
			log.Printf("%s", err)
			return err
		}

		inputs = pb.PrepInputs(inputs, "./")

		task := &task.Task{Name: pb.Name, Inputs: inputs, Workdir: "./", Emitter: nil}

		out := map[string]any{}

		ins, _ := pb.GetInputs(task)
		out["inputs"] = ins

		outputs := map[string]any{}

		sinks, _ := pb.GetSinks(task)
		for k, v := range sinks {
			outputs[k] = v
		}

		emitters, _ := pb.GetEmitters(task)
		for k, v := range emitters {
			outputs[k] = v
		}

		scriptInputs := pb.GetScriptInputs(task)
		for k, v := range scriptInputs {
			inputs[k] = v
		}

		scriptOutputs := pb.GetScriptOutputs(task)
		for k, v := range scriptOutputs {
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
	flags.StringVarP(&inputFile, "inputfile", "f", inputFile, "Input variables file")
}
