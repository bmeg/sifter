package transform

import (
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type EmitStep struct {
	Name    string `json:"name"`
	UseName bool   `json:"UseName"`
}

type emitProcess struct {
	config EmitStep
	task   task.RuntimeTask
}

func (ts EmitStep) Init(t task.RuntimeTask) (Processor, error) {
	return &emitProcess{ts, t}, nil
}

func (ts *emitProcess) Close() {}

func (ts *emitProcess) Process(i map[string]interface{}) []map[string]interface{} {
	name, err := evaluate.ExpressionString(ts.config.Name, ts.task.GetConfig(), i)
	if err == nil {
		ts.task.Emit(name, i, ts.config.UseName)
	}
	return []map[string]any{i}
}
