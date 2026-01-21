package playbook

import (
	"path/filepath"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/task"
)

func (pb *Playbook) GetRequiredParams() []config.ParamRequest {
	out := []config.ParamRequest{}

	for _, v := range pb.Inputs {
		out = append(out, v.GetRequiredParams()...)
	}

	for _, v := range pb.Pipelines {
		for _, s := range v {
			out = append(out, s.GetRequiredParams()...)
		}
	}

	return out
}

func (pb *Playbook) GetOutputs(task task.RuntimeTask) (map[string]string, error) {
	out := map[string]string{}
	//inputs := task.GetInputs()

	for k, v := range pb.Outputs {
		filePath := filepath.Join(pb.GetOutDir(task), v.Path)
		out[k] = filePath
	}
	return out, nil
}

func (pb *Playbook) GetDefaultOutDir() string {
	if pb.Outdir == "" {
		out, _ := filepath.Abs("./")
		return out
	}
	if filepath.IsAbs(pb.Outdir) {
		return pb.Outdir
	}
	path := filepath.Join(filepath.Dir(pb.path), pb.Outdir)
	out, _ := filepath.Abs(path)
	return out
}

func (pb *Playbook) GetOutDir(task task.RuntimeTask) string {
	if pb.Outdir == "" {
		return task.OutDir()
	}
	if filepath.IsAbs(pb.Outdir) {
		return pb.Outdir
	}
	path := filepath.Join(filepath.Dir(pb.path), pb.Outdir)
	out, _ := filepath.Abs(path)
	return out
}
