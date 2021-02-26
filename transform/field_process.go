package transform

import (
	"log"
	"sync"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/manager"
)

type FieldProcessStep struct {
	Field   string            `json:"field"`
	Steps   Pipe              `json:"steps"`
	Mapping map[string]string `json:"mapping"`
}

func (fs *FieldProcessStep) Init(task manager.RuntimeTask) {
	fs.Steps.Init(task)
}

func (fs FieldProcessStep) Start(in chan map[string]interface{}, task manager.RuntimeTask, wg *sync.WaitGroup) (chan map[string]interface{}, error) {
	inChan := make(chan map[string]interface{}, 100)
	tout, _ := fs.Steps.Start(inChan, task.Child("fieldProcess"), wg)
	out := make(chan map[string]interface{}, 10)

	go func() {
		defer close(inChan)
		defer close(out)
		for i := range in {
			out <- i
			if v, err := evaluate.GetJSONPath(fs.Field, i); err == nil {
				if vList, ok := v.([]interface{}); ok {
					for _, l := range vList {
						if m, ok := l.(map[string]interface{}); ok {
							r := map[string]interface{}{}
							for k, v := range m {
								r[k] = v
							}
							for k, v := range fs.Mapping {
								val, _ := evaluate.ExpressionString(v, task.GetInputs(), i)
								r[k] = val
							}
							inChan <- r
						} else {
							log.Printf("Incorrect Field Type: %s", l)
						}
					}
				} else {
					log.Printf("Field list incorrect type: %s", v)
				}
			} else {
				log.Printf("Field %s missing", fs.Field)
			}
		}
	}()

	//consume output of child pipeline
	go func() {
		for range tout {
		}
	}()

	return out, nil
}

func (fs FieldProcessStep) Close() {
	fs.Steps.Close()
}
