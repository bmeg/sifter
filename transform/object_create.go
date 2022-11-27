package transform

import (
	"fmt"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/schema"
	"github.com/bmeg/sifter/task"
)

type ObjectCreateStep struct {
	Class  string `json:"class" jsonschema_description:"Object class, should match declared class in JSON Schema"`
	Schema string `json:"schema" jsonschema_description:"Directory with JSON schema files"`
}

type objectProcess struct {
	config ObjectCreateStep
	task   task.RuntimeTask
	schema schema.Schema
}

func (ts ObjectCreateStep) Init(task task.RuntimeTask) (Processor, error) {
	className, err := evaluate.ExpressionString(ts.Class, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}
	path, err := evaluate.ExpressionString(ts.Schema, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	sc, err := schema.Load(path)
	if err != nil {
		return nil, err
	}
	if c, ok := sc.Classes[className]; !ok {
		return nil, fmt.Errorf("class %s not found", className)
	} else {
		return &objectProcess{ts, task, c}, nil
	}
}

func (ts ObjectCreateStep) GetConfigFields() []config.ConfigVar {
	out := []config.ConfigVar{}
	if ts.Schema != "" {
		for _, s := range evaluate.ExpressionIDs(ts.Schema) {
			out = append(out, config.ConfigVar{Type: config.Dir, Name: config.TrimPrefix(s)})
		}
	}
	return out
}

func (ts *objectProcess) Process(i map[string]interface{}) []map[string]interface{} {
	o, err := ts.schema.Validate(i)
	if err == nil {
		return []map[string]any{o}
	}
	return []map[string]any{}
}

func (ts *objectProcess) Close() {

}
