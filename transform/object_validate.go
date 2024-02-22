package transform

import (
	"encoding/json"
	"fmt"

	schema "github.com/bmeg/jsonschemagraph/util"
	"github.com/bmeg/sifter/config"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader" // setup the httploader for the jsonschema checker
)

type ObjectValidateStep struct {
	Title  string `json:"title" jsonschema_description:"Object class, should match declared class title in JSON Schema"`
	URI    string `json:"uri"`
	Schema string `json:"schema" jsonschema_description:"Directory with JSON schema files"`
}

type objectProcess struct {
	config      ObjectValidateStep
	task        task.RuntimeTask
	className   string
	schema      schema.GraphSchema
	class       *jsonschema.Schema
	errorCount  int
	objectCount uint64
}

func (ts ObjectValidateStep) Init(task task.RuntimeTask) (Processor, error) {
	path, err := evaluate.ExpressionString(ts.Schema, task.GetConfig(), nil)
	if err != nil {
		return nil, err
	}

	sc, err := schema.Load(path)
	if err != nil {
		return nil, err
	}
	if ts.Title != "" {
		if cls := sc.GetClass(ts.Title); cls != nil {
			return &objectProcess{ts, task, ts.Title, sc, cls, 0, 0}, nil
		}
		return nil, fmt.Errorf("class %s not found", ts.Title)
	}
	if ts.URI != "" {
		uri := path + "/" + ts.URI
		if cls := sc.GetClass(uri); cls != nil {
			return &objectProcess{ts, task, cls.Title, sc, cls, 0, 0}, nil
		}
		return nil, fmt.Errorf("uri %s not found", uri)
	}
	return nil, fmt.Errorf("class not configured")
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

func (ts *objectProcess) PoolReady() bool {
	return true
}

func (ts *objectProcess) Process(i map[string]interface{}) []map[string]interface{} {
	ts.objectCount++
	out, err := ts.schema.CleanAndValidate(ts.class, i)
	if err == nil {
		return []map[string]any{out}
	}
	//if ts.errorCount < 10 {
	data, _ := json.Marshal(i)
	logger.Error("validation error", "className", ts.className, "error", err)
	logger.Debug("validation data", "data", data)
	//}
	ts.errorCount++
	return []map[string]any{}
}

func (ts *objectProcess) Close() {
	logger.Info("Validation Summary", "name", ts.className, "errorCount", ts.errorCount, "validationCount", ts.objectCount)
}
