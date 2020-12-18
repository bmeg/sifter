package transform

import (
	"sync"

	"github.com/bmeg/sifter/pipeline"
)

type CleanStep struct {
	Fields []string `json:"fields" jsonschema_description:"List of valid fields that will be left. All others will be removed"`
}

func (fs CleanStep) has(name string) bool {
	for _, s := range fs.Fields {
		if s == name {
			return true
		}
	}
	return false
}

func (fs CleanStep) Start(in chan map[string]interface{}, task *pipeline.Task, wg *sync.WaitGroup) (chan map[string]interface{}, error) {
	out := make(chan map[string]interface{}, 10)

	go func() {
		defer close(out)
		for i := range in {
			o := map[string]interface{}{}
			for k, v := range i {
				if fs.has(k) {
					o[k] = v
				}
			}
			out <- o
		}
	}()

	return out, nil
}
