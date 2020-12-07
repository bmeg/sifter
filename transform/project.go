package transform

import (
  "github.com/bmeg/sifter/evaluate"
  "github.com/bmeg/sifter/pipeline"
)

type ProjectStep struct {
	Mapping map[string]interface{} `json:"mapping" jsonschema_description:"New fields to be generated from template"`
  Rename  map[string]string      `json:"rename" jsonschema_description:"Rename field (no template engine)"`
}


func valueRender(v interface{}, task *pipeline.Task, row map[string]interface{}) (interface{}, error) {
	if vStr, ok := v.(string); ok {
		return evaluate.ExpressionString(vStr, task.Inputs, row)
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
			j, _ := evaluate.ExpressionString(vStr, task.Inputs, row)
			o = append(o, j)
		}
		return o, nil
	}
	return v, nil
}

func (pr ProjectStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {

	o := map[string]interface{}{}
	for k, v := range i {
    if r, ok := pr.Rename[k]; ok {
      o[r] = v
    } else {
		  o[k] = v
    }
	}

	for k, v := range pr.Mapping {
		o[k], _ = valueRender(v, task, i)
	}
	return o
}
