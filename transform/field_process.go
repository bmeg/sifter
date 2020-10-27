package transform

import (
  "log"
  "sync"

  "github.com/bmeg/sifter/evaluate"
  "github.com/bmeg/sifter/pipeline"
)

type FieldProcessStep struct {
	Field  string             `json:"field"`
	Steps   TransformPipe     `json:"steps"`
	Mapping map[string]string `json:"mapping"`
}


func (fs *FieldProcessStep) Init(task *pipeline.Task) {
  fs.Steps.Init(task)
}

func (fs FieldProcessStep) Start(in chan map[string]interface{}, task *pipeline.Task, wg *sync.WaitGroup) (chan map[string]interface{}, error) {
	inChan := make(chan map[string]interface{}, 100)
	tout, _ := fs.Steps.Start(inChan, task.Child("fieldProcess"), wg)

  go func() {
		defer close(inChan)
		for i := range in {
    	if v, err := evaluate.GetJSONPath(fs.Field, i); err == nil {
    		if vList, ok := v.([]interface{}); ok {
    			for _, l := range vList {
    				if m, ok := l.(map[string]interface{}); ok {
    					r := map[string]interface{}{}
    					for k, v := range m {
    						r[k] = v
    					}
    					for k, v := range fs.Mapping {
    						val, _ := evaluate.ExpressionString(v, task.Inputs, i)
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
	return tout, nil
}

func (fs FieldProcessStep) Close() {
  fs.Steps.Close()
}
