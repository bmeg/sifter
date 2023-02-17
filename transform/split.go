package transform

import (
	"strings"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type SplitStep struct {
	Field string `json:"field"`
	Sep   string `json:"sep"`
}

func (ss SplitStep) Init(task task.RuntimeTask) (Processor, error) {
	return ss, nil
}

func (ss SplitStep) Process(i map[string]any) []map[string]any {
	if v, err := evaluate.GetJSONPath(ss.Field, i); err == nil {
		if vStr, ok := v.(string); ok {
			vArray := strings.Split(vStr, ss.Sep)
			out := make([]any, len(vArray))
			for i := range vArray {
				out[i] = vArray[i]
			}
			evaluate.SetJSONPath(ss.Field, i, out)
		}
	}
	return []map[string]any{i}
}

func (ss SplitStep) Close() {}

func (pr SplitStep) PoolReady() bool {
	return true
}
