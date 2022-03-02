package transform

import (
	"log"
	"sync"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type FieldProcessStep struct {
	Field     string            `json:"field"`
	Mapping   map[string]string `json:"mapping"`
	ItemField string            `json:"itemField" jsonschema_description:"If processing an array of non-dict elements, create a dict as {itemField:element}"`
}

func (fs *FieldProcessStep) Init(task task.RuntimeTask) {
}

func (fs FieldProcessStep) Start(in chan map[string]interface{}, task task.RuntimeTask, wg *sync.WaitGroup) (chan map[string]interface{}, error) {
	out := make(chan map[string]interface{}, 10)

	go func() {
		defer close(out)
		for i := range in {
			if v, err := evaluate.GetJSONPath(fs.Field, i); err == nil {
				if vList, ok := v.([]interface{}); ok {
					for _, l := range vList {
						m := map[string]interface{}{}
						if x, ok := l.(map[string]interface{}); ok {
							m = x
						} else {
							m[fs.ItemField] = l
						}
						r := map[string]interface{}{}
						for k, v := range m {
							r[k] = v
						}
						for k, v := range fs.Mapping {
							val, _ := evaluate.ExpressionString(v, task.GetInputs(), i)
							r[k] = val
						}
						out <- r
					}
				} else if vList, ok := v.([]string); ok {
					for _, l := range vList {
						m := map[string]interface{}{}
						m[fs.ItemField] = l
						r := map[string]interface{}{}
						for k, v := range m {
							r[k] = v
						}
						for k, v := range fs.Mapping {
							val, _ := evaluate.ExpressionString(v, task.GetInputs(), i)
							r[k] = val
						}
						out <- r
					}
				} else {
					log.Printf("Field list incorrect type: %s", v)
				}
			} else {
				//log.Printf("Field %s missing", fs.Field)
			}
		}
	}()

	return out, nil
}

func (fs FieldProcessStep) Close() {
}
