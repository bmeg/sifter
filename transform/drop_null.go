package transform

import (
	"github.com/bmeg/sifter/task"
)

type DropNullStep struct {
}

func (ss DropNullStep) Init(task task.RuntimeTask) (Processor, error) {
	return ss, nil
}

func (ss DropNullStep) Process(i map[string]any) []map[string]any {
	out := map[string]any{}
	for k, v := range i {
		if v != nil {
			out[k] = v
		}
	}
	return []map[string]any{out}
}

func (ss DropNullStep) Close() {}
