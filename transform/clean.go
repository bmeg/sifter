package transform

import (
	"github.com/bmeg/sifter/task"
)

type CleanStep struct {
	Fields      []string `json:"fields" jsonschema_description:"List of valid fields that will be left. All others will be removed"`
	RemoveEmpty bool     `json:"removeEmpty"`
	StoreExtra  string   `json:"storeExtra"`
}

func (cs *CleanStep) has(name string) bool {
	for _, s := range cs.Fields {
		if s == name {
			return true
		}
	}
	return false
}

func (cs *CleanStep) Init(task task.RuntimeTask) (Processor, error) {
	return cs, nil
}

func (fs *CleanStep) Close() {}

func (fs *CleanStep) Process(i map[string]interface{}) []map[string]interface{} {
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
	return []map[string]any{o}
}
