package plan

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
	"github.com/spf13/cobra"
)

var inputFile string = ""
var cmdInputs map[string]string

type Step struct {
	Command string
	Inputs  []string
	Outputs []string
}

var snakeFile string = `

{{range $key, $val := .}}
rule {{$key}}:
{{- if $val.Inputs }}
	input:
		{{range $index, $file := $val.Inputs -}}
			{{- if $index -}},
			{{- end -}}
			"{{- $file -}}"
		{{- end}}
{{- end}}
	output:
		{{range $index, $file := $val.Outputs -}}
			{{- if $index -}},
			{{- end -}}
			"{{- $file -}}"
		{{- end}}
	shell:
		"{{$val.Command}}"
{{end}}


`

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "plan",
	Short: "Scan directory to plan operations",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir, _ := filepath.Abs(args[0])

		userInputs := map[string]any{}

		steps := map[string]Step{}

		filepath.Walk(baseDir,
			func(path string, info fs.FileInfo, err error) error {
				if strings.HasSuffix(path, ".yaml") {
					fmt.Printf("%s\n", path)

					pb := playbook.Playbook{}
					if err := playbook.ParseFile(path, &pb); err == nil {

						if len(pb.Pipelines) > 0 || len(pb.Sources) > 0 || len(pb.Scripts) > 0 {

							localInputs := pb.PrepInputs(userInputs, "./")
							task := &task.Task{Name: pb.Name, Inputs: localInputs, Workdir: "./", Emitter: nil}

							taskInputs, _ := pb.GetInputs(task)

							inputs := []string{}
							outputs := []string{}
							for _, p := range taskInputs {
								inputs = append(inputs, p)
							}

							sinks, _ := pb.GetSinks(task)
							for _, v := range sinks {
								for _, p := range v {
									outputs = append(outputs, p)
								}
							}

							emitters, _ := pb.GetEmitters(task)
							for _, v := range emitters {
								outputs = append(outputs, v)
							}

							scriptInputs := pb.GetScriptInputs(task)
							for _, v := range scriptInputs {
								for _, p := range v {
									inputs = append(inputs, p)
								}
							}

							scriptOutputs := pb.GetScriptOutputs(task)
							for _, v := range scriptOutputs {
								for _, p := range v {
									outputs = append(outputs, p)
								}
							}

							steps[pb.Name] = Step{
								Command: fmt.Sprintf("sifter run %s", path),
								Inputs:  inputs,
								Outputs: outputs,
							}
						}
					}
				}
				return nil
			})

		tmpl, err := template.New("snakefile").Parse(snakeFile)
		if err != nil {
			panic(err)
		}
		err = tmpl.Execute(os.Stdout, steps)
		return err
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringToStringVarP(&cmdInputs, "inputs", "i", cmdInputs, "Input variables")
	flags.StringVarP(&inputFile, "inputfile", "f", inputFile, "Input variables file")
}
