package transform

import (
	"strings"

	"github.com/bmeg/sifter/task"
)

type FieldMapStep struct {
	Column string `json:"col"`
	Sep    string `json:"sep"`
	Assign string `json:"assign"`
}

func (fm FieldMapStep) Run(i map[string]interface{}, task task.RuntimeTask) map[string]interface{} {

	sep := fm.Sep
	if sep == "" {
		sep = ";"
	}
	assign := fm.Assign
	if assign == "" {
		assign = "="
	}

	o := map[string]interface{}{}
	for x, y := range i {
		o[x] = y
	}
	if v, ok := i[fm.Column]; ok {
		if vStr, ok := v.(string); ok {
			a := strings.Split(vStr, sep)
			t := map[string]interface{}{}
			for _, s := range a {
				kv := strings.Split(s, assign)
				if len(kv) > 1 {
					t[kv[0]] = kv[1]
				} else {
					t[kv[0]] = true
				}
			}
			o[fm.Column] = t
		}
	}
	return o
}
