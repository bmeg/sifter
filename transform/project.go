package transform

import (
	"strings"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type ProjectStep struct {
	Mapping map[string]interface{} `json:"mapping" jsonschema_description:"New fields to be generated from template"`
	Rename  map[string]string      `json:"rename" jsonschema_description:"Rename field (no template engine)"`
}

type projectStepProcess struct {
	project ProjectStep
	task    task.RuntimeTask
}

func (pr ProjectStep) Init(t task.RuntimeTask) (Processor, error) {
	return &projectStepProcess{pr, t}, nil
}

func valueRender(v interface{}, task task.RuntimeTask, row map[string]interface{}) (interface{}, error) {
	if vStr, ok := v.(string); ok {
		return evaluate.ExpressionString(vStr, task.GetConfig(), row)
	} else if vMap, ok := v.(map[string]interface{}); ok {
		o := map[string]interface{}{}
		for key, val := range vMap {
			o[key], _ = valueRender(val, task, row)
		}
		return o, nil
	} else if vArray, ok := v.([]interface{}); ok {
		o := []interface{}{}
		for _, val := range vArray {
			j, _ := valueRender(val, task, row)
			o = append(o, j)
		}
		return o, nil
	} else if vArray, ok := v.([]string); ok {
		o := []string{}
		for _, vStr := range vArray {
			j, _ := evaluate.ExpressionString(vStr, task.GetConfig(), row)
			o = append(o, j)
		}
		return o, nil
	}
	return v, nil
}

func setProjectValue(i map[string]interface{}, key string, val interface{}) error {
	if strings.HasPrefix(key, "$.") {
		return evaluate.SetJSONPath(key, i, val)
	}
	i[key] = val
	return nil
}

func (pr *projectStepProcess) Process(i map[string]interface{}) []map[string]interface{} {

	o := map[string]interface{}{}
	for k, v := range i {
		if r, ok := pr.project.Rename[k]; ok {
			o[r] = v
		} else {
			o[k] = v
		}
	}
	for k, v := range pr.project.Mapping {
		t, _ := valueRender(v, pr.task, i)
		setProjectValue(o, k, t)
	}
	return []map[string]any{o}
}

func (pr *projectStepProcess) Close() {

}
