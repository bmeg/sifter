package transform

import (
	"strings"

	"github.com/bmeg/sifter/task"
)

type FieldParseStep struct {
	Field  string `json:"field"`
	Sep    string `json:"sep"`
	Assign string `json:"assign"`
}

func (fp *FieldParseStep) Init(t task.RuntimeTask) (Processor, error) {
	return fp, nil
}

func (fp *FieldParseStep) Close() {}

func (fs *FieldParseStep) PoolReady() bool {
	return true
}
func (fp *FieldParseStep) Process(i map[string]interface{}) []map[string]interface{} {

	sep := fp.Sep
	if sep == "" {
		sep = ";"
	}
	assign := fp.Assign
	if assign == "" {
		assign = "="
	}

	o := map[string]interface{}{}
	for x, y := range i {
		o[x] = y
	}
	if v, ok := i[fp.Field]; ok {
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
			o[fp.Field] = t
		}
	}
	return []map[string]any{o}
}
