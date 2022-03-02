package transform

import (
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

/*
type ObjectCreateStep struct {
	Class string `json:"class" jsonschema_description:"Object class, should match declared class in JSON Schema"`
	Name  string `json:"name" jsonschema_description:"domain name of stream, to separate it from other output streams of the same output type"`
}
*/

type EmitStep struct {
	Name string `json:"name"`
}

/*
func (ts ObjectCreateStep) Run(i map[string]interface{}, task task.RuntimeTask) map[string]interface{} {
	name, err := evaluate.ExpressionString(ts.Name, task.GetInputs(), i)
	if err == nil {
		task.EmitObject(name, ts.Class, i)
	}
	return i
}
*/

func (ts EmitStep) Run(i map[string]interface{}, task task.RuntimeTask) map[string]interface{} {
	name, err := evaluate.ExpressionString(ts.Name, task.GetInputs(), i)
	if err == nil {
		task.Emit(name, i)
	}
	return i
}
