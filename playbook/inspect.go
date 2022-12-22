package playbook

import (
	"fmt"
	"path/filepath"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/task"
)

func (pb *Playbook) GetConfigFields() []config.Variable {
	out := []config.Variable{}

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
