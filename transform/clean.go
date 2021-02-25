package transform

import (
	"sync"

	"github.com/bmeg/sifter/manager"
)

type CleanStep struct {
	Fields      []string `json:"fields" jsonschema_description:"List of valid fields that will be left. All others will be removed"`
	RemoveEmpty bool     `json:"removeEmpty"`
	StoreExtra  string   `json:"storeExtra"`
}

func (fs CleanStep) has(name string) bool {
	for _, s := range fs.Fields {
		if s == name {
			return true
		}
	}
	return false
}

func (fs CleanStep) Start(in chan map[string]interface{}, task *manager.Task, wg *sync.WaitGroup) (chan map[string]interface{}, error) {
	out := make(chan map[string]interface{}, 10)

	go func() {
		defer close(out)
		for i := range in {
			o := map[string]interface{}{}
			if len(fs.Fields) > 0 {
				extra := map[string]interface{}{}
				for k, v := range i {
					if fs.has(k) {
						o[k] = v
					} else if fs.StoreExtra != "" {
						extra[k] = v
					}
				}
				if fs.StoreExtra != "" {
					o[fs.StoreExtra] = extra
				}
			} else if fs.RemoveEmpty {
				for k, v := range i {
					copy := true
					if vs, ok := v.(string); ok {
						if len(vs) == 0 {
							copy = false
						}
					}
					if copy {
						o[k] = v
					}
				}
			}
			out <- o
		}
	}()

	return out, nil
}
