package scan

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
	"github.com/spf13/cobra"
)

var jsonOut = false
var objectsOnly = false
var baseDir = ""

type Entry struct {
	ObjectType string `json:"objectType"`
	SifterFile string `json:"sifterFile"`
	Outfile    string `json:"outFile"`
}

var ObjectCommand = &cobra.Command{
	Use:   "objects",
	Short: "Scan for outputs",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		scanDir := args[0]

		outputs := []Entry{}

		PathWalker(scanDir, func(pb *playbook.Playbook) {
			for pname, p := range pb.Pipelines {
				emitName := ""
				for _, s := range p {
					if s.Emit != nil {
						emitName = s.Emit.Name
					}
				}
				if emitName != "" {
					for _, s := range p {
						outdir := pb.GetDefaultOutDir()
						outname := fmt.Sprintf("%s.%s.%s.json.gz", pb.Name, pname, emitName)
						outpath := filepath.Join(outdir, outname)
						o := Entry{SifterFile: pb.GetPath(), Outfile: outpath}
						if s.ObjectValidate != nil {
							//outpath, _ = filepath.Rel(baseDir, outpath)
							//fmt.Printf("%s\t%s\n", s.ObjectValidate.Title, outpath)
							o.ObjectType = s.ObjectValidate.Title
						}
						if objectsOnly {
							if o.ObjectType != "" {
								outputs = append(outputs, o)
							}
						} else {
							outputs = append(outputs, o)
						}
					}
				}
			}
		})

		if jsonOut {
			j := json.NewEncoder(os.Stdout)
			j.SetIndent("", "  ")
			j.Encode(outputs)
		} else {
			for _, i := range outputs {
				fmt.Printf("%s\t%s\n", i.ObjectType, i.Outfile)
			}
		}

		return nil

	},
}

type ScriptEntry struct {
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Inputs  []string `json:"inputs"`
	Outputs []string `json:"outputs"`
}

func removeDuplicates(s []string) []string {
	t := map[string]bool{}

	for _, i := range s {
		t[i] = true
	}
	out := []string{}
	for k := range t {
		out = append(out, k)
	}
	return out
}

func relPathArray(basedir string, paths []string) []string {
	out := []string{}
	for _, i := range paths {
		if o, err := filepath.Rel(baseDir, i); err == nil {
			out = append(out, o)
		}
	}
	return out
}

var ScriptCommand = &cobra.Command{
	Use:   "scripts",
	Short: "Scan for scripts",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		scanDir := args[0]

		scripts := []ScriptEntry{}

		if baseDir == "" {
			baseDir, _ = os.Getwd()
		}
		baseDir, _ = filepath.Abs(baseDir)
		//fmt.Printf("basedir: %s\n", baseDir)

		userInputs := map[string]string{}

		PathWalker(scanDir, func(pb *playbook.Playbook) {
			path := pb.GetPath()
			scriptDir := filepath.Dir(path)

			config, _ := pb.PrepConfig(userInputs, baseDir)

			task := task.NewTask(pb.Name, scriptDir, baseDir, pb.GetDefaultOutDir(), config)
			sourcePath, _ := filepath.Abs(path)

			cmdPath, _ := filepath.Rel(baseDir, sourcePath)

			inputs := []string{}
			outputs := []string{}
			for _, p := range pb.GetConfigFields() {
				if p.IsDir() || p.IsFile() {
					inputs = append(inputs, config[p.Name])
				}
			}
			//inputs = append(inputs, sourcePath)

			sinks, _ := pb.GetOutputs(task)
			for _, v := range sinks {
				outputs = append(outputs, v...)
			}

			emitters, _ := pb.GetEmitters(task)
			for _, v := range emitters {
				outputs = append(outputs, v)
			}

			//for _, e := range pb.Inputs {
			//}

			s := ScriptEntry{
				Path:    cmdPath,
				Name:    pb.Name,
				Outputs: relPathArray(baseDir, removeDuplicates(outputs)),
				Inputs:  relPathArray(baseDir, removeDuplicates(inputs)),
			}
			scripts = append(scripts, s)
		})

		if jsonOut {
			e := json.NewEncoder(os.Stdout)
			e.SetIndent("", "  ")
			e.Encode(scripts)
		} else {
			for _, i := range scripts {
				fmt.Printf("%s\n", i)
			}
		}

		return nil
	},
}

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for scripts or objects",
}

func init() {
	Cmd.AddCommand(ObjectCommand)
	Cmd.AddCommand(ScriptCommand)

	objFlags := ObjectCommand.Flags()
	objFlags.BoolVarP(&objectsOnly, "objects", "s", objectsOnly, "Objects Only")
	objFlags.BoolVarP(&jsonOut, "json", "j", jsonOut, "Output JSON")

	scriptFlags := ScriptCommand.Flags()
	scriptFlags.StringVarP(&baseDir, "base", "b", baseDir, "Base Dir")
	scriptFlags.BoolVarP(&jsonOut, "json", "j", jsonOut, "Output JSON")

}

func PathWalker(baseDir string, userFunc func(*playbook.Playbook)) {
	filepath.Walk(baseDir,
		func(path string, info fs.FileInfo, err error) error {
			if strings.HasSuffix(path, ".yaml") {
				pb := playbook.Playbook{}
				if parseErr := playbook.ParseFile(path, &pb); parseErr == nil {
					userFunc(&pb)
				}
			}
			return nil
		})
}
