package graph_plan

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
	"github.com/spf13/cobra"
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

var outScriptDir = ""
var outDataDir = "./"

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "graph-plan",
	Short: "Scan directory to plan operations",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir, _ := filepath.Abs(args[0])

		if outScriptDir != "" {
			baseDir, _ = filepath.Abs(outScriptDir)
		} else if len(args) > 1 {
			return fmt.Errorf("for multiple input directories, based dir must be defined")
		}

		_ = baseDir

		outDataDir, _ = filepath.Abs(outDataDir)
		outScriptDir, _ = filepath.Abs(outScriptDir)
		//outScriptDir, _ = filepath.Rel(baseDir, outScriptDir)

		userInputs := map[string]string{}

		for _, dir := range args {
			startDir, _ := filepath.Abs(dir)
			filepath.Walk(startDir,
				func(path string, info fs.FileInfo, err error) error {
					if strings.HasSuffix(path, ".yaml") {
						log.Printf("Scanning: %s", path)
						pb := playbook.Playbook{}
						if sifterErr := playbook.ParseFile(path, &pb); sifterErr == nil {
							if len(pb.Pipelines) > 0 || len(pb.Inputs) > 0 {

								localInputs, err := pb.PrepConfig(userInputs, baseDir)
								if err == nil {
									scriptDir := filepath.Dir(path)
									task := task.NewTask(pb.Name, scriptDir, baseDir, pb.GetDefaultOutDir(), localInputs)

									curDataDir, err := filepath.Rel(outScriptDir, outDataDir)
									if err != nil {
										log.Printf("Path error: %s", err)
									}

									gb := GraphBuildStep{Name: pb.Name, Objects: []ObjectConvertStep{}, Outdir: curDataDir}

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
													schema, _ := evaluate.ExpressionString(s.ObjectValidate.Schema, task.GetConfig(), map[string]any{})
													outdir := pb.GetDefaultOutDir()
													outname := fmt.Sprintf("%s.%s.%s.json.gz", pb.Name, pname, emitName)

													outpath := filepath.Join(outdir, outname)
													outpath, _ = filepath.Rel(outScriptDir, outpath)

													schemaPath, _ := filepath.Rel(outScriptDir, schema)

													_ = schemaPath

													objCreate := ObjectConvertStep{Name: pname, Input: outpath, Class: s.ObjectValidate.Title, Schema: schemaPath}
													gb.Objects = append(gb.Objects, objCreate)

												}
											}
										}
									}

									if len(gb.Objects) > 0 {
										log.Printf("Found %d objects", len(gb.Objects))
										tmpl, err := template.New("graphscript").Parse(graphScript)
										if err != nil {
											panic(err)
										}

										outfile, err := os.Create(filepath.Join(outScriptDir, fmt.Sprintf("%s.yaml", pb.Name)))
										if err != nil {
											fmt.Printf("Error: %s\n", err)
										}
										err = tmpl.Execute(outfile, gb)
										outfile.Close()
										if err != nil {
											fmt.Printf("Error: %s\n", err)
										}
									}
								}
							}
						} else {
							//log.Printf("Error: %s", sifterErr)
						}
					}
					return nil
				})
		}
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&outScriptDir, "dir", "C", outScriptDir, "Change Directory for script base")
	flags.StringVarP(&outDataDir, "out", "o", outDataDir, "Change output Directory")
}
