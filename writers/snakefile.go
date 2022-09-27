package writers

import (
	"fmt"
	"html/template"
	"os"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type CommandLineTemplate struct {
	Template string   `json:"template"`
	Outputs  []string `json:"outputs"`
	Inputs   []string `json:"inputs"`
}

type SnakeFileWriter struct {
	FromName string                `json:"from"`
	Commands []CommandLineTemplate `json:"commands"`
}

type snakefileProcess struct {
	cmdConfig *SnakeFileWriter
	config    map[string]any
	commands  []step
	count     int
}

func (cw *SnakeFileWriter) Init(task task.RuntimeTask) (WriteProcess, error) {
	return &snakefileProcess{cw, task.GetConfig(), []step{}, 0}, nil
}

func (cw *SnakeFileWriter) From() string {
	return cw.FromName
}

func (cl *SnakeFileWriter) GetOutputs(task task.RuntimeTask) []string {
	return []string{}
}

func (cp *snakefileProcess) Close() {

	//find all final outputs
	outputs := map[string]int{}
	for _, s := range cp.commands {
		for _, f := range s.Outputs {
			outputs[f] = 0
		}
	}

	for _, s := range cp.commands {
		for _, f := range s.Inputs {
			if x, ok := outputs[f]; ok {
				outputs[f] = x + 1
			}
		}
	}

	allStep := step{
		Name:   "all",
		Inputs: []string{},
	}
	for k, v := range outputs {
		if v == 0 {
			allStep.Inputs = append(allStep.Inputs, k)
		}
	}
	steps := append([]step{allStep}, cp.commands...)

	tmpl, err := template.New("snakefile").Parse(snakeFile)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, steps)
}

func (tp *snakefileProcess) Write(i map[string]interface{}) {

	for c, t := range tp.cmdConfig.Commands {
		cmdLine, err := evaluate.ExpressionString(t.Template, tp.config, i)
		if err != nil {
			continue
		}

		inputs := []string{}
		outputs := []string{}

		for _, ti := range t.Inputs {
			n, err := evaluate.ExpressionString(ti, tp.config, i)
			if err == nil {
				inputs = append(inputs, n)
			}
		}

		for _, to := range t.Outputs {
			n, err := evaluate.ExpressionString(to, tp.config, i)
			if err == nil {
				outputs = append(outputs, n)
			}
		}

		s := step{
			Name:    fmt.Sprintf("cmd_%d_%d", tp.count, c),
			Command: cmdLine,
			Inputs:  inputs,
			Outputs: outputs,
		}
		tp.commands = append(tp.commands, s)
	}
	tp.count += 1
}

type step struct {
	Name    string
	Command string
	Inputs  []string
	Outputs []string
}

var snakeFile string = `

{{range .}}
rule {{.Name}}:
{{- if .Inputs }}
	input:
		{{range $index, $file := .Inputs -}}
			{{- if $index -}},
			{{- end -}}
			"{{- $file -}}"
		{{- end}}
{{- end}}
{{- if .Outputs }}
	output:
		{{range $index, $file := .Outputs -}}
			{{- if $index -}},
			{{- end -}}
			"{{- $file -}}"
		{{- end}}
{{- end}}
{{- if .Command }}
	shell:
		"{{.Command}}"
{{- end}}
{{end}}


`
