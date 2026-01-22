package playbook

import (
	"path/filepath"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
)

func (pout *Output) GetOutputs(task task.RuntimeTask) []string {
	if pout.JSON != nil {
		return pout.JSON.GetOutputs(task)
	} else if pout.Graph != nil {
		return pout.Graph.GetOutputs(task)
	} else if pout.Table != nil {
		return pout.Table.GetOutputs(task)
	}

	return []string{}
}

type OutputProcessor interface {
	Process(i map[string]any)
	//GetOutputs(task task.RuntimeTask) []string
	Close()
}

type OutputJSON struct {
	Path string `json:"path"`
}

func (oj *OutputJSON) GetOutputs(task task.RuntimeTask) []string {
	output, err := evaluate.ExpressionString(oj.Path, task.GetConfig(), nil)
	if err != nil {
		return []string{}
	}
	outputPath := filepath.Join(task.OutDir(), output)
	logger.Debug("table output %s %s", task.OutDir(), output)
	return []string{outputPath}
}

func (ts OutputJSON) Init(task task.RuntimeTask) (OutputProcessor, error) {
	return &jsonOutputProcess{config: ts, task: task}, nil
}

type jsonOutputProcess struct {
	config OutputJSON
	task   task.RuntimeTask
	count  uint64
}

func (op *jsonOutputProcess) Close() {
	logger.Info("Emit Summary", "name", op.config.Path, "count", op.count)
}

func (op *jsonOutputProcess) Process(i map[string]interface{}) {
	op.count++
	op.task.Emit(op.config.Path, i)
}
