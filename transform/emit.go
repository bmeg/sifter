package transform

import (
	"log"

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

type emitProcess struct {
	config EmitStep
	task   task.RuntimeTask
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

func (ts EmitStep) Init(t task.RuntimeTask) (Processor, error) {
	return &emitProcess{ts, t}, nil
}

func (ts *emitProcess) Close() {}

func (ts *emitProcess) Process(i map[string]interface{}) []map[string]interface{} {
	name, err := evaluate.ExpressionString(ts.config.Name, ts.task.GetInputs(), i)
	if err == nil {
		log.Printf("Emitting: %s", i)
		ts.task.Emit(name, i)
	}
	return []map[string]any{i}
}
