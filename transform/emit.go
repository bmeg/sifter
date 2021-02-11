package transform

import (
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/pipeline"
)

type ObjectCreateStep struct {
	Class string `json:"class" jsonschema_description:"Object class, should match declared class in JSON Schema"`
	Name  string `json:"name" jsonschema_description:"domain name of stream, to separate it from other output streams of the same output type"`
}

type EmitStep struct {
	Name string `json:"name"`
}

func (ts ObjectCreateStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	name, err := evaluate.ExpressionString(ts.Name, task.Inputs, i)
	if err == nil {
		task.Runtime.EmitObject(name, ts.Class, i)
	}
	return i
}

func (ts EmitStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	name, err := evaluate.ExpressionString(ts.Name, task.Inputs, i)
	if err == nil {
		task.Runtime.Emit(name, i)
	}
	return i
}
