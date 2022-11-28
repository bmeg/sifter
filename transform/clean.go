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

func (cs *CleanStep) Close() {}

func (cs *CleanStep) Process(i map[string]interface{}) []map[string]interface{} {
	o := map[string]interface{}{}
	if len(cs.Fields) > 0 {
		extra := map[string]interface{}{}
		for k, v := range i {
			if cs.has(k) {
				o[k] = v
			} else if cs.StoreExtra != "" {
				extra[k] = v
			}
		}
		if cs.StoreExtra != "" {
			o[cs.StoreExtra] = extra
		}
	} else if cs.RemoveEmpty {
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
