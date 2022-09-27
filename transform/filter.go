package transform

import (
	"log"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type FilterStep struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Match   string `json:"match"`
	Check   string `json:"check" jsonschema_description:"How to check value, 'exists' or 'hasValue'"`
	Method  string `json:"method"`
	Python  string `json:"python"`
	GPython string `json:"gpython"`
	Steps   Pipe   `json:"steps"`
}

type filterProcessor struct {
	config FilterStep
	proc   evaluate.Processor
	task   task.RuntimeTask
}

func (fs FilterStep) Init(task task.RuntimeTask) (Processor, error) {

	if fs.Python != "" && fs.Method != "" {
		log.Printf("Starting Filter Map: %s", fs.Python)
		e := evaluate.GetEngine("python", task.WorkDir())
		c, err := e.Compile(fs.Python, fs.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		return &filterProcessor{fs, c, task}, nil
	} else if fs.GPython != "" && fs.Method != "" {
		log.Printf("Starting Filter Map: %s", fs.GPython)
		e := evaluate.GetEngine("gpython", task.WorkDir())
		c, err := e.Compile(fs.GPython, fs.Method)
		if err != nil {
			log.Printf("Compile Error: %s", err)
		}
		return &filterProcessor{fs, c, task}, nil
	}
	return &filterProcessor{fs, nil, task}, nil
}

func (fs *filterProcessor) Process(i map[string]interface{}) []map[string]any {
	if fs.proc != nil {
		out, err := fs.proc.EvaluateBool(i)
		if err != nil {
			log.Printf("Filter Error: %s", err)
		}
		if out {
			return []map[string]any{i}
		}
		return []map[string]any{}
	}
	value := ""
	var err error
	if fs.config.Value != "" {
		value, err = evaluate.ExpressionString(fs.config.Value, fs.task.GetConfig(), i)
	} else if fs.config.Field != "" {
		i, e := evaluate.GetJSONPath(fs.config.Field, i)
		err = e
		if vstr, ok := i.(string); ok {
			value = vstr
		}
	}
	if (fs.config.Check == "" && fs.config.Match == "") || fs.config.Check == "exists" {
		if err == nil {
			return []map[string]any{i}
		}
		return []map[string]any{}
	} else if fs.config.Check == "hasValue" {
		if err == nil && value != "" {
			return []map[string]any{i}
		}
		return []map[string]any{}
	}

	match, _ := evaluate.ExpressionString(fs.config.Match, fs.task.GetConfig(), i)
	if value == match {
		return []map[string]any{i}
	}
	return []map[string]any{}
}

func (fs *filterProcessor) Close() {
	if fs.proc != nil {
		fs.proc.Close()
	}
}
