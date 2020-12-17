package transform

import (
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
	task.Runtime.EmitObject(ts.Name, ts.Class, i)
	return i
}

func (ts EmitStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	task.Runtime.Emit(ts.Name, i)
	return i
}
