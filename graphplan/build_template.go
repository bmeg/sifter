package graphplan

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
)

type ObjectConvertStep struct {
	Name   string
	Input  string
	Class  string
	Schema string
}

type GraphBuildStep struct {
	Name    string
	Outdir  string
	Objects []ObjectConvertStep
}

var graphScript string = `

name: {{.Name}}
class: sifter

outdir: {{.Outdir}}

config:
{{range .Objects}}
    {{.Name}}: {{.Input}}
    {{.Name}}Schema: {{.Schema}}
{{end}}

inputs:
{{range .Objects}}
    {{.Name}}:
        jsonLoad:
            input: "{{ "{{config." }}{{.Name}}{{"}}"}}"
{{end}}

pipelines:
{{range .Objects}}
    {{.Name}}-graph:
        - from: {{.Name}}
        - graphBuild:
            schema: "{{ "{{config."}}{{.Name}}Schema{{ "}}" }}"
            title: {{.Class}}
{{end}}
`

func contains(n string, c []string) bool {
	for _, c := range c {
		if n == c {
			return true
		}
	}
	return false
}

func uniqueName(name string, used []string) string {
	if !contains(name, used) {
		return name
	}
	for i := 1; ; i++ {
		f := fmt.Sprintf("%s_%d", name, i)
		if !contains(f, used) {
			return f
		}
	}
}

func NewGraphBuild(pb *playbook.Playbook, scriptOutDir, dataDir string, objectExclude []string) error {
	userInputs := map[string]string{}
	localInputs, _ := pb.PrepConfig(userInputs, filepath.Dir(pb.GetPath()))

	task := task.NewTask(pb.Name, filepath.Dir(pb.GetPath()), filepath.Dir(pb.GetPath()), pb.GetDefaultOutDir(), localInputs)

	convertName := fmt.Sprintf("%s-graph", pb.Name)

	gb := GraphBuildStep{Name: convertName, Objects: []ObjectConvertStep{}, Outdir: dataDir}

	for pname, p := range pb.Pipelines {
		emitName := ""
		for _, s := range p {
			if s.Emit != nil {
				emitName = s.Emit.Name
			}
		}
		if emitName != "" {
			for _, s := range p {
				if s.ObjectValidate != nil {
					if !contains(s.ObjectValidate.Title, objectExclude) {
						schema, _ := evaluate.ExpressionString(s.ObjectValidate.Schema, task.GetConfig(), map[string]any{})
						outdir := pb.GetDefaultOutDir()
						outname := fmt.Sprintf("%s.%s.%s.json.gz", pb.Name, pname, emitName)

						outpath := filepath.Join(outdir, outname)
						outpath, _ = filepath.Rel(scriptOutDir, outpath)

						schemaPath, _ := filepath.Rel(scriptOutDir, schema)

						_ = schemaPath

						objCreate := ObjectConvertStep{Name: pname, Input: outpath, Class: s.ObjectValidate.Title, Schema: schemaPath}
						gb.Objects = append(gb.Objects, objCreate)
					}
				}
			}
		}
	}

	if len(gb.Objects) > 0 {
		tmpl, err := template.New("graphscript").Parse(graphScript)
		if err != nil {
			panic(err)
		}

		outPath := filepath.Join(scriptOutDir, fmt.Sprintf("%s.yaml", pb.Name))
		outfile, err := os.Create(outPath)
		if err != nil {
			logger.Error("File Error", "error", err)
		}

		logger.Info("Summary", "ObjectFound", len(gb.Objects), "outPath", outPath)

		err = tmpl.Execute(outfile, gb)
		outfile.Close()
		if err != nil {
			logger.Error("Template Error", "error", err)
		}
	}
	return nil
}
