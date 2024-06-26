package transform

import (
	"fmt"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/logger"
	"github.com/bmeg/sifter/task"
)

type FilterStep struct {
	Field   string     `json:"field"`
	Value   string     `json:"value"`
	Match   string     `json:"match"`
	Check   string     `json:"check" jsonschema_description:"How to check value, 'exists' or 'hasValue'"`
	Method  string     `json:"method"`
	Python  string     `json:"python"`
	GPython *CodeBlock `json:"gpython"`
}

type filterProcessor struct {
	config FilterStep
	proc   evaluate.Processor
	task   task.RuntimeTask
}

func (fs FilterStep) Init(task task.RuntimeTask) (Processor, error) {

	if fs.Python != "" && fs.Method != "" {
		logger.Debug("Starting Filter Map: %s", fs.Python)
		e := evaluate.GetEngine("python", task.WorkDir())
		c, err := e.Compile(fs.Python, fs.Method)
		if err != nil {
			logger.Error("Compile Error: %s", err)
		}
		return &filterProcessor{fs, c, task}, nil
	} else if fs.GPython != nil && fs.Method != "" {
		logger.Debug("Starting Filter Map: %s", fs.GPython)
		fs.GPython.SetBaseDir(task.BaseDir())
		e := evaluate.GetEngine("gpython", task.WorkDir())
		c, err := e.Compile(fs.GPython.String(), fs.Method)
		if err != nil {
			logger.Error("Compile Error: %s", err)
		}
		return &filterProcessor{fs, c, task}, nil
	}
	if fs.Check != "" {
		switch fs.Check {
		case "hasValue", "exists", "not":
		default:
			return nil, fmt.Errorf("check %s not value (hasValue/exists/not)", fs.Check)
		}
	}
	return &filterProcessor{fs, nil, task}, nil
}

func (fs *filterProcessor) Process(row map[string]interface{}) []map[string]any {
	if fs.proc != nil {
		out, err := fs.proc.EvaluateBool(row)
		if err != nil {
			logger.Error("Filter Error: %s", err)
		}
		if out {
			return []map[string]any{row}
		}
		return []map[string]any{}
	}
	value := ""
	var err error
	if fs.config.Value != "" {
		value, err = evaluate.ExpressionString(fs.config.Value, fs.task.GetConfig(), row)
		if err != nil {
			logger.Error("Expression Error: %s", err)
		}
	} else if fs.config.Field != "" {
		i, e := evaluate.GetJSONPath(fs.config.Field, row)
		err = e
		if vstr, ok := i.(string); ok {
			value = vstr
		}
	}
	if (fs.config.Check == "" && fs.config.Match == "") || fs.config.Check == "exists" {
		if err == nil {
			return []map[string]any{row}
		}
		return []map[string]any{}
	} else if fs.config.Check == "hasValue" {
		if err != nil {
			logger.Error("Field Error: %s", err)
		}
		if err == nil && value != "" {
			return []map[string]any{row}
		}
		return []map[string]any{}
	} else if fs.config.Check == "not" {
		if err != nil {
			logger.Error("Field Error: %s", err)
		}
		if err == nil && value != fs.config.Match {
			return []map[string]any{row}
		}
		return []map[string]any{}
	}

	match, _ := evaluate.ExpressionString(fs.config.Match, fs.task.GetConfig(), row)
	if value == match {
		return []map[string]any{row}
	}
	return []map[string]any{}
}

func (fs *filterProcessor) PoolReady() bool {
	return true
}

func (fs *filterProcessor) Close() {
	if fs.proc != nil {
		fs.proc.Close()
	}
}
