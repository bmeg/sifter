package transform

import (
	"log"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type FieldProcessStep struct {
	Field     string            `json:"field"`
	Mapping   map[string]string `json:"mapping"`
	ItemField string            `json:"itemField" jsonschema_description:"If processing an array of non-dict elements, create a dict as {itemField:element}"`
}

type fieldProcess struct {
	config FieldProcessStep
	task   task.RuntimeTask
}

func (fs FieldProcessStep) Init(task task.RuntimeTask) (Processor, error) {
	return &fieldProcess{fs, task}, nil

}

func (fs *fieldProcess) Close() {}

func (fs *fieldProcess) Process(i map[string]any) []map[string]any {
	out := []map[string]any{}
	if v, err := evaluate.GetJSONPath(fs.config.Field, i); err == nil {
		if vList, ok := v.([]interface{}); ok {
			for _, l := range vList {
				m := map[string]interface{}{}
				if x, ok := l.(map[string]interface{}); ok {
					m = x
				} else {
					m[fs.config.ItemField] = l
				}
				r := map[string]interface{}{}
				for k, v := range m {
					r[k] = v
				}
				for k, v := range fs.config.Mapping {
					val, _ := evaluate.ExpressionString(v, fs.task.GetInputs(), i)
					r[k] = val
				}
				out = append(out, r)
			}
		} else if vList, ok := v.([]string); ok {
			for _, l := range vList {
				m := map[string]interface{}{}
				m[fs.config.ItemField] = l
				r := map[string]interface{}{}
				for k, v := range m {
					r[k] = v
				}
				for k, v := range fs.config.Mapping {
					val, _ := evaluate.ExpressionString(v, fs.task.GetInputs(), i)
					r[k] = val
				}
				out = append(out, r)
			}
		} else {
			log.Printf("Field list incorrect type: %s", v)
		}
	} else {
		//log.Printf("Field %s missing", fs.Field)
	}
	return out
}
