package transform

import (
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
)

type DebugStep struct {
	Label  string `json:"label"`
	Format bool   `json:"format"`
}

func (ds DebugStep) Init(task task.RuntimeTask) (Processor, error) {
	return ds, nil
}

func (ds DebugStep) Process(i map[string]interface{}) []map[string]interface{} {
	logger.Info("UserData", "label", ds.Label, "data", i)
	return []map[string]any{i}
}

func (ds DebugStep) Close() {}
