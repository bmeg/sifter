package inspect

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
	"github.com/spf13/cobra"
)

var outDir string = ""
var cmdInputs map[string]string

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect a sifter script file to view i/o config setup",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputs := map[string]string{}

		for k, v := range cmdInputs {
			inputs[k] = v
		}

		playFile := args[0]

		pb := playbook.Playbook{}
		if err := playbook.ParseFile(playFile, &pb); err != nil {
			logger.Info("%s", err)
			return err
		}
		var err error
		inputs, err = pb.PrepConfig(inputs, "./")
		logger.Info("inputs: %s", inputs)
		if err != nil {
			return err
		}

		if outDir == "" {
			outDir = pb.GetDefaultOutDir()
		}

		logger.Info("outdir: %s", outDir)

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
