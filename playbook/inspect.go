package playbook

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/task"
)

func (pb *Playbook) GetConfig(task task.RuntimeTask) (map[string]string, error) {
	out := map[string]string{}
	inputs := task.GetConfig()

	for k, v := range pb.Config {
		if v.IsFile() || v.IsDir() {
			if iv, ok := inputs[k]; ok {
				if ivStr, ok := iv.(string); ok {
					out[k], _ = task.AbsPath(ivStr)
				}
			} else {
				out[k] = ""
			}
		}
	}
	return out, nil
}

func (pb *Playbook) GetConfigFields() []config.ConfigVar {
	out := []config.ConfigVar{}

	for _, v := range pb.Inputs {
		out = append(out, v.GetConfigFields()...)
	}

	for _, v := range pb.Pipelines {
		for _, s := range v {
			out = append(out, s.GetConfigFields()...)
		}
	}

	return out
}

func (pb *Playbook) GetOutputs(task task.RuntimeTask) (map[string][]string, error) {
	out := map[string][]string{}
	//inputs := task.GetInputs()

	for k, v := range pb.Outputs {
		out[k] = v.GetOutputs(task)
	}

	return out, nil
}

func (pb *Playbook) GetEmitters(task task.RuntimeTask) (map[string]string, error) {
	out := map[string]string{}

	for k, v := range pb.Pipelines {
		for _, s := range v {
			for _, e := range s.GetEmitters() {
				fileName := fmt.Sprintf("%s.%s.%s.json.gz", pb.Name, k, e)
				filePath := filepath.Join(pb.GetOutDir(task), fileName)
				out[k+"."+e] = filePath
			}
		}
	}
	return out, nil
}

func (pb *Playbook) GetDefaultOutDir() string {
	if pb.Outdir == "" {
		out, _ := filepath.Abs("./")
		return out
	}
	path := filepath.Join(filepath.Dir(pb.path), pb.Outdir)
	out, _ := filepath.Abs(path)
	log.Printf("default: %s %s %s", pb.path, pb.Outdir, out)
	return out
}

func (pb *Playbook) GetOutDir(task task.RuntimeTask) string {
	if pb.Outdir == "" {
		return task.OutDir()
	}
	path := filepath.Join(filepath.Dir(pb.path), pb.Outdir)
	out, _ := filepath.Abs(path)
	return out
}

func (pb *Playbook) GetScriptInputs(task task.RuntimeTask) map[string][]string {
	out := map[string][]string{}
	for k, v := range pb.Scripts {
		o := []string{}
		for _, p := range v.Inputs {
			path := filepath.Join(filepath.Dir(pb.path), p)
			npath, _ := filepath.Abs(path)
			o = append(o, npath)
		}
		out[k] = o
	}
	return out
}

func (pb *Playbook) GetScriptOutputs(task task.RuntimeTask) map[string][]string {
	out := map[string][]string{}
	for k, v := range pb.Scripts {
		o := []string{}
		for _, p := range v.Outputs {
			path := filepath.Join(filepath.Dir(pb.path), p)
			npath, _ := filepath.Abs(path)
			o = append(o, npath)
		}
		out[k] = o
	}
	return out
}
