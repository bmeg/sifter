package transform

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/schema"
	"github.com/bmeg/sifter/task"
)

type ObjectValidateStep struct {
	Class  string `json:"class" jsonschema_description:"Object class, should match declared class in JSON Schema"`
	Schema string `json:"schema" jsonschema_description:"Directory with JSON schema files"`
}

type objectProcess struct {
	config     ObjectValidateStep
	task       task.RuntimeTask
	className  string
	schema     schema.GraphSchema
	errorCount int
}

func (ts ObjectValidateStep) Init(task task.RuntimeTask) (Processor, error) {
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
	if _, ok := sc.Classes[className]; ok {
		return &objectProcess{ts, task, className, sc, 0}, nil
	}
	return nil, fmt.Errorf("class %s not found", className)
}

func (ts ObjectValidateStep) GetConfigFields() []config.Variable {
	out := []config.Variable{}
	if ts.Schema != "" {
		for _, s := range evaluate.ExpressionIDs(ts.Schema) {
			out = append(out, config.Variable{Type: config.Dir, Name: config.TrimPrefix(s)})
		}
	}
	return out
}

func (ts *objectProcess) Process(i map[string]interface{}) []map[string]interface{} {
	out, err := ts.schema.CleanAndValidate(ts.className, i)
	if err == nil {
		return []map[string]any{out}
	} else {
		//if ts.errorCount < 10 {
		data, _ := json.Marshal(i)
		log.Printf("validate %s error: %s on %s", ts.className, err, data)
		//}
		ts.errorCount++
	}
	return []map[string]any{}
}

func (ts *objectProcess) Close() {
	log.Printf("Total incorrect rows: %d", ts.errorCount)
}
