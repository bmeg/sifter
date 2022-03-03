package transform

import (
	"encoding/json"

	"github.com/bmeg/sifter/task"
)

type JSONParseStep struct {
	Field string `json:"field"`
	Dest  string `json:"dest"`
}

func (jp JSONParseStep) Run(i map[string]interface{}, task task.RuntimeTask) map[string]interface{} {
	o := map[string]interface{}{}
	for x, y := range i {
		o[x] = y
	}
	if v, ok := i[jp.Field]; ok {
		if vStr, ok := v.(string); ok {
			var v interface{}
			err := json.Unmarshal([]byte(vStr), &v)
			if err == nil {
				o[jp.Dest] = v
			}
		}
	}
	return o
}
