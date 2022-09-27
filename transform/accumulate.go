package transform

import (
	"github.com/bmeg/sifter/task"
)

type AccumulateStep struct {
	Field string `json:"field" jsonschema_description:"Field to use for group definition"`
	Dest  string `json:"dest"`
}

func (as *AccumulateStep) Init(task task.RuntimeTask) (Processor, error) {
	return as, nil
}

func (as *AccumulateStep) Close() {}

func (as *AccumulateStep) GetKey(i map[string]any) string {
	if x, ok := i[as.Field]; ok {
		if xStr, ok := x.(string); ok {
			return xStr
		}
	}
	return ""
}

func (as *AccumulateStep) Accumulate(key string, i []map[string]interface{}) map[string]any {
	return map[string]any{
		as.Field: key,
		as.Dest:  i,
	}
}
