package playbook

import (
	"log"

	"github.com/bmeg/sifter/task"
)

func (pb *Playbook) GetInputs(task task.RuntimeTask) (map[string]string, error) {
	out := map[string]string{}
	inputs := task.GetInputs()

	for k, v := range pb.Inputs {
		if v.IsFile() {
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

func (in *Input) IsFile() bool {
	if in.Type == "File" || in.Type == "file" {
		return true
	}
	return false
}

func (pb *Playbook) GetSinks(task task.RuntimeTask) (map[string][]string, error) {
	out := map[string][]string{}
	//inputs := task.GetInputs()

	for k, v := range pb.Sinks {
		out[k] = v.GetOutputs(task)
	}

	return out, nil
}

func (pb *Playbook) GetEmitters(task task.RuntimeTask) (map[string]string, error) {
	out := map[string]string{}

	for k, v := range pb.Pipelines {
		for _, s := range v {
			for _, e := range s.GetEmitters() {
				log.Printf("Inspecting: %s %s", k, e)
				out[k+"."+e] = e
			}
		}
	}
	return out, nil
}
