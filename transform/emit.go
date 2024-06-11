package transform

import (
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
)

type EmitStep struct {
	Name    string `json:"name"`
	UseName bool   `json:"UseName"`
}

type emitProcess struct {
	config EmitStep
	task   task.RuntimeTask
	count  uint64
}

func (ts EmitStep) Init(t task.RuntimeTask) (Processor, error) {
	return &emitProcess{ts, t, 0}, nil
}

func (ts *emitProcess) Close() {
	logger.Info("Emit Summary", "name", ts.config.Name, "count", ts.count)
}

func (ts *emitProcess) Process(i map[string]interface{}) []map[string]interface{} {
	name, err := evaluate.ExpressionString(ts.config.Name, ts.task.GetConfig(), i)
	if err == nil {
		ts.count++
		ts.task.Emit(name, i, ts.config.UseName)
	}
	return []map[string]any{i}
}
