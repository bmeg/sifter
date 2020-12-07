package transform

import (
	"github.com/bmeg/sifter/pipeline"
)

func (ts ObjectCreateStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	task.Runtime.EmitObject(ts.Name, ts.Class, i)
	return i
}

func (ts EmitStep) Run(i map[string]interface{}, task *pipeline.Task) map[string]interface{} {
	task.Runtime.Emit(ts.Name, i)
	return i
}
